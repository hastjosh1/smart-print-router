package printer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SilentPrint sends a PDF to a named Windows printer with no dialog.
//
// It uses SumatraPDF's command-line printing, which is reliable and silent:
//
//	SumatraPDF.exe -print-to "<printer>" -silent -print-settings "fit" <file.pdf>
//
// sumatraPath is the configured path to SumatraPDF.exe; if it's empty or doesn't
// exist, common install locations are searched automatically (recent SumatraPDF
// installs go to %LOCALAPPDATA%, not Program Files). printerName must match the
// printer's exact Windows name.
//
// printSettings maps to SumatraPDF's -print-settings (comma-separated). The
// important one for labels is "fit", which scales the page to the printer's
// paper/label size — the same as choosing "Fit" in Acrobat's print dialog.
func SilentPrint(sumatraPath, printerName, pdfPath, printSettings string) error {
	exe, err := resolveSumatra(sumatraPath)
	if err != nil {
		return err
	}

	args := []string{
		"-print-to", printerName,
		"-silent",
		"-exit-when-done",
	}
	if printSettings != "" {
		args = append(args, "-print-settings", printSettings)
	}
	args = append(args, pdfPath)

	cmd := exec.Command(exe, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("print to %q failed: %w (%s)", printerName, err, string(out))
	}
	return nil
}

// resolveSumatra returns the first SumatraPDF.exe that exists, checking the
// configured path first, then the standard install locations.
func resolveSumatra(configured string) (string, error) {
	var candidates []string
	if configured != "" {
		candidates = append(candidates, configured)
	}
	if la := os.Getenv("LOCALAPPDATA"); la != "" {
		candidates = append(candidates, filepath.Join(la, "SumatraPDF", "SumatraPDF.exe"))
	}
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		candidates = append(candidates, filepath.Join(pf, "SumatraPDF", "SumatraPDF.exe"))
	}
	if pf := os.Getenv("ProgramFiles(x86)"); pf != "" {
		candidates = append(candidates, filepath.Join(pf, "SumatraPDF", "SumatraPDF.exe"))
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf(
		"SumatraPDF.exe not found. Looked in: %v. Install it from "+
			"https://www.sumatrapdfreader.org/ or set sumatra_path in config.json",
		candidates,
	)
}
