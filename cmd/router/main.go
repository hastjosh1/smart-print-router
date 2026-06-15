// Command router is invoked once per print job by the Redmon port redirector.
// It reads the PDF from stdin, detects whether it's a label or a report, and
// silently prints it to the correct physical printer (composing N-up for
// labels). It is short-lived: one process per job.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/smart-print-router/composer"
	"github.com/yourorg/smart-print-router/config"
	"github.com/yourorg/smart-print-router/detector"
	"github.com/yourorg/smart-print-router/printer"
	"github.com/yourorg/smart-print-router/rawlabel"
)

func main() {
	configPath := flag.String("config", "", "path to config.json (default: next to exe)")
	inFile := flag.String("in", "", "read PDF from this file instead of stdin (for testing)")
	jobName := flag.String("job", "", "print job name (used as a routing fallback)")
	dryRun := flag.Bool("dryrun", false, "don't print; write the routed/composed PDF to -out and log the decision")
	outFile := flag.String("out", "", "with -dryrun: where to write the result PDF (default: <in>.routed.pdf)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config warning:", err)
	}

	effCfgPath := *configPath
	if effCfgPath == "" {
		effCfgPath = config.DefaultPath()
	}

	logger := setupLogger(cfg.LogFile)
	logger.Printf("---- new job (name=%q, dryrun=%v) ----", *jobName, *dryRun)
	logger.Printf("config: %s", effCfgPath)
	logger.Printf("loaded: label=%q report=%q sumatra=%q", cfg.LabelPrinter, cfg.ReportPrinter, cfg.SumatraPath)
	logger.Printf("print settings: label=%q report=%q", cfg.LabelPrint(), cfg.ReportPrint())
	if cfgErr := config.CheckFileExists(effCfgPath); cfgErr != nil {
		logger.Printf("WARNING: %v (using built-in defaults — your edits are NOT being read)", cfgErr)
	}

	outPath := *outFile
	if outPath == "" {
		outPath = strings.TrimSuffix(*inFile, ".pdf") + ".routed.pdf"
	}

	// emit is the terminal action for the SumatraPDF path: print, or (dry-run)
	// copy the result to a file.
	emit := func(printerName, pdfPath, printSettings string) error {
		if *dryRun {
			logger.Printf("[dryrun] would print to %q (settings %q); writing result to %q", printerName, printSettings, outPath)
			return copyFile(pdfPath, outPath)
		}
		return printer.SilentPrint(cfg.SumatraPath, printerName, pdfPath, printSettings)
	}

	if err := run(cfg, *inFile, *jobName, *dryRun, outPath, emit, logger); err != nil {
		logger.Printf("ERROR: %v", err)
		os.Exit(1)
	}
	logger.Printf("job complete")
}

// emitFunc performs the final action on a ready-to-print PDF.
type emitFunc func(printerName, pdfPath, printSettings string) error

func run(cfg config.Config, inFile, jobName string, dryRun bool, outPath string, emit emitFunc, logger *log.Logger) error {
	pdfData, err := readInput(inFile)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	logger.Printf("received %d bytes", len(pdfData))

	tmp, err := saveTempPDF(pdfData)
	if err != nil {
		return fmt.Errorf("save temp: %w", err)
	}
	defer os.Remove(tmp)

	ps, err := detector.Detect(tmp)
	if err != nil {
		return fmt.Errorf("detect size: %w", err)
	}
	logger.Printf("page size: %.1fx%.1f mm, %d page(s)", ps.WidthMM, ps.HeightMM, ps.Pages)

	// Page size is authoritative; a job-name keyword can override it.
	isLabel := detector.IsLabel(ps, cfg.LabelDetection.MaxWidthMM, cfg.LabelDetection.MaxHeightMM)
	if hint := jobNameHint(jobName); hint != routeUnknown {
		logger.Printf("job-name hint: %s (overrides size detection)", hint)
		isLabel = hint == routeLabel
	}

	if isLabel {
		return routeToLabel(cfg, tmp, ps, dryRun, outPath, emit, logger)
	}
	logger.Printf("routing to REPORT printer %q", cfg.ReportPrinter)
	return emit(cfg.ReportPrinter, tmp, cfg.ReportPrint())
}

func routeToLabel(cfg config.Config, tmp string, ps detector.PageSize, dryRun bool, outPath string, emit emitFunc, logger *log.Logger) error {
	// Pick 2-up settings + label dimensions from the matched profile, if any.
	twoUp := cfg.DefaultLabel.TwoUp
	labelW, labelH := ps.WidthMM, ps.HeightMM
	if prof := cfg.MatchProfile(ps.WidthMM, ps.HeightMM); prof != nil {
		logger.Printf("matched profile %q", prof.Name)
		twoUp = prof.TwoUp
		labelW, labelH = prof.OutputCell()
		logger.Printf("output cell (physical sticker): %.1fx%.1f mm", labelW, labelH)
	} else {
		logger.Printf("no profile matched; using default label settings")
	}

	toPrint := tmp
	if twoUp.Enabled && twoUp.Columns > 1 {
		composed, err := composer.ComposeNUp(tmp, ps.Pages, twoUp.Columns, twoUp.GapMM, labelW, labelH)
		if err != nil {
			return fmt.Errorf("%d-up compose: %w", twoUp.Columns, err)
		}
		defer os.Remove(composed)
		toPrint = composed
		logger.Printf("composed %d-up label sheet (gap %.1f mm)", twoUp.Columns, twoUp.GapMM)
	}

	// Preferred path: native TSPL sent raw to the printer (no driver scaling/rotation).
	if cfg.LabelRaw.Enabled {
		return printLabelRaw(cfg, toPrint, dryRun, outPath, logger)
	}

	logger.Printf("routing to LABEL printer %q via SumatraPDF", cfg.LabelPrinter)
	return emit(cfg.LabelPrinter, toPrint, cfg.LabelPrint())
}

// printLabelRaw rasterizes the composed label PDF and prints it as native TSPL.
func printLabelRaw(cfg config.Config, pdfPath string, dryRun bool, outPath string, logger *log.Logger) error {
	ps, err := detector.Detect(pdfPath)
	if err != nil {
		return fmt.Errorf("measure composed label: %w", err)
	}
	dpi := cfg.LabelRaw.DPI
	if dpi <= 0 {
		dpi = 203
	}
	widthPt := ps.WidthMM / 25.4 * 72
	heightPt := ps.HeightMM / 25.4 * 72

	img, err := rawlabel.Render(cfg.LabelRaw.GSPath, pdfPath, dpi, widthPt, heightPt)
	if err != nil {
		return err
	}
	logger.Printf("rasterized label: %dx%d dots @ %d dpi (%.1fx%.1f mm)", img.Width, img.Height, dpi, ps.WidthMM, ps.HeightMM)

	data := rawlabel.BuildTSPL(img, rawlabel.Options{
		WidthMM:   ps.WidthMM,
		HeightMM:  ps.HeightMM,
		GapMM:     cfg.LabelRaw.GapMM,
		Direction: cfg.LabelRaw.Direction,
		Density:   cfg.LabelRaw.Density,
		Speed:     cfg.LabelRaw.Speed,
		Copies:    cfg.LabelRaw.Copies,
	})

	if dryRun {
		prn := strings.TrimSuffix(outPath, ".pdf") + ".prn"
		logger.Printf("[dryrun] writing %d bytes of TSPL to %q", len(data), prn)
		return os.WriteFile(prn, data, 0o644)
	}

	logger.Printf("sending %d bytes of TSPL to LABEL printer %q", len(data), cfg.LabelPrinter)
	return rawlabel.PrintRaw(cfg.LabelPrinter, data)
}

// copyFile copies src to dst (used by -dryrun to save the result).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func readInput(inFile string) ([]byte, error) {
	if inFile != "" {
		return os.ReadFile(inFile)
	}
	return io.ReadAll(os.Stdin)
}

func saveTempPDF(data []byte) (string, error) {
	f, err := os.CreateTemp("", "spr-job-*.pdf")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return "", err
	}
	return f.Name(), nil
}

type route int

const (
	routeUnknown route = iota
	routeLabel
	routeReport
)

func (r route) String() string {
	switch r {
	case routeLabel:
		return "label"
	case routeReport:
		return "report"
	default:
		return "unknown"
	}
}

func jobNameHint(name string) route {
	n := strings.ToLower(name)
	for _, kw := range []string{"barcode", "label", "sticker"} {
		if strings.Contains(n, kw) {
			return routeLabel
		}
	}
	for _, kw := range []string{"report", "invoice", "order"} {
		if strings.Contains(n, kw) {
			return routeReport
		}
	}
	return routeUnknown
}

func setupLogger(logFile string) *log.Logger {
	if logFile == "" {
		return log.New(os.Stderr, "", log.LstdFlags)
	}
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return log.New(os.Stderr, "", log.LstdFlags)
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return log.New(os.Stderr, "", log.LstdFlags)
	}
	return log.New(f, "", log.LstdFlags)
}
