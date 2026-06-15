package rawlabel

import (
	"bytes"
	"fmt"
)

// Options configures the TSPL output for one label sheet (one row = the full
// 2-across media).
type Options struct {
	WidthMM   float64 // full media width (across the print head)
	HeightMM  float64 // one row height (feed direction)
	GapMM     float64 // gap between rows (feed direction)
	Direction int     // TSPL DIRECTION (0 or 1) — flip if printed upside down
	Density   int     // TSPL DENSITY (0-15), darkness
	Speed     int     // TSPL SPEED (printer-dependent); 0 = omit
	Copies    int     // number of sheets to print
}

// BuildTSPL wraps a 1-bit image in a TSPL program ready to send raw to the
// printer. The image's black pixels (bit 1 in PBM) are inverted to TSPL's
// convention (bit 0 = printed dot).
func BuildTSPL(img *MonoImage, opt Options) []byte {
	var b bytes.Buffer

	fmt.Fprintf(&b, "SIZE %.2f mm,%.2f mm\r\n", opt.WidthMM, opt.HeightMM)
	fmt.Fprintf(&b, "GAP %.2f mm,0 mm\r\n", opt.GapMM)
	fmt.Fprintf(&b, "DIRECTION %d\r\n", opt.Direction)
	if opt.Speed > 0 {
		fmt.Fprintf(&b, "SPEED %d\r\n", opt.Speed)
	}
	fmt.Fprintf(&b, "DENSITY %d\r\n", opt.Density)
	b.WriteString("REFERENCE 0,0\r\n")
	b.WriteString("CLS\r\n")

	// PBM: bit 1 = black. TSPL BITMAP mode 0: bit 0 = printed dot. Invert.
	inv := make([]byte, len(img.Data))
	for i, v := range img.Data {
		inv[i] = ^v
	}

	fmt.Fprintf(&b, "BITMAP 0,0,%d,%d,0,", img.WidthBytes, img.Height)
	b.Write(inv)
	b.WriteString("\r\n")

	copies := opt.Copies
	if copies < 1 {
		copies = 1
	}
	fmt.Fprintf(&b, "PRINT %d,1\r\n", copies)
	return b.Bytes()
}
