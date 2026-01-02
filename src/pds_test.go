package suptext

import (
	"testing"
)

func TestNewPaletteData_Valid(t *testing.T) {
	// Valid palette with 2 entries
	// Header: ID (1) + Version (1) = 2 bytes
	// Entry 1: ID (0) + Y (10) + Cr (20) + Cb (30) + A (40) = 5 bytes
	// Entry 2: ID (1) + Y (50) + Cr (60) + Cb (70) + A (80) = 5 bytes
	bytes := []byte{
		0x01,       // ID
		0x02,       // Version
		0x00, 10, 20, 30, 40, // Entry 0
		0x01, 50, 60, 70, 80, // Entry 1
	}

	pds, err := NewPaletteData(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if pds.ID != 0x01 {
		t.Errorf("Expected ID 0x01, got 0x%02X", pds.ID)
	}
	if pds.Version != 0x02 {
		t.Errorf("Expected Version 0x02, got 0x%02X", pds.Version)
	}
	if pds.NumPalettes != 2 {
		t.Errorf("Expected NumPalettes 2, got %d", pds.NumPalettes)
	}
	if pds.Palettes[0].Y != 10 {
		t.Errorf("Expected Palettes[0].Y 10, got %d", pds.Palettes[0].Y)
	}
	if pds.Palettes[1].Y != 50 {
		t.Errorf("Expected Palettes[1].Y 50, got %d", pds.Palettes[1].Y)
	}
}

func TestNewPaletteData_TruncatedHeader(t *testing.T) {
	// Too short for header
	bytes := []byte{0x01}
	_, err := NewPaletteData(bytes)
	if err == nil {
		t.Error("Expected error for truncated header")
	}
}

func TestNewPaletteData_TruncatedData(t *testing.T) {
	// Header OK but data not multiple of 5
	bytes := []byte{
		0x01,       // ID
		0x02,       // Version
		0x00, 10, 20, 30, // Incomplete entry (only 4 bytes)
	}
	_, err := NewPaletteData(bytes)
	if err == nil {
		t.Error("Expected error for truncated data")
	}
}

func TestNewPaletteData_InvalidPaletteID(t *testing.T) {
	// Palette entry with ID > 255 (should be skipped with warning)
	bytes := []byte{
		0x01,       // ID
		0x02,       // Version
		0xFF, 10, 20, 30, 40, // Valid entry with ID 255
	}
	pds, err := NewPaletteData(bytes)
	if err != nil {
		t.Fatalf("Should handle invalid palette ID gracefully, got error: %v", err)
	}
	if pds.NumPalettes != 1 {
		t.Errorf("Expected NumPalettes 1, got %d", pds.NumPalettes)
	}
}

func TestNewPaletteData_Empty(t *testing.T) {
	// Only header, no entries
	bytes := []byte{
		0x01, // ID
		0x02, // Version
	}
	pds, err := NewPaletteData(bytes)
	if err != nil {
		t.Fatalf("Expected no error for empty palette, got: %v", err)
	}
	if pds.NumPalettes != 0 {
		t.Errorf("Expected NumPalettes 0, got %d", pds.NumPalettes)
	}
}

func TestNewPaletteData_MultipleEntries(t *testing.T) {
	// Create palette with 10 entries
	bytes := make([]byte, 2+10*5) // Header + 10 entries
	bytes[0] = 0x01                // ID
	bytes[1] = 0x02                // Version
	for i := 0; i < 10; i++ {
		offset := 2 + i*5
		bytes[offset] = byte(i)     // Entry ID
		bytes[offset+1] = byte(i * 10) // Y
		bytes[offset+2] = byte(i * 20) // Cr
		bytes[offset+3] = byte(i * 30) // Cb
		bytes[offset+4] = byte(i * 40) // A
	}

	pds, err := NewPaletteData(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if pds.NumPalettes != 10 {
		t.Errorf("Expected NumPalettes 10, got %d", pds.NumPalettes)
	}
	for i := 0; i < 10; i++ {
		if pds.Palettes[i].Y != byte(i*10) {
			t.Errorf("Expected Palettes[%d].Y %d, got %d", i, i*10, pds.Palettes[i].Y)
		}
	}
}

