package detector

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// pointsToMM converts PDF user-space units (1/72 inch) to millimetres.
const pointsToMM = 25.4 / 72.0

// PageSize describes the first page of a PDF in millimetres.
type PageSize struct {
	WidthMM  float64
	HeightMM float64
	Pages    int
}

// Detect reads the first page dimensions of the PDF at path.
// Dimensions are normalised so WidthMM <= HeightMM is NOT enforced — the
// raw page width/height are returned as authored.
func Detect(path string) (PageSize, error) {
	dims, err := api.PageDimsFile(path)
	if err != nil {
		return PageSize{}, fmt.Errorf("read page dims: %w", err)
	}
	if len(dims) == 0 {
		return PageSize{}, fmt.Errorf("pdf has no pages")
	}

	first := dims[0]
	return PageSize{
		WidthMM:  first.Width * pointsToMM,
		HeightMM: first.Height * pointsToMM,
		Pages:    len(dims),
	}, nil
}

// IsLabel reports whether the page is small enough to be a label rather than
// an A4/Letter report. Orientation-independent: it compares the smaller and
// larger sides against the configured maxima.
func IsLabel(ps PageSize, maxWidthMM, maxHeightMM float64) bool {
	shortSide, longSide := ps.WidthMM, ps.HeightMM
	if shortSide > longSide {
		shortSide, longSide = longSide, shortSide
	}
	maxShort, maxLong := maxWidthMM, maxHeightMM
	if maxShort > maxLong {
		maxShort, maxLong = maxLong, maxShort
	}
	return shortSide <= maxShort && longSide <= maxLong
}
