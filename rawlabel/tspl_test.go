package rawlabel

import (
	"bytes"
	"testing"
)

func TestBuildTSPLHeaderAndInversion(t *testing.T) {
	// 8x1 image: first 4 px black (PBM bits 1111), last 4 white (0000) => 0xF0.
	img := &MonoImage{Width: 8, Height: 1, WidthBytes: 1, Data: []byte{0xF0}}

	out := BuildTSPL(img, Options{
		WidthMM: 101.6, HeightMM: 25.4, GapMM: 2,
		Direction: 1, Density: 8, Speed: 4, Copies: 1,
	})

	for _, want := range [][]byte{
		[]byte("SIZE 101.60 mm,25.40 mm\r\n"),
		[]byte("GAP 2.00 mm,0 mm\r\n"),
		[]byte("DIRECTION 1\r\n"),
		[]byte("BITMAP 0,0,1,1,0,"),
		[]byte("PRINT 1,1\r\n"),
	} {
		if !bytes.Contains(out, want) {
			t.Errorf("TSPL output missing %q", want)
		}
	}

	// The single raster byte must be inverted: 0xF0 (PBM black=1) -> 0x0F (TSPL black=0).
	marker := []byte("BITMAP 0,0,1,1,0,")
	idx := bytes.Index(out, marker)
	if idx < 0 {
		t.Fatal("no BITMAP marker")
	}
	got := out[idx+len(marker)]
	if got != 0x0F {
		t.Errorf("raster byte = %#02x, want 0x0f (inverted)", got)
	}
}

func TestBuildTSPLCopiesDefault(t *testing.T) {
	img := &MonoImage{Width: 8, Height: 1, WidthBytes: 1, Data: []byte{0x00}}
	out := BuildTSPL(img, Options{WidthMM: 50, HeightMM: 25, Copies: 0})
	if !bytes.Contains(out, []byte("PRINT 1,1\r\n")) {
		t.Error("copies=0 should default to PRINT 1,1")
	}
}
