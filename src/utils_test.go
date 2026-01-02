package suptext

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"testing"
)

// Helper function to create a section header
func createSectionHeader(pts, dts uint32, segType uint8, size uint16) []byte {
	header := make([]byte, SegmentHeaderSize)
	copy(header[0:2], MagicBytes)
	binary.BigEndian.PutUint32(header[2:6], pts*TimestampAccuracy)
	binary.BigEndian.PutUint32(header[6:10], dts*TimestampAccuracy)
	header[10] = segType
	binary.BigEndian.PutUint16(header[11:13], size)
	return header
}

func TestNewSection_Valid(t *testing.T) {
	header := createSectionHeader(1000, 1000, PCS, 11)
	section, err := NewSection(header)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if section.PTS != 1000 {
		t.Errorf("Expected PTS 1000, got %d", section.PTS)
	}
	if section.Type != PCS {
		t.Errorf("Expected Type PCS (0x%02X), got 0x%02X", PCS, section.Type)
	}
	if section.Size != 11 {
		t.Errorf("Expected Size 11, got %d", section.Size)
	}
}

func TestNewSection_InvalidMagic(t *testing.T) {
	header := make([]byte, SegmentHeaderSize)
	copy(header[0:2], "XX") // Invalid magic
	binary.BigEndian.PutUint32(header[2:6], 1000*TimestampAccuracy)
	binary.BigEndian.PutUint32(header[6:10], 1000*TimestampAccuracy)
	header[10] = PCS
	binary.BigEndian.PutUint16(header[11:13], 11)

	_, err := NewSection(header)
	if err == nil {
		t.Error("Expected error for invalid magic bytes")
	}
}

func TestReadPGS_WithEND(t *testing.T) {
	// Create a simple PGS with PCS and END
	var buf bytes.Buffer
	
	// PCS section
	pcsHeader := createSectionHeader(1000, 1000, PCS, 11)
	buf.Write(pcsHeader)
	pcsData := make([]byte, 11)
	binary.BigEndian.PutUint16(pcsData[0:2], 1920)  // Width
	binary.BigEndian.PutUint16(pcsData[2:4], 1080)  // Height
	pcsData[4] = 0x18                                // Framerate
	binary.BigEndian.PutUint16(pcsData[5:7], 1)     // Num
	pcsData[7] = 0x80                                // State (epoch start)
	pcsData[8] = 0x00                                // PaletteUpdate
	pcsData[9] = 0x01                                // PaletteID
	pcsData[10] = 0                                  // NumComps
	buf.Write(pcsData)
	
	// END section (size 0)
	endHeader := createSectionHeader(2000, 2000, END, 0)
	buf.Write(endHeader)

	reader := bufio.NewReader(&buf)
	pgs, err := ReadPGS(reader)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(pgs.Sections) != 1 {
		t.Errorf("Expected 1 DisplaySet, got %d", len(pgs.Sections))
	}
	if pgs.Sections[0].END.Type != END {
		t.Error("Expected END marker")
	}
}

func TestReadPGS_MissingEND(t *testing.T) {
	// Create PGS without END marker
	var buf bytes.Buffer
	
	// PCS section
	pcsHeader := createSectionHeader(1000, 1000, PCS, 11)
	buf.Write(pcsHeader)
	pcsData := make([]byte, 11)
	binary.BigEndian.PutUint16(pcsData[0:2], 1920)
	binary.BigEndian.PutUint16(pcsData[2:4], 1080)
	pcsData[4] = 0x18
	binary.BigEndian.PutUint16(pcsData[5:7], 1)
	pcsData[7] = 0x80 // Epoch start
	pcsData[8] = 0x00
	pcsData[9] = 0x01
	pcsData[10] = 0
	buf.Write(pcsData)
	// No END marker

	reader := bufio.NewReader(&buf)
	pgs, err := ReadPGS(reader)
	if err != nil {
		t.Fatalf("Expected no error (should handle missing END), got: %v", err)
	}
	if len(pgs.Sections) != 1 {
		t.Errorf("Expected 1 DisplaySet (appended at EOF), got %d", len(pgs.Sections))
	}
}

func TestReadPGS_IncompleteODS(t *testing.T) {
	// Create PGS with incomplete ODS sequence (no END, EOF reached)
	var buf bytes.Buffer
	
	// PCS section
	pcsHeader := createSectionHeader(1000, 1000, PCS, 11)
	buf.Write(pcsHeader)
	pcsData := make([]byte, 11)
	binary.BigEndian.PutUint16(pcsData[0:2], 1920)
	binary.BigEndian.PutUint16(pcsData[2:4], 1080)
	pcsData[4] = 0x18
	binary.BigEndian.PutUint16(pcsData[5:7], 1)
	pcsData[7] = 0x80
	pcsData[8] = 0x00
	pcsData[9] = 0x01
	pcsData[10] = 0
	buf.Write(pcsData)
	
	// ODS first sequence (not ended)
	odsHeader := createSectionHeader(1000, 1000, ODS, 11)
	buf.Write(odsHeader)
	odsData := make([]byte, 11)
	binary.BigEndian.PutUint16(odsData[0:2], 0x1234) // ID
	odsData[2] = 0x01                                 // Version
	odsData[3] = 0x80                                 // First sequence (not ended)
	binary.BigEndian.PutUint16(odsData[4:6], 0x0000)
	odsData[6] = 0x0A                                 // Length
	binary.BigEndian.PutUint16(odsData[7:9], 100)    // Width
	binary.BigEndian.PutUint16(odsData[9:11], 200)   // Height
	buf.Write(odsData)
	// No END marker, EOF reached

	reader := bufio.NewReader(&buf)
	pgs, err := ReadPGS(reader)
	if err != nil {
		t.Fatalf("Expected no error (should merge incomplete ODS at EOF), got: %v", err)
	}
	if len(pgs.Sections) != 1 {
		t.Errorf("Expected 1 DisplaySet, got %d", len(pgs.Sections))
	}
	if len(pgs.Sections[0].ODS) == 0 {
		t.Error("Expected incomplete ODS to be merged into DisplaySet")
	}
}

func TestReadPGS_AcquisitionPoint(t *testing.T) {
	// Test acquisition point (State = 0x40)
	var buf bytes.Buffer
	
	// PCS with acquisition point
	pcsHeader := createSectionHeader(1000, 1000, PCS, 11)
	buf.Write(pcsHeader)
	pcsData := make([]byte, 11)
	binary.BigEndian.PutUint16(pcsData[0:2], 1920)
	binary.BigEndian.PutUint16(pcsData[2:4], 1080)
	pcsData[4] = 0x18
	binary.BigEndian.PutUint16(pcsData[5:7], 1)
	pcsData[7] = 0x40 // Acquisition point (not epoch start 0x80)
	pcsData[8] = 0x00
	pcsData[9] = 0x01
	pcsData[10] = 0
	buf.Write(pcsData)
	
	// END
	endHeader := createSectionHeader(2000, 2000, END, 0)
	buf.Write(endHeader)

	reader := bufio.NewReader(&buf)
	pgs, err := ReadPGS(reader)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(pgs.Sections) != 1 {
		t.Errorf("Expected 1 DisplaySet, got %d", len(pgs.Sections))
	}
	
	// Check that IsEpochStart recognizes 0x40
	ds := pgs.Sections[0]
	if !ds.IsEpochStart() {
		t.Error("Expected IsEpochStart to return true for State 0x40 (acquisition point)")
	}
}

func TestReadPGS_EpochStart(t *testing.T) {
	// Test epoch start (State = 0x80)
	var buf bytes.Buffer
	
	// PCS with epoch start
	pcsHeader := createSectionHeader(1000, 1000, PCS, 11)
	buf.Write(pcsHeader)
	pcsData := make([]byte, 11)
	binary.BigEndian.PutUint16(pcsData[0:2], 1920)
	binary.BigEndian.PutUint16(pcsData[2:4], 1080)
	pcsData[4] = 0x18
	binary.BigEndian.PutUint16(pcsData[5:7], 1)
	pcsData[7] = 0x80 // Epoch start
	pcsData[8] = 0x00
	pcsData[9] = 0x01
	pcsData[10] = 0
	buf.Write(pcsData)
	
	// END
	endHeader := createSectionHeader(2000, 2000, END, 0)
	buf.Write(endHeader)
	
	reader := bufio.NewReader(&buf)
	pgs, err := ReadPGS(reader)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	ds := pgs.Sections[0]
	if !ds.IsEpochStart() {
		t.Error("Expected IsEpochStart to return true for State 0x80")
	}
}

func TestReadPGS_MultipleODSSequences(t *testing.T) {
	// Test ODS with multiple sequences
	var buf bytes.Buffer
	
	// PCS
	pcsHeader := createSectionHeader(1000, 1000, PCS, 11)
	buf.Write(pcsHeader)
	pcsData := make([]byte, 11)
	binary.BigEndian.PutUint16(pcsData[0:2], 1920)
	binary.BigEndian.PutUint16(pcsData[2:4], 1080)
	pcsData[4] = 0x18
	binary.BigEndian.PutUint16(pcsData[5:7], 1)
	pcsData[7] = 0x80
	pcsData[8] = 0x00
	pcsData[9] = 0x01
	pcsData[10] = 0
	buf.Write(pcsData)
	
	// ODS first sequence (not ended)
	ods1Header := createSectionHeader(1000, 1000, ODS, 11)
	buf.Write(ods1Header)
	ods1Data := make([]byte, 11)
	binary.BigEndian.PutUint16(ods1Data[0:2], 0x1234)
	ods1Data[2] = 0x01
	ods1Data[3] = 0x80 // First, not ended
	binary.BigEndian.PutUint16(ods1Data[4:6], 0x0000)
	ods1Data[6] = 0x0A
	binary.BigEndian.PutUint16(ods1Data[7:9], 100)
	binary.BigEndian.PutUint16(ods1Data[9:11], 200)
	buf.Write(ods1Data)
	
	// ODS continuation (ended)
	ods2Header := createSectionHeader(1000, 1000, ODS, 5)
	buf.Write(ods2Header)
	ods2Data := make([]byte, 5)
	binary.BigEndian.PutUint16(ods2Data[0:2], 0x1234)
	ods2Data[2] = 0x01
	ods2Data[3] = 0x40 // Last sequence
	ods2Data[4] = 0x01 // Data
	buf.Write(ods2Data)
	
	// END
	endHeader := createSectionHeader(2000, 2000, END, 0)
	buf.Write(endHeader)

	reader := bufio.NewReader(&buf)
	pgs, err := ReadPGS(reader)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(pgs.Sections[0].ODS) != 1 {
		t.Errorf("Expected 1 ODS (merged), got %d", len(pgs.Sections[0].ODS))
	}
}

func TestFormatMilliseconds(t *testing.T) {
	// Test timestamp formatting
	ts := uint32(3661000) // 1 hour, 1 minute, 1 second
	formatted := FormatMilliseconds(ts)
	expected := "01:01:01,000"
	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}
	
	ts2 := uint32(125000) // 2 minutes, 5 seconds
	formatted2 := FormatMilliseconds(ts2)
	expected2 := "00:02:05,000"
	if formatted2 != expected2 {
		t.Errorf("Expected %s, got %s", expected2, formatted2)
	}
}

