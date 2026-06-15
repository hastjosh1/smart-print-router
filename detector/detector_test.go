package detector

import "testing"

func TestIsLabel(t *testing.T) {
	const maxW, maxH = 150.0, 200.0

	tests := []struct {
		name string
		ps   PageSize
		want bool
	}{
		{"50x25 barcode", PageSize{WidthMM: 50, HeightMM: 25}, true},
		{"40x20 barcode", PageSize{WidthMM: 39.9, HeightMM: 19.8}, true},
		{"A4 portrait", PageSize{WidthMM: 210, HeightMM: 297}, false},
		{"A4 landscape", PageSize{WidthMM: 297, HeightMM: 210}, false},
		{"Letter", PageSize{WidthMM: 215.9, HeightMM: 279.4}, false},
		{"label landscape orientation", PageSize{WidthMM: 25, HeightMM: 50}, true},
		{"at boundary", PageSize{WidthMM: 150, HeightMM: 200}, true},
		{"large label still within bounds", PageSize{WidthMM: 151, HeightMM: 100}, true},
		{"short side past max", PageSize{WidthMM: 160, HeightMM: 210}, false},
		{"long side past max", PageSize{WidthMM: 100, HeightMM: 210}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLabel(tt.ps, maxW, maxH); got != tt.want {
				t.Fatalf("IsLabel(%v) = %v, want %v", tt.ps, got, tt.want)
			}
		})
	}
}

func TestDetectRealPDF(t *testing.T) {
	// Optional integration check: runs only if a sample PDF is present.
	const sample = "testdata/sample-barcode.pdf"
	ps, err := Detect(sample)
	if err != nil {
		t.Skipf("no sample PDF (%v)", err)
	}
	if ps.Pages < 1 {
		t.Fatalf("expected >=1 page, got %d", ps.Pages)
	}
}
