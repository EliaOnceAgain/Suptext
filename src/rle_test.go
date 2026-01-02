package suptext

import (
	"testing"
)

func TestRLEDecode_Simple(t *testing.T) {
	// Simple RLE: single pixel, line end
	bytes := []byte{0x01, 0x00, 0x00}
	img, err := RLEDecode(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(img) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(img))
	}
	if len(img[0]) != 1 {
		t.Fatalf("Expected line length 1, got %d", len(img[0]))
	}
	if img[0][0] != 0x01 {
		t.Errorf("Expected pixel value 0x01, got 0x%02X", img[0][0])
	}
}

func TestRLEDecode_MultipleLines(t *testing.T) {
	// Two lines: [1, 2] and [3, 4]
	bytes := []byte{
		0x01, 0x02, 0x00, 0x00, // Line 1
		0x03, 0x04, 0x00, 0x00, // Line 2
	}
	img, err := RLEDecode(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(img) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(img))
	}
	if len(img[0]) != 2 || img[0][0] != 0x01 || img[0][1] != 0x02 {
		t.Error("First line incorrect")
	}
	if len(img[1]) != 2 || img[1][0] != 0x03 || img[1][1] != 0x04 {
		t.Error("Second line incorrect")
	}
}

func TestRLEDecode_RLE_ZeroColor(t *testing.T) {
	// RLE: 3 zeros (0x00, 0x03 means 3 zeros, no color)
	bytes := []byte{0x00, 0x03, 0x00, 0x00}
	img, err := RLEDecode(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(img) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(img))
	}
	if len(img[0]) != 3 {
		t.Fatalf("Expected line length 3, got %d", len(img[0]))
	}
	for i := 0; i < 3; i++ {
		if img[0][i] != 0 {
			t.Errorf("Expected pixel %d to be 0, got 0x%02X", i, img[0][i])
		}
	}
}

func TestRLEDecode_RLE_WithColor(t *testing.T) {
	// RLE: 5 pixels of color 0x42 (0x00, 0x85, 0x05 means 5 pixels of color 0x42)
	// Format: 0x00 (RLE marker), 0x85 (has_color flag + count 5), 0x42 (color)
	bytes := []byte{0x00, 0x85, 0x42, 0x00, 0x00}
	img, err := RLEDecode(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(img) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(img))
	}
	if len(img[0]) != 5 {
		t.Fatalf("Expected line length 5, got %d", len(img[0]))
	}
	for i := 0; i < 5; i++ {
		if img[0][i] != 0x42 {
			t.Errorf("Expected pixel %d to be 0x42, got 0x%02X", i, img[0][i])
		}
	}
}

func TestRLEDecode_Malformed_Truncated(t *testing.T) {
	// Truncated: starts RLE but missing bytes
	bytes := []byte{0x00, 0x85} // Missing color byte
	_, err := RLEDecode(bytes)
	// Should handle gracefully or return error
	if err == nil {
		// If it doesn't error, it should at least not panic
		t.Log("No error returned for truncated RLE (may be acceptable)")
	}
}

func TestRLEDecode_Malformed_UnterminatedLine(t *testing.T) {
	// Line not terminated
	bytes := []byte{0x01, 0x02, 0x03}
	_, err := RLEDecode(bytes)
	if err == nil {
		t.Error("Expected error for unterminated line")
	}
}

func TestRLEDecode_Empty(t *testing.T) {
	bytes := []byte{}
	img, err := RLEDecode(bytes)
	if err != nil {
		t.Fatalf("Expected no error for empty input, got: %v", err)
	}
	if len(img) != 0 {
		t.Errorf("Expected 0 lines, got %d", len(img))
	}
}

func TestCreateImage_Valid(t *testing.T) {
	pixels := [][]uint8{
		{1, 2, 3},
		{4, 5, 6},
	}
	var palettes [256]PaletteDefinition
	palettes[1] = PaletteDefinition{Y: 100, Cr: 110, Cb: 120, A: 255}
	palettes[2] = PaletteDefinition{Y: 200, Cr: 210, Cb: 220, A: 255}
	palettes[3] = PaletteDefinition{Y: 50, Cr: 60, Cb: 70, A: 128}
	palettes[4] = PaletteDefinition{Y: 150, Cr: 160, Cb: 170, A: 255}
	palettes[5] = PaletteDefinition{Y: 250, Cr: 255, Cb: 255, A: 255}
	palettes[6] = PaletteDefinition{Y: 10, Cr: 20, Cb: 30, A: 64}

	img, err := CreateImage(pixels, palettes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if img.Bounds().Dx() != 3 {
		t.Errorf("Expected width 3, got %d", img.Bounds().Dx())
	}
	if img.Bounds().Dy() != 2 {
		t.Errorf("Expected height 2, got %d", img.Bounds().Dy())
	}
}

func TestCreateImage_Empty(t *testing.T) {
	pixels := [][]uint8{}
	var palettes [256]PaletteDefinition
	_, err := CreateImage(pixels, palettes)
	if err == nil {
		t.Error("Expected error for empty pixels")
	}
}

func TestCreateImage_ZeroWidth(t *testing.T) {
	pixels := [][]uint8{{}}
	var palettes [256]PaletteDefinition
	_, err := CreateImage(pixels, palettes)
	if err == nil {
		t.Error("Expected error for zero width")
	}
}

func TestRLEDecode_Complex(t *testing.T) {
	// Complex: mix of single pixels and RLE
	// Line 1: [1, 0, 0, 0, 2] (1, then 3 zeros, then 2)
	// Line 2: [3, 4]
	bytes := []byte{
		0x01,           // Single pixel 1
		0x00, 0x03,     // 3 zeros
		0x02,           // Single pixel 2
		0x00, 0x00,     // Line end
		0x03, 0x04,     // Line 2
		0x00, 0x00,     // Line end
	}
	img, err := RLEDecode(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(img) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(img))
	}
	expectedLine1 := []uint8{1, 0, 0, 0, 2}
	if len(img[0]) != len(expectedLine1) {
		t.Fatalf("Line 1: expected length %d, got %d", len(expectedLine1), len(img[0]))
	}
	for i, v := range expectedLine1 {
		if img[0][i] != v {
			t.Errorf("Line 1[%d]: expected %d, got %d", i, v, img[0][i])
		}
	}
}

