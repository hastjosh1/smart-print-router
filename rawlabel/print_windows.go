//go:build windows

package rawlabel

import (
	"fmt"

	"github.com/alexbrainman/printer"
)

// PrintRaw sends raw bytes (e.g. a TSPL program) straight to a Windows printer
// using the RAW datatype, bypassing the printer driver's rendering.
func PrintRaw(printerName string, data []byte) error {
	p, err := printer.Open(printerName)
	if err != nil {
		return fmt.Errorf("open printer %q: %w", printerName, err)
	}
	defer p.Close()

	if err := p.StartRawDocument("Smart Print Router label"); err != nil {
		return fmt.Errorf("start raw document: %w", err)
	}
	defer p.EndDocument()

	if _, err := p.Write(data); err != nil {
		return fmt.Errorf("write %d bytes to printer: %w", len(data), err)
	}
	return nil
}
