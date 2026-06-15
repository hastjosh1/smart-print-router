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
//	SumatraPDF.exe -print-to "<printer>" -silent -print-settings "fit" <file.pdf>
//
// sumatraPath is the full path to SumatraPDF.exe (configurable so portable
// installs work). printerName must match the printer's exact Windows name.
//
// printSettings maps to SumatraPDF's -print-settings (comma-separated). The
// important one for labels is "fit", which scales the page to the printer's
// paper/label size — the same as choosing "Fit" in Acrobat's print dialog.
// Other useful values: "shrink" (only shrink if too big), "noscale" (1:1),
// "monochrome". Empty string omits the flag (uses SumatraPDF defaults).
func SilentPrint(sumatraPath, printerName, pdfPath, printSettings string) error {
	if _, err := os.Stat(sumatraPath); err != nil {
		return fmt.Errorf("SumatraPDF not found at %q: %w", sumatraPath, err)
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

	cmd := exec.Command(sumatraPath, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("print to %q failed: %w (%s)", printerName, err, string(out))
	}
	return nil
}
