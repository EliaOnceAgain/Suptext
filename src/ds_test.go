package suptext

import (
	"encoding/binary"
	"testing"
)

func TestValidateWindowCompositionLinkage_Valid(t *testing.T) {
	// Create a DisplaySet with valid PCS and WDS
	ds := DisplaySet{}
	
	// Create PCS with composition objects
	pcsBytes := make([]byte, 11+8) // Header + 1 composition object
	binary.BigEndian.PutUint16(pcsBytes[0:2], 1920)
	binary.BigEndian.PutUint16(pcsBytes[2:4], 1080)
	pcsBytes[4] = 0x18
	binary.BigEndian.PutUint16(pcsBytes[5:7], 1)
	pcsBytes[7] = 0x80
	pcsBytes[8] = 0x00
	pcsBytes[9] = 0x01
	pcsBytes[10] = 1 // NumComps
	binary.BigEndian.PutUint16(pcsBytes[11:13], 0x1234) // ObjID
	pcsBytes[13] = 0x01                                  // WinID
	pcsBytes[14] = 0x00                                  // Cropped
	binary.BigEndian.PutUint16(pcsBytes[15:17], 100)
	binary.BigEndian.PutUint16(pcsBytes[17:19], 200)
	
	pcsData, err := NewPresentationData(pcsBytes)
	if err != nil {
		t.Fatalf("Failed to create PCS data: %v", err)
	}
	ds.PCS = Section{Data: pcsData}
	
	// Create WDS with matching window
	wdsBytes := make([]byte, 1+9)
	wdsBytes[0] = 1 // NumWindows
	wdsBytes[1] = 0x01 // WinID
	binary.BigEndian.PutUint16(wdsBytes[2:4], 100)
	binary.BigEndian.PutUint16(wdsBytes[4:6], 200)
	binary.BigEndian.PutUint16(wdsBytes[6:8], 300)
	binary.BigEndian.PutUint16(wdsBytes[8:10], 400)
	
	wdsData, err := NewWindowsData(wdsBytes)
	if err != nil {
		t.Fatalf("Failed to create WDS data: %v", err)
	}
	ds.WDS = Section{Data: wdsData}
	
	// Should not return error for valid linkage
	err = ds.ValidateWindowCompositionLinkage()
	if err != nil {
		t.Errorf("Expected no error for valid linkage, got: %v", err)
	}
}

func TestValidateWindowCompositionLinkage_InvalidWinID(t *testing.T) {
	// Create a DisplaySet with PCS referencing non-existent window
	ds := DisplaySet{}
	
	// Create PCS with composition object referencing WinID 0x02
	pcsBytes := make([]byte, 11+8)
	binary.BigEndian.PutUint16(pcsBytes[0:2], 1920)
	binary.BigEndian.PutUint16(pcsBytes[2:4], 1080)
	pcsBytes[4] = 0x18
	binary.BigEndian.PutUint16(pcsBytes[5:7], 1)
	pcsBytes[7] = 0x80
	pcsBytes[8] = 0x00
	pcsBytes[9] = 0x01
	pcsBytes[10] = 1
	binary.BigEndian.PutUint16(pcsBytes[11:13], 0x1234)
	pcsBytes[13] = 0x02 // WinID 0x02 (doesn't exist in WDS)
	pcsBytes[14] = 0x00
	binary.BigEndian.PutUint16(pcsBytes[15:17], 100)
	binary.BigEndian.PutUint16(pcsBytes[17:19], 200)
	
	pcsData, err := NewPresentationData(pcsBytes)
	if err != nil {
		t.Fatalf("Failed to create PCS data: %v", err)
	}
	ds.PCS = Section{Data: pcsData}
	
	// Create WDS with only WinID 0x01
	wdsBytes := make([]byte, 1+9)
	wdsBytes[0] = 1
	wdsBytes[1] = 0x01 // Only WinID 0x01
	binary.BigEndian.PutUint16(wdsBytes[2:4], 100)
	binary.BigEndian.PutUint16(wdsBytes[4:6], 200)
	binary.BigEndian.PutUint16(wdsBytes[6:8], 300)
	binary.BigEndian.PutUint16(wdsBytes[8:10], 400)
	
	wdsData, err := NewWindowsData(wdsBytes)
	if err != nil {
		t.Fatalf("Failed to create WDS data: %v", err)
	}
	ds.WDS = Section{Data: wdsData}
	
	// Should not return error (just logs warning)
	err = ds.ValidateWindowCompositionLinkage()
	if err != nil {
		t.Errorf("Expected no error (only warning), got: %v", err)
	}
}

func TestValidateWindowCompositionLinkage_MissingPCS(t *testing.T) {
	// DisplaySet without PCS
	ds := DisplaySet{}
	
	wdsBytes := make([]byte, 1+9)
	wdsBytes[0] = 1
	wdsBytes[1] = 0x01
	binary.BigEndian.PutUint16(wdsBytes[2:4], 100)
	binary.BigEndian.PutUint16(wdsBytes[4:6], 200)
	binary.BigEndian.PutUint16(wdsBytes[6:8], 300)
	binary.BigEndian.PutUint16(wdsBytes[8:10], 400)
	
	wdsData, err := NewWindowsData(wdsBytes)
	if err != nil {
		t.Fatalf("Failed to create WDS data: %v", err)
	}
	ds.WDS = Section{Data: wdsData}
	
	// Should return nil (can't validate without PCS)
	err = ds.ValidateWindowCompositionLinkage()
	if err != nil {
		t.Errorf("Expected nil error when PCS is missing, got: %v", err)
	}
}

func TestValidateWindowCompositionLinkage_MissingWDS(t *testing.T) {
	// DisplaySet without WDS
	ds := DisplaySet{}
	
	pcsBytes := make([]byte, 11+8)
	binary.BigEndian.PutUint16(pcsBytes[0:2], 1920)
	binary.BigEndian.PutUint16(pcsBytes[2:4], 1080)
	pcsBytes[4] = 0x18
	binary.BigEndian.PutUint16(pcsBytes[5:7], 1)
	pcsBytes[7] = 0x80
	pcsBytes[8] = 0x00
	pcsBytes[9] = 0x01
	pcsBytes[10] = 1
	binary.BigEndian.PutUint16(pcsBytes[11:13], 0x1234)
	pcsBytes[13] = 0x01
	pcsBytes[14] = 0x00
	binary.BigEndian.PutUint16(pcsBytes[15:17], 100)
	binary.BigEndian.PutUint16(pcsBytes[17:19], 200)
	
	pcsData, err := NewPresentationData(pcsBytes)
	if err != nil {
		t.Fatalf("Failed to create PCS data: %v", err)
	}
	ds.PCS = Section{Data: pcsData}
	
	// Should return nil (can't validate without WDS)
	err = ds.ValidateWindowCompositionLinkage()
	if err != nil {
		t.Errorf("Expected nil error when WDS is missing, got: %v", err)
	}
}

func TestValidateWindowCompositionLinkage_MultipleComps(t *testing.T) {
	// Test with multiple composition objects, some valid, some invalid
	ds := DisplaySet{}
	
	// Create PCS with 2 composition objects
	pcsBytes := make([]byte, 11+8+8) // Header + 2 objects
	binary.BigEndian.PutUint16(pcsBytes[0:2], 1920)
	binary.BigEndian.PutUint16(pcsBytes[2:4], 1080)
	pcsBytes[4] = 0x18
	binary.BigEndian.PutUint16(pcsBytes[5:7], 1)
	pcsBytes[7] = 0x80
	pcsBytes[8] = 0x00
	pcsBytes[9] = 0x01
	pcsBytes[10] = 2 // NumComps
	// Object 1 - valid WinID
	binary.BigEndian.PutUint16(pcsBytes[11:13], 0x1234)
	pcsBytes[13] = 0x01 // WinID 0x01 (valid)
	pcsBytes[14] = 0x00
	binary.BigEndian.PutUint16(pcsBytes[15:17], 100)
	binary.BigEndian.PutUint16(pcsBytes[17:19], 200)
	// Object 2 - invalid WinID
	binary.BigEndian.PutUint16(pcsBytes[19:21], 0x5678)
	pcsBytes[21] = 0x03 // WinID 0x03 (invalid)
	pcsBytes[22] = 0x00
	binary.BigEndian.PutUint16(pcsBytes[23:25], 300)
	binary.BigEndian.PutUint16(pcsBytes[25:27], 400)
	
	pcsData, err := NewPresentationData(pcsBytes)
	if err != nil {
		t.Fatalf("Failed to create PCS data: %v", err)
	}
	ds.PCS = Section{Data: pcsData}
	
	// Create WDS with only WinID 0x01
	wdsBytes := make([]byte, 1+9)
	wdsBytes[0] = 1
	wdsBytes[1] = 0x01
	binary.BigEndian.PutUint16(wdsBytes[2:4], 100)
	binary.BigEndian.PutUint16(wdsBytes[4:6], 200)
	binary.BigEndian.PutUint16(wdsBytes[6:8], 300)
	binary.BigEndian.PutUint16(wdsBytes[8:10], 400)
	
	wdsData, err := NewWindowsData(wdsBytes)
	if err != nil {
		t.Fatalf("Failed to create WDS data: %v", err)
	}
	ds.WDS = Section{Data: wdsData}
	
	// Should not return error (just logs warnings for invalid ones)
	err = ds.ValidateWindowCompositionLinkage()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestGetScreenDimensions_WithPCS(t *testing.T) {
	ds := DisplaySet{}
	
	// Create PCS with 4K dimensions
	pcsBytes := make([]byte, 11+8)
	binary.BigEndian.PutUint16(pcsBytes[0:2], 3840) // 4K width
	binary.BigEndian.PutUint16(pcsBytes[2:4], 2160) // 4K height
	pcsBytes[4] = 0x18
	binary.BigEndian.PutUint16(pcsBytes[5:7], 1)
	pcsBytes[7] = 0x80
	pcsBytes[8] = 0x00
	pcsBytes[9] = 0x01
	pcsBytes[10] = 1
	binary.BigEndian.PutUint16(pcsBytes[11:13], 0x1234)
	pcsBytes[13] = 0x01
	pcsBytes[14] = 0x00
	binary.BigEndian.PutUint16(pcsBytes[15:17], 100)
	binary.BigEndian.PutUint16(pcsBytes[17:19], 200)
	
	pcsData, err := NewPresentationData(pcsBytes)
	if err != nil {
		t.Fatalf("Failed to create PCS data: %v", err)
	}
	ds.PCS = Section{Data: pcsData}
	
	width, height := ds.GetScreenDimensions()
	if width != 3840 {
		t.Errorf("Expected width 3840, got %d", width)
	}
	if height != 2160 {
		t.Errorf("Expected height 2160, got %d", height)
	}
}

func TestGetScreenDimensions_WithoutPCS(t *testing.T) {
	ds := DisplaySet{}
	// No PCS data
	
	width, height := ds.GetScreenDimensions()
	if width != DefaultScreenWidth {
		t.Errorf("Expected default width %d, got %d", DefaultScreenWidth, width)
	}
	if height != DefaultScreenHeight {
		t.Errorf("Expected default height %d, got %d", DefaultScreenHeight, height)
	}
}

func TestGetScreenDimensions_InvalidPCSDimensions(t *testing.T) {
	ds := DisplaySet{}
	
	// Create PCS with zero dimensions
	pcsBytes := make([]byte, 11+8)
	binary.BigEndian.PutUint16(pcsBytes[0:2], 0) // Zero width
	binary.BigEndian.PutUint16(pcsBytes[2:4], 0) // Zero height
	pcsBytes[4] = 0x18
	binary.BigEndian.PutUint16(pcsBytes[5:7], 1)
	pcsBytes[7] = 0x80
	pcsBytes[8] = 0x00
	pcsBytes[9] = 0x01
	pcsBytes[10] = 1
	binary.BigEndian.PutUint16(pcsBytes[11:13], 0x1234)
	pcsBytes[13] = 0x01
	pcsBytes[14] = 0x00
	binary.BigEndian.PutUint16(pcsBytes[15:17], 100)
	binary.BigEndian.PutUint16(pcsBytes[17:19], 200)
	
	pcsData, err := NewPresentationData(pcsBytes)
	if err != nil {
		t.Fatalf("Failed to create PCS data: %v", err)
	}
	ds.PCS = Section{Data: pcsData}
	
	// Should return defaults when dimensions are invalid
	width, height := ds.GetScreenDimensions()
	if width != DefaultScreenWidth {
		t.Errorf("Expected default width %d for invalid PCS, got %d", DefaultScreenWidth, width)
	}
	if height != DefaultScreenHeight {
		t.Errorf("Expected default height %d for invalid PCS, got %d", DefaultScreenHeight, height)
	}
}

func TestGetScreenDimensions_InvalidPCSDataType(t *testing.T) {
	ds := DisplaySet{}
	// PCS with wrong data type
	ds.PCS = Section{Data: "invalid"}
	
	width, height := ds.GetScreenDimensions()
	if width != DefaultScreenWidth {
		t.Errorf("Expected default width %d for invalid PCS type, got %d", DefaultScreenWidth, width)
	}
	if height != DefaultScreenHeight {
		t.Errorf("Expected default height %d for invalid PCS type, got %d", DefaultScreenHeight, height)
	}
}

func TestGetActiveCompositionObjects_WithPCS(t *testing.T) {
	ds := DisplaySet{}
	
	// Create PCS with 2 composition objects
	pcsBytes := make([]byte, 11+8+8)
	binary.BigEndian.PutUint16(pcsBytes[0:2], 1920)
	binary.BigEndian.PutUint16(pcsBytes[2:4], 1080)
	pcsBytes[4] = 0x18
	binary.BigEndian.PutUint16(pcsBytes[5:7], 1)
	pcsBytes[7] = 0x80
	pcsBytes[8] = 0x00
	pcsBytes[9] = 0x01
	pcsBytes[10] = 2
	// Object 1
	binary.BigEndian.PutUint16(pcsBytes[11:13], 0x1234)
	pcsBytes[13] = 0x01
	pcsBytes[14] = 0x00
	binary.BigEndian.PutUint16(pcsBytes[15:17], 100)
	binary.BigEndian.PutUint16(pcsBytes[17:19], 200)
	// Object 2
	binary.BigEndian.PutUint16(pcsBytes[19:21], 0x5678)
	pcsBytes[21] = 0x02
	pcsBytes[22] = 0x00
	binary.BigEndian.PutUint16(pcsBytes[23:25], 300)
	binary.BigEndian.PutUint16(pcsBytes[25:27], 400)
	
	pcsData, err := NewPresentationData(pcsBytes)
	if err != nil {
		t.Fatalf("Failed to create PCS data: %v", err)
	}
	ds.PCS = Section{Data: pcsData}
	
	comps := ds.GetActiveCompositionObjects()
	if len(comps) != 2 {
		t.Errorf("Expected 2 composition objects, got %d", len(comps))
	}
	if comps[0].ObjID != 0x1234 {
		t.Errorf("Expected first ObjID 0x1234, got 0x%04X", comps[0].ObjID)
	}
	if comps[1].ObjID != 0x5678 {
		t.Errorf("Expected second ObjID 0x5678, got 0x%04X", comps[1].ObjID)
	}
}

func TestGetActiveCompositionObjects_WithoutPCS(t *testing.T) {
	ds := DisplaySet{}
	// No PCS data
	
	comps := ds.GetActiveCompositionObjects()
	if len(comps) != 0 {
		t.Errorf("Expected 0 composition objects, got %d", len(comps))
	}
}

func TestGetActiveCompositionObjects_EmptyComps(t *testing.T) {
	ds := DisplaySet{}
	
	// Create PCS with no composition objects
	pcsBytes := make([]byte, 11)
	binary.BigEndian.PutUint16(pcsBytes[0:2], 1920)
	binary.BigEndian.PutUint16(pcsBytes[2:4], 1080)
	pcsBytes[4] = 0x18
	binary.BigEndian.PutUint16(pcsBytes[5:7], 1)
	pcsBytes[7] = 0x80
	pcsBytes[8] = 0x00
	pcsBytes[9] = 0x01
	pcsBytes[10] = 0 // NumComps = 0
	
	pcsData, err := NewPresentationData(pcsBytes)
	if err != nil {
		t.Fatalf("Failed to create PCS data: %v", err)
	}
	ds.PCS = Section{Data: pcsData}
	
	comps := ds.GetActiveCompositionObjects()
	if len(comps) != 0 {
		t.Errorf("Expected 0 composition objects, got %d", len(comps))
	}
}

func TestGetActiveCompositionObjects_InvalidPCSDataType(t *testing.T) {
	ds := DisplaySet{}
	// PCS with wrong data type
	ds.PCS = Section{Data: "invalid"}
	
	comps := ds.GetActiveCompositionObjects()
	if len(comps) != 0 {
		t.Errorf("Expected 0 composition objects for invalid PCS type, got %d", len(comps))
	}
}

