//go:build !windows

package rawlabel

import "errors"

// PrintRaw is only implemented on Windows; this stub lets the code build and be
// tested on other platforms (use -dryrun to inspect the generated TSPL there).
func PrintRaw(printerName string, data []byte) error {
	return errors.New("raw printing is only supported on Windows")
}
