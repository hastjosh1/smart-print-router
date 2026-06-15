package main

import "testing"

func TestJobNameHint(t *testing.T) {
	tests := []struct {
		name string
		job  string
		want route
	}{
		{"barcode keyword", "barcode-12345", routeLabel},
		{"label keyword", "Patient Label", routeLabel},
		{"sticker keyword", "sticker_run", routeLabel},
		{"report keyword", "Lab Report PDF", routeReport},
		{"invoice keyword", "invoice-99", routeReport},
		{"order keyword", "work order", routeReport},
		{"no keyword", "document", routeUnknown},
		{"empty", "", routeUnknown},
		{"case insensitive", "BARCODE", routeLabel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jobNameHint(tt.job); got != tt.want {
				t.Fatalf("jobNameHint(%q) = %v, want %v", tt.job, got, tt.want)
			}
		})
	}
}
