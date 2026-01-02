package suptext

import (
	"encoding/binary"
	"testing"
)

func TestNewObjectData_FirstSequence(t *testing.T) {
	// Valid first sequence
	bytes := make([]byte, 20)
	binary.BigEndian.PutUint16(bytes[0:2], 0x1234) // ID
	bytes[2] = 0x01                                  // Version
	bytes[3] = 0x80                                  // First sequence flag
	binary.BigEndian.PutUint16(bytes[4:6], 0x0000)  // Length high
	bytes[6] = 0x0A                                  // Length low (10 bytes)
	binary.BigEndian.PutUint16(bytes[7:9], 100)     // Width
	binary.BigEndian.PutUint16(bytes[9:11], 200)    // Height
	bytes[11] = 0x01                                 // Data
	bytes[12] = 0x02                                 // Data

	obj, err := NewObjectData(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if obj.ID != 0x1234 {
		t.Errorf("Expected ID 0x1234, got 0x%04X", obj.ID)
	}
	if obj.Version != 0x01 {
		t.Errorf("Expected Version 0x01, got 0x%02X", obj.Version)
	}
	if !obj.IsFirstSequence() {
		t.Error("Expected IsFirstSequence to be true")
	}
	if obj.Width != 100 {
		t.Errorf("Expected Width 100, got %d", obj.Width)
	}
	if obj.Height != 200 {
		t.Errorf("Expected Height 200, got %d", obj.Height)
	}
}

func TestNewObjectData_ContinuationSequence(t *testing.T) {
	// Valid continuation sequence
	bytes := make([]byte, 10)
	binary.BigEndian.PutUint16(bytes[0:2], 0x1234) // ID
	bytes[2] = 0x01                                  // Version
	bytes[3] = 0x00                                // Continuation (no flags)
	bytes[4] = 0x01                                // Data
	bytes[5] = 0x02                                // Data

	obj, err := NewObjectData(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if obj.ID != 0x1234 {
		t.Errorf("Expected ID 0x1234, got 0x%04X", obj.ID)
	}
	if obj.IsFirstSequence() {
		t.Error("Expected IsFirstSequence to be false")
	}
	if len(obj.Data) != 6 {
		t.Errorf("Expected Data length 6, got %d", len(obj.Data))
	}
}

func TestNewObjectData_LastSequence(t *testing.T) {
	// Last sequence flag
	bytes := make([]byte, 10)
	binary.BigEndian.PutUint16(bytes[0:2], 0x1234) // ID
	bytes[2] = 0x01                                  // Version
	bytes[3] = 0x40                                  // Last sequence flag
	bytes[4] = 0x01                                  // Data

	obj, err := NewObjectData(bytes)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !obj.Ended {
		t.Error("Expected Ended to be true")
	}
}

func TestNewObjectData_TruncatedHeader(t *testing.T) {
	// Too short for even basic header
	bytes := make([]byte, 2)
	_, err := NewObjectData(bytes)
	if err == nil {
		t.Error("Expected error for truncated header")
	}
}

func TestNewObjectData_TruncatedFirstSequence(t *testing.T) {
	// Too short for first sequence
	bytes := make([]byte, 6)
	binary.BigEndian.PutUint16(bytes[0:2], 0x1234) // ID
	bytes[2] = 0x01                                  // Version
	bytes[3] = 0x80                                  // First sequence flag
	bytes[4] = 0x00                                  // Partial length

	_, err := NewObjectData(bytes)
	if err == nil {
		t.Error("Expected error for truncated first sequence")
	}
}

func TestNewObjectData_TruncatedWidthHeight(t *testing.T) {
	// First sequence but missing Width/Height
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint16(bytes[0:2], 0x1234) // ID
	bytes[2] = 0x01                                  // Version
	bytes[3] = 0x80                                  // First sequence flag
	bytes[4] = 0x00                                  // Length
	bytes[5] = 0x00
	bytes[6] = 0x0A

	_, err := NewObjectData(bytes)
	if err == nil {
		t.Error("Expected error for missing Width/Height")
	}
}

func TestObjectData_MergeSequence(t *testing.T) {
	// Create first sequence
	firstBytes := make([]byte, 15)
	binary.BigEndian.PutUint16(firstBytes[0:2], 0x1234)
	firstBytes[2] = 0x01
	firstBytes[3] = 0x80 // First sequence
	binary.BigEndian.PutUint16(firstBytes[4:6], 0x0000)
	firstBytes[6] = 0x0A // Length
	binary.BigEndian.PutUint16(firstBytes[7:9], 100)
	binary.BigEndian.PutUint16(firstBytes[9:11], 200)
	firstBytes[11] = 0x01

	first, err := NewObjectData(firstBytes)
	if err != nil {
		t.Fatalf("Failed to create first sequence: %v", err)
	}

	// Create continuation sequence
	contBytes := make([]byte, 6)
	binary.BigEndian.PutUint16(contBytes[0:2], 0x1234)
	contBytes[2] = 0x01
	contBytes[3] = 0x00 // Continuation
	contBytes[4] = 0x02

	cont, err := NewObjectData(contBytes)
	if err != nil {
		t.Fatalf("Failed to create continuation: %v", err)
	}

	// Merge
	err = first.MergeSequence(cont)
	if err != nil {
		t.Fatalf("Failed to merge: %v", err)
	}
	// First sequence: data starts at byte 11 (after 11-byte header: ID, version, sequence, length(3), width(2), height(2))
	// Continuation: data starts at byte 4 (after 4-byte header: ID, version, sequence)
	// Merged should have: (15 - 11) + (6 - 4) = 4 + 2 = 6 bytes
	firstDataLen := len(firstBytes) - 11
	contDataLen := len(contBytes) - SequenceShortHeaderSize
	expectedLen := firstDataLen + contDataLen
	if len(first.Data) != expectedLen {
		t.Errorf("Expected merged data length %d, got %d (first: %d bytes, cont: %d bytes)", expectedLen, len(first.Data), firstDataLen, contDataLen)
	}
	// Check that the data is correctly merged
	if len(first.Data) > 0 && first.Data[0] != 0x01 {
		t.Errorf("Expected first byte 0x01, got 0x%02X", first.Data[0])
	}
	if len(first.Data) >= 2 && first.Data[len(first.Data)-2] != 0x02 {
		t.Errorf("Expected second-to-last data byte 0x02, got 0x%02X", first.Data[len(first.Data)-2])
	}
}

func TestObjectData_MergeSequence_AlreadyEnded(t *testing.T) {
	first := ObjectData{ID: 0x1234, Ended: true}
	other := ObjectData{ID: 0x1234}

	err := first.MergeSequence(other)
	if err == nil {
		t.Error("Expected error when merging to ended sequence")
	}
}

func TestObjectData_MergeSequence_FirstSequence(t *testing.T) {
	first := ObjectData{ID: 0x1234, Ended: false}
	other := ObjectData{ID: 0x1234, Sequence: 0x80} // First sequence flag

	err := first.MergeSequence(other)
	if err == nil {
		t.Error("Expected error when merging first sequence")
	}
}

