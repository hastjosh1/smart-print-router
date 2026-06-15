// Package rawlabel prints labels by rasterizing the composed PDF and sending it
// to a thermal printer in its native TSPL language, bypassing the Windows print
// driver entirely. This avoids the scaling/rotation that generic PDF printing
// (SumatraPDF + driver) applies to small label pages.
package rawlabel

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
)

// MonoImage is a 1-bit raster. Data holds Height rows of WidthBytes each, MSB
// first, with bit value 1 = black (the PBM convention).
type MonoImage struct {
	Width      int
	Height     int
	WidthBytes int
	Data       []byte
}

// Render rasterizes pdfPath to a 1-bit image at dpi, on a page of widthPt x
// heightPt points, using Ghostscript.
func Render(gsPath, pdfPath string, dpi int, widthPt, heightPt float64) (*MonoImage, error) {
	gs, err := ResolveGS(gsPath)
	if err != nil {
		return nil, err
	}

	tmp, err := os.CreateTemp("", "spr-*.pbm")
	if err != nil {
		return nil, err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	args := []string{
		"-dNOPAUSE", "-dBATCH", "-dSAFER", "-dQUIET",
		"-sDEVICE=pbmraw",
		"-r" + strconv.Itoa(dpi),
		"-dFIXEDMEDIA",
		fmt.Sprintf("-dDEVICEWIDTHPOINTS=%.2f", widthPt),
		fmt.Sprintf("-dDEVICEHEIGHTPOINTS=%.2f", heightPt),
		"-dPDFFitPage",
		"-sOutputFile=" + tmp.Name(),
		pdfPath,
	}
	if out, err := exec.Command(gs, args...).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ghostscript render failed: %w (%s)", err, string(out))
	}
	return parsePBM(tmp.Name())
}

func parsePBM(path string) (*MonoImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	magic, err := readToken(r)
	if err != nil {
		return nil, err
	}
	if magic != "P4" {
		return nil, fmt.Errorf("unexpected PBM magic %q (want P4)", magic)
	}
	w, err := readIntToken(r)
	if err != nil {
		return nil, err
	}
	h, err := readIntToken(r)
	if err != nil {
		return nil, err
	}

	widthBytes := (w + 7) / 8
	data := make([]byte, widthBytes*h)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("read PBM raster (%dx%d): %w", w, h, err)
	}
	return &MonoImage{Width: w, Height: h, WidthBytes: widthBytes, Data: data}, nil
}

// readToken reads one whitespace-delimited token, skipping leading whitespace
// and #-comments, and consumes exactly one trailing whitespace byte (so the
// byte after the final header token — the start of the raster — is preserved).
func readToken(r *bufio.Reader) (string, error) {
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == '#' {
			for {
				c, err := r.ReadByte()
				if err != nil {
					return "", err
				}
				if c == '\n' {
					break
				}
			}
			continue
		}
		if isSpace(b) {
			continue
		}
		token := []byte{b}
		for {
			c, err := r.ReadByte()
			if err != nil {
				return string(token), nil
			}
			if isSpace(c) {
				break
			}
			token = append(token, c)
		}
		return string(token), nil
	}
}

func readIntToken(r *bufio.Reader) (int, error) {
	t, err := readToken(r)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(t)
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
}

// ResolveGS finds the Ghostscript executable: the configured path first, then
// the system PATH, then standard install locations.
func ResolveGS(configured string) (string, error) {
	var candidates []string
	add := func(s string) {
		if s != "" {
			candidates = append(candidates, s)
		}
	}

	add(configured)

	if runtime.GOOS == "windows" {
		// On PATH (covers installs that registered themselves).
		for _, name := range []string{"gswin64c.exe", "gswin32c.exe", "gswin64c", "gswin32c", "gs.exe"} {
			if p, err := exec.LookPath(name); err == nil {
				add(p)
			}
		}
		// Standard install folders: both "gs" and "Ghostscript" dir names, both
		// "gsX.YY" and any subfolder layouts, 64- and 32-bit Program Files.
		bases := []string{
			`C:\Program Files\gs`, `C:\Program Files (x86)\gs`,
			`C:\Program Files\Ghostscript`, `C:\Program Files (x86)\Ghostscript`,
		}
		for _, base := range bases {
			for _, exe := range []string{"gswin64c.exe", "gswin32c.exe"} {
				for _, pat := range []string{
					filepath.Join(base, "gs*", "bin", exe),
					filepath.Join(base, "*", "bin", exe),
					filepath.Join(base, "bin", exe),
				} {
					m, _ := filepath.Glob(pat)
					sort.Sort(sort.Reverse(sort.StringSlice(m))) // newest version first
					candidates = append(candidates, m...)
				}
			}
		}
	} else {
		if p, err := exec.LookPath("gs"); err == nil {
			add(p)
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf(
		"Ghostscript not found (looked in: %v). Install the 64-bit build from "+
			"https://www.ghostscript.com/releases/gsdnld.html and re-run, or set "+
			"label_raw.gs_path in config.json to the full path of gswin64c.exe",
		candidates,
	)
}
