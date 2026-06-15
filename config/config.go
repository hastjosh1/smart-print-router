package config

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// TwoUp describes how many labels to place side by side on one printed sheet.
type TwoUp struct {
	Enabled bool    `json:"enabled"`
	Columns int     `json:"columns"`
	GapMM   float64 `json:"gap_mm"`
}

// LabelProfile is a named sticker type. The router matches an incoming PDF's
// page size against this profile (within tolerance), then composes output at
// the physical sticker size so it lands 1:1 on the media.
//
//   - WidthMM/HeightMM  = the size of the INCOMING PDF page (used for matching).
//   - LabelWidthMM/LabelHeightMM = the PHYSICAL sticker size (used to size the
//     output cells). If 0, falls back to WidthMM/HeightMM (print at native size).
//
// Add new sticker types by appending entries to config.json.
type LabelProfile struct {
	Name        string  `json:"name"`
	WidthMM     float64 `json:"width_mm"`
	HeightMM    float64 `json:"height_mm"`
	ToleranceMM float64 `json:"tolerance_mm"`

	LabelWidthMM  float64 `json:"label_width_mm"`
	LabelHeightMM float64 `json:"label_height_mm"`

	TwoUp TwoUp `json:"two_up"`
}

// OutputCell returns the physical sticker size to compose each label at,
// falling back to the incoming page size when not configured.
func (p LabelProfile) OutputCell() (w, h float64) {
	w, h = p.LabelWidthMM, p.LabelHeightMM
	if w <= 0 {
		w = p.WidthMM
	}
	if h <= 0 {
		h = p.HeightMM
	}
	return w, h
}

// Config holds all user-editable settings, loaded from config.json.
type Config struct {
	ReportPrinter string `json:"report_printer"`
	LabelPrinter  string `json:"label_printer"`

	SumatraPath string `json:"sumatra_path"`
	LogFile     string `json:"log_file"`

	// Passed to SumatraPDF's -print-settings. "fit" scales each page to the
	// printer's label size (like Acrobat's "Fit") — needed for thermal labels.
	SumatraPrintSettings string `json:"sumatra_print_settings"`

	// Page-size cutoff for the label-vs-report decision (orientation-independent).
	LabelDetection struct {
		MaxWidthMM  float64 `json:"max_width_mm"`
		MaxHeightMM float64 `json:"max_height_mm"`
	} `json:"label_detection"`

	// Used when a label doesn't match any profile.
	DefaultLabel struct {
		TwoUp TwoUp `json:"two_up"`
	} `json:"default_label"`

	LabelProfiles []LabelProfile `json:"label_profiles"`
}

func defaults() Config {
	var c Config
	c.ReportPrinter = "Canon LBP2900B"
	c.LabelPrinter = "TVS LP46 Delite"
	c.LabelDetection.MaxWidthMM = 150
	c.LabelDetection.MaxHeightMM = 200
	c.DefaultLabel.TwoUp = TwoUp{Enabled: false, Columns: 1, GapMM: 0}
	c.SumatraPath = `C:/Program Files/SumatraPDF/SumatraPDF.exe`
	c.LogFile = `C:/SmartPrintRouter/logs/router.log`
	c.SumatraPrintSettings = "fit"
	return c
}

// Load reads config.json. If path is empty it looks next to the running
// executable, then falls back to built-in defaults. A missing file is not an
// error (defaults are used) so a print job never fails just for lack of config.
func Load(path string) (Config, error) {
	cfg := defaults()

	if path == "" {
		path = DefaultPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// CheckFileExists reports an error if no config file exists at path. Used to
// warn (in the log) when an edited config.json isn't where the program looks.
func CheckFileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("config file not found at %q", path)
	}
	return nil
}

// DefaultPath returns config.json next to the executable.
func DefaultPath() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "config.json")
	}
	return "config.json"
}

// MatchProfile returns the best label profile for a page of the given size, or
// nil if none match within tolerance. Orientation-independent: it compares the
// short/long sides so a 50x25 page matches a 25x50 profile.
func (c Config) MatchProfile(widthMM, heightMM float64) *LabelProfile {
	pShort, pLong := minmax(widthMM, heightMM)

	var best *LabelProfile
	bestDelta := math.MaxFloat64
	for i := range c.LabelProfiles {
		prof := &c.LabelProfiles[i]
		profShort, profLong := minmax(prof.WidthMM, prof.HeightMM)
		tol := prof.ToleranceMM
		if tol <= 0 {
			tol = 2
		}
		dShort := math.Abs(pShort - profShort)
		dLong := math.Abs(pLong - profLong)
		if dShort <= tol && dLong <= tol {
			if delta := dShort + dLong; delta < bestDelta {
				bestDelta = delta
				best = prof
			}
		}
	}
	return best
}

func minmax(a, b float64) (float64, float64) {
	if a <= b {
		return a, b
	}
	return b, a
}
