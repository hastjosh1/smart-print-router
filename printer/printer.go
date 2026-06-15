package printer

import (
	"fmt"
	"os"
	"os/exec"
)

// SilentPrint sends a PDF to a named Windows printer with no dialog.
//
// It uses SumatraPDF's command-line printing, which is reliable and silent:
//
//	SumatraPDF.exe -print-to "<printer>" -silent <file.pdf>
//
// sumatraPath is the full path to SumatraPDF.exe (configurable so portable
// installs work). printerName must match the printer's exact Windows name.
func SilentPrint(sumatraPath, printerName, pdfPath string) error {
	if _, err := os.Stat(sumatraPath); err != nil {
		return fmt.Errorf("SumatraPDF not found at %q: %w", sumatraPath, err)
	}

	cmd := exec.Command(
		sumatraPath,
		"-print-to", printerName,
		"-silent",
		"-exit-when-done",
		pdfPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("print to %q failed: %w (%s)", printerName, err, string(out))
	}
	return nil
}
