package config

import "testing"

func testConfig() Config {
	c := defaults()
	c.LabelProfiles = []LabelProfile{
		{
			Name: "Barcode", WidthMM: 39.9, HeightMM: 19.8, ToleranceMM: 4,
			TwoUp: TwoUp{Enabled: true, Columns: 2, GapMM: 2},
		},
		{
			Name: "Address", WidthMM: 100, HeightMM: 50, ToleranceMM: 3,
			TwoUp: TwoUp{Enabled: false, Columns: 1},
		},
	}
	return c
}

func TestMatchProfile(t *testing.T) {
	c := testConfig()

	tests := []struct {
		name     string
		w, h     float64
		wantName string // "" => no match
	}{
		{"exact barcode", 39.9, 19.8, "Barcode"},
		{"within tolerance", 41.0, 21.0, "Barcode"},
		{"orientation swapped", 19.8, 39.9, "Barcode"},
		{"address exact", 100, 50, "Address"},
		{"no match (A4)", 210, 297, ""},
		{"just outside tolerance", 45, 25, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.MatchProfile(tt.w, tt.h)
			if tt.wantName == "" {
				if got != nil {
					t.Fatalf("expected no match, got %q", got.Name)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected %q, got no match", tt.wantName)
			}
			if got.Name != tt.wantName {
				t.Fatalf("expected %q, got %q", tt.wantName, got.Name)
			}
		})
	}
}

func TestMatchProfilePicksClosest(t *testing.T) {
	c := defaults()
	c.LabelProfiles = []LabelProfile{
		{Name: "A", WidthMM: 40, HeightMM: 20, ToleranceMM: 10},
		{Name: "B", WidthMM: 42, HeightMM: 22, ToleranceMM: 10},
	}
	// 41.5x21.5 is within tolerance of both; B is closer (delta 1 vs 3).
	got := c.MatchProfile(41.5, 21.5)
	if got == nil || got.Name != "B" {
		t.Fatalf("expected closest match B, got %v", got)
	}
}
