package suptext

import (
	"encoding/binary"
	"testing"
)

func TestNewWindowsData_Valid(t *testing.T) {
	// Valid window definition with 2 windows
	bytes := make([]byte, 1+2*9) // Header + 2 windows
	bytes[0] = 2                 // NumWindows
	// Window 1
	bytes[1] = 0x01              // WinID
	binary.BigEndian.PutUint16(bytes[2:4], 100)  // Hpos
	binary.BigEndian.PutUint16(bytes[4:6], 200)  // Vpos
	binary.BigEndian.PutUint16(bytes[6:8], 300)  // Width
	binary.BigEndian.PutUint16(bytes[8:10], 400) // Height
	// Window 2
	bytes[10] = 0x02             // WinID
	binary.BigEndian.PutUint16(bytes[11:13], 500) // Hpos
	binary.BigEndian.PutUint16(bytes[13:15], 600) // Vpos
	binary.BigEndian.PutUint16(bytes[15:17], 700) // Width
	binary.BigEndian.PutUint16(bytes[17:19], 800) // Height

	wds, err := NewWindowsData(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if wds.NumWindows != 2 {
		t.Errorf("Expected NumWindows 2, got %d", wds.NumWindows)
	}
	if len(wds.Windows) != 2 {
		t.Errorf("Expected 2 windows, got %d", len(wds.Windows))
	}
	if wds.Windows[0].WinID != 0x01 {
		t.Errorf("Expected first window WinID 0x01, got 0x%02X", wds.Windows[0].WinID)
	}
	if wds.Windows[0].Width != 300 {
		t.Errorf("Expected first window Width 300, got %d", wds.Windows[0].Width)
	}
}

func TestNewWindowsData_TruncatedHeader(t *testing.T) {
	// Empty buffer
	bytes := []byte{}
	_, err := NewWindowsData(bytes)
	if err == nil {
		t.Error("Expected error for truncated header")
	}
}

func TestNewWindowsData_TruncatedData(t *testing.T) {
	// Says 2 windows but only 1 complete window
	bytes := make([]byte, 1+9) // Header + 1 window
	bytes[0] = 2                // NumWindows (but only 1 window present)
	bytes[1] = 0x01              // WinID
	binary.BigEndian.PutUint16(bytes[2:4], 100)
	binary.BigEndian.PutUint16(bytes[4:6], 200)
	binary.BigEndian.PutUint16(bytes[6:8], 300)
	binary.BigEndian.PutUint16(bytes[8:10], 400)

	_, err := NewWindowsData(bytes)
	if err == nil {
		t.Error("Expected error for truncated data")
	}
}

func TestNewWindowsData_Empty(t *testing.T) {
	// No windows
	bytes := []byte{0x00}
	wds, err := NewWindowsData(bytes)
	if err != nil {
		t.Fatalf("Expected no error for empty windows, got: %v", err)
	}
	if wds.NumWindows != 0 {
		t.Errorf("Expected NumWindows 0, got %d", wds.NumWindows)
	}
	if len(wds.Windows) != 0 {
		t.Errorf("Expected 0 windows, got %d", len(wds.Windows))
	}
}

func TestValidateWindowBounds_Valid(t *testing.T) {
	window := WindowDefinition{
		WinID:  0x01,
		Hpos:   100,
		Vpos:   200,
		Width:  300,
		Height: 400,
	}
	err := ValidateWindowBounds(window, 1920, 1080)
	if err != nil {
		t.Errorf("Expected no error for valid window, got: %v", err)
	}
}

func TestValidateWindowBounds_ExceedsWidth(t *testing.T) {
	window := WindowDefinition{
		WinID:  0x01,
		Hpos:   1900,
		Vpos:   200,
		Width:  300, // Hpos + Width = 2200 > 1920
		Height: 400,
	}
	err := ValidateWindowBounds(window, 1920, 1080)
	if err == nil {
		t.Error("Expected error for window exceeding width")
	}
}

func TestValidateWindowBounds_ExceedsHeight(t *testing.T) {
	window := WindowDefinition{
		WinID:  0x01,
		Hpos:   100,
		Vpos:   1000,
		Width:  300,
		Height: 200, // Vpos + Height = 1200 > 1080
	}
	err := ValidateWindowBounds(window, 1920, 1080)
	if err == nil {
		t.Error("Expected error for window exceeding height")
	}
}

func TestValidateWindowBounds_ZeroDimensions(t *testing.T) {
	window := WindowDefinition{
		WinID:  0x01,
		Hpos:   100,
		Vpos:   200,
		Width:  0, // Zero width
		Height: 400,
	}
	err := ValidateWindowBounds(window, 1920, 1080)
	if err == nil {
		t.Error("Expected error for zero dimensions")
	}
}

func TestNewWindowsDataWithBounds_CustomBounds(t *testing.T) {
	// Window that fits in custom bounds
	bytes := make([]byte, 1+9)
	bytes[0] = 1
	bytes[1] = 0x01
	binary.BigEndian.PutUint16(bytes[2:4], 50)  // Hpos
	binary.BigEndian.PutUint16(bytes[4:6], 50)  // Vpos
	binary.BigEndian.PutUint16(bytes[6:8], 100)  // Width
	binary.BigEndian.PutUint16(bytes[8:10], 100) // Height

	wds, err := NewWindowsDataWithBounds(bytes, 200, 200)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(wds.Windows) != 1 {
		t.Errorf("Expected 1 window, got %d", len(wds.Windows))
	}
}

