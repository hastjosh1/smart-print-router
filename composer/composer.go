package composer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

const mmToPoints = 72.0 / 25.4

// ComposeNUp places `columns` labels side by side on a single sheet sized to
// the physical label stock, and returns the path to a new temp PDF.
//
//   - labelWmm/labelHmm are the size of ONE label (from the matched profile or
//     detected page size).
//   - If the source PDF has fewer pages than `columns`, page 1 is duplicated so
//     the row is filled (common case: browser emits one label, stock holds two
//     across).
//   - The output sheet width = columns*labelW + (columns-1)*gap; height = labelH.
//
// Gap handling is approximate: pdfcpu fits each source page into an equal grid
// cell. Sizing the sheet to include the gap spreads the labels apart; fine-tune
// `gap_mm` against a real print if the spacing is off.
func ComposeNUp(srcPath string, pages, columns int, gapMM, labelWmm, labelHmm float64) (string, error) {
	if columns < 1 {
		columns = 1
	}

	working := srcPath
	if pages < columns {
		dup, err := repeatFirstPage(srcPath, columns)
		if err != nil {
			return "", fmt.Errorf("duplicate label to %d copies: %w", columns, err)
		}
		working = dup
		defer os.Remove(dup)
	}

	// pdfcpu's grid mode treats `dimensions` as the size of ONE cell and grows
	// the sheet to grid (rows x cols) of that cell. So we pass the single-label
	// size; the resulting sheet becomes columns x labelW wide.
	cellWPt := labelWmm * mmToPoints
	cellHPt := labelHmm * mmToPoints

	// gap between labels is added as a margin inside each cell. Splitting it in
	// half means two adjacent cells contribute gapMM of combined spacing.
	marginPt := gapMM / 2 * mmToPoints
	if marginPt < 0 {
		marginPt = 0
	}

	outPath := tempPath("compose-*.pdf")

	conf := model.NewDefaultConfiguration()

	// 1 row x `columns` cells, no border, per-cell size = one label.
	desc := fmt.Sprintf("dimensions:%s %s, border:off, margin:%s",
		strconv.FormatFloat(cellWPt, 'f', 2, 64),
		strconv.FormatFloat(cellHPt, 'f', 2, 64),
		strconv.FormatFloat(marginPt, 'f', 2, 64),
	)
	nup, err := api.PDFGridConfig(1, columns, desc, conf)
	if err != nil {
		return "", fmt.Errorf("grid config: %w", err)
	}

	if err := api.NUpFile([]string{working}, outPath, nil, nup, conf); err != nil {
		return "", fmt.Errorf("compose %d-up: %w", columns, err)
	}
	return outPath, nil
}

// repeatFirstPage writes a PDF containing page 1 of src repeated `n` times.
func repeatFirstPage(src string, n int) (string, error) {
	out := tempPath("dup-*.pdf")
	pages := make([]string, n)
	for i := range pages {
		pages[i] = "1"
	}
	conf := model.NewDefaultConfiguration()
	if err := api.CollectFile(src, out, pages, conf); err != nil {
		return "", err
	}
	return out, nil
}

func tempPath(pattern string) string {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return filepath.Join(os.TempDir(), "spr-"+pattern)
	}
	name := f.Name()
	f.Close()
	return name
}
