package suptext

import (
	"encoding/binary"
	"testing"
)

func TestNewPresentationData_Valid(t *testing.T) {
	// Valid presentation composition with 1 composition object
	bytes := make([]byte, 11+8) // Header + 1 composition object
	binary.BigEndian.PutUint16(bytes[0:2], 1920)  // Width
	binary.BigEndian.PutUint16(bytes[2:4], 1080)  // Height
	bytes[4] = 0x18                                // Framerate
	binary.BigEndian.PutUint16(bytes[5:7], 1)     // Num
	bytes[7] = 0x80                                // State (epoch start)
	bytes[8] = 0x00                                // PaletteUpdate
	bytes[9] = 0x01                                // PaletteID
	bytes[10] = 1                                  // NumComps
	// Composition object
	binary.BigEndian.PutUint16(bytes[11:13], 0x1234) // ObjID
	bytes[13] = 0x01                                  // WinID
	bytes[14] = 0x00                                  // Cropped (no)
	binary.BigEndian.PutUint16(bytes[15:17], 100)    // Hpos
	binary.BigEndian.PutUint16(bytes[17:19], 200)    // Vpos

	pcs, err := NewPresentationData(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if pcs.Width != 1920 {
		t.Errorf("Expected Width 1920, got %d", pcs.Width)
	}
	if pcs.Height != 1080 {
		t.Errorf("Expected Height 1080, got %d", pcs.Height)
	}
	if pcs.State != 0x80 {
		t.Errorf("Expected State 0x80, got 0x%02X", pcs.State)
	}
	if len(pcs.Comps) != 1 {
		t.Errorf("Expected 1 composition object, got %d", len(pcs.Comps))
	}
	if pcs.Comps[0].ObjID != 0x1234 {
		t.Errorf("Expected ObjID 0x1234, got 0x%04X", pcs.Comps[0].ObjID)
	}
}

func TestNewPresentationData_TruncatedHeader(t *testing.T) {
	// Too short for header
	bytes := make([]byte, 10)
	_, err := NewPresentationData(bytes)
	if err == nil {
		t.Error("Expected error for truncated header")
	}
}

func TestNewPresentationData_TruncatedCompositionObjects(t *testing.T) {
	// Header says 2 objects but only 1 present
	bytes := make([]byte, 11+8) // Header + 1 object (but says 2)
	binary.BigEndian.PutUint16(bytes[0:2], 1920)
	binary.BigEndian.PutUint16(bytes[2:4], 1080)
	bytes[4] = 0x18
	binary.BigEndian.PutUint16(bytes[5:7], 1)
	bytes[7] = 0x80
	bytes[8] = 0x00
	bytes[9] = 0x01
	bytes[10] = 2 // Says 2 objects
	// Only 1 object
	binary.BigEndian.PutUint16(bytes[11:13], 0x1234)
	bytes[13] = 0x01
	bytes[14] = 0x00
	binary.BigEndian.PutUint16(bytes[15:17], 100)
	binary.BigEndian.PutUint16(bytes[17:19], 200)

	pcs, err := NewPresentationData(bytes)
	// Should handle gracefully
	if err != nil {
		t.Fatalf("Should handle truncated objects gracefully, got error: %v", err)
	}
	if len(pcs.Comps) != 1 {
		t.Errorf("Expected 1 composition object (truncated), got %d", len(pcs.Comps))
	}
}

func TestNewCompositionObject_Valid(t *testing.T) {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint16(bytes[0:2], 0x1234) // ObjID
	bytes[2] = 0x01                                  // WinID
	bytes[3] = 0x00                                  // Cropped (no)
	binary.BigEndian.PutUint16(bytes[4:6], 100)     // Hpos
	binary.BigEndian.PutUint16(bytes[6:8], 200)     // Vpos

	comp, err := NewCompositionObject(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if comp.ObjID != 0x1234 {
		t.Errorf("Expected ObjID 0x1234, got 0x%04X", comp.ObjID)
	}
	if comp.Cropped != 0 {
		t.Error("Expected Cropped to be 0")
	}
}

func TestNewCompositionObject_WithExtension(t *testing.T) {
	bytes := make([]byte, 16)
	binary.BigEndian.PutUint16(bytes[0:2], 0x1234) // ObjID
	bytes[2] = 0x01                                  // WinID
	bytes[3] = 0x01                                  // Cropped (yes)
	binary.BigEndian.PutUint16(bytes[4:6], 100)     // Hpos
	binary.BigEndian.PutUint16(bytes[6:8], 200)     // Vpos
	binary.BigEndian.PutUint16(bytes[8:10], 10)     // HCropPos
	binary.BigEndian.PutUint16(bytes[10:12], 20)   // VCropPos
	binary.BigEndian.PutUint16(bytes[12:14], 50)   // CropWidth
	binary.BigEndian.PutUint16(bytes[14:16], 60)   // CropHeight

	comp, err := NewCompositionObject(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if comp.Cropped != 1 {
		t.Error("Expected Cropped to be 1")
	}
	if comp.HCropPos != 10 {
		t.Errorf("Expected HCropPos 10, got %d", comp.HCropPos)
	}
	if comp.CropWidth != 50 {
		t.Errorf("Expected CropWidth 50, got %d", comp.CropWidth)
	}
}

func TestNewCompositionObject_Truncated(t *testing.T) {
	// Too short for basic object
	bytes := make([]byte, 6)
	_, err := NewCompositionObject(bytes)
	if err == nil {
		t.Error("Expected error for truncated object")
	}
}

func TestNewCompositionObject_TruncatedExtension(t *testing.T) {
	// Says cropped but extension truncated
	bytes := make([]byte, 12)
	binary.BigEndian.PutUint16(bytes[0:2], 0x1234)
	bytes[2] = 0x01
	bytes[3] = 0x01 // Cropped (yes)
	binary.BigEndian.PutUint16(bytes[4:6], 100)
	binary.BigEndian.PutUint16(bytes[6:8], 200)
	// Extension incomplete (only 4 bytes instead of 8)
	binary.BigEndian.PutUint16(bytes[8:10], 10)
	binary.BigEndian.PutUint16(bytes[10:12], 20)

	comp, err := NewCompositionObject(bytes)
	// Should handle gracefully
	if err != nil {
		t.Fatalf("Should handle truncated extension gracefully, got error: %v", err)
	}
	if comp.Cropped != 1 {
		t.Error("Expected Cropped to be 1")
	}
	// Extension fields should be zero
	if comp.HCropPos != 0 {
		t.Errorf("Expected HCropPos 0 (truncated), got %d", comp.HCropPos)
	}
}

func TestNewPresentationData_MultipleCompositionObjects(t *testing.T) {
	// 2 composition objects
	bytes := make([]byte, 11+8+8) // Header + 2 objects
	binary.BigEndian.PutUint16(bytes[0:2], 1920)
	binary.BigEndian.PutUint16(bytes[2:4], 1080)
	bytes[4] = 0x18
	binary.BigEndian.PutUint16(bytes[5:7], 1)
	bytes[7] = 0x80
	bytes[8] = 0x00
	bytes[9] = 0x01
	bytes[10] = 2 // NumComps
	// Object 1
	binary.BigEndian.PutUint16(bytes[11:13], 0x1234)
	bytes[13] = 0x01
	bytes[14] = 0x00
	binary.BigEndian.PutUint16(bytes[15:17], 100)
	binary.BigEndian.PutUint16(bytes[17:19], 200)
	// Object 2
	binary.BigEndian.PutUint16(bytes[19:21], 0x5678)
	bytes[21] = 0x02
	bytes[22] = 0x00
	binary.BigEndian.PutUint16(bytes[23:25], 300)
	binary.BigEndian.PutUint16(bytes[25:27], 400)

	pcs, err := NewPresentationData(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(pcs.Comps) != 2 {
		t.Errorf("Expected 2 composition objects, got %d", len(pcs.Comps))
	}
	if pcs.Comps[1].ObjID != 0x5678 {
		t.Errorf("Expected second ObjID 0x5678, got 0x%04X", pcs.Comps[1].ObjID)
	}
}

