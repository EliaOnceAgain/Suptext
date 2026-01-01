package suptext

import (
    "encoding/binary"
    "fmt"
    "log"
)

const SequenceShortHeaderSize = 4       // ID, version, sequence
const FirstSequenceHeaderSize = 7       // ID, version, sequence, length

type ObjectData struct {
    ID uint16
    Version uint8
    Sequence uint8
    Length uint32 // uint24
    BytesRead uint32 // uint24
    Width uint16
    Height uint16
    Ended bool
    Data []byte `json:"-"`
}

func (o *ObjectData) IsFirstSequence() bool {
    return o.Sequence & 0x80 != 0
}

func (o *ObjectData) IsLastSequence() bool {
    return o.Sequence & 0x40 != 0
}

func NewObjectData(bytes []byte) (ObjectData, error) {
    // Validate buffer length
    if len(bytes) < SequenceShortHeaderSize {
        log.Printf("Warning: Truncated object definition header (got %d bytes, need %d)", len(bytes), SequenceShortHeaderSize)
        return ObjectData{}, fmt.Errorf("Truncated object definition")
    }

    // Read basic object fields
    id := binary.BigEndian.Uint16(bytes[:2])
    v := uint8(bytes[2])
    sequence := uint8(bytes[3])

    obj := ObjectData{
        ID: id,
        Version: v,
        Sequence: sequence,
    }
    obj.Ended = obj.IsLastSequence()

    // Handle ODS Start Sequence
    if obj.IsFirstSequence() {
        if len(bytes) < FirstSequenceHeaderSize {
            log.Printf("Warning: Truncated first sequence header for ODS ID %d (got %d bytes, need %d)", id, len(bytes), FirstSequenceHeaderSize)
            return ObjectData{}, fmt.Errorf("Truncated first sequence header")
        }
        obj.Length = uint32(bytes[4])<<16 | uint32(bytes[5])<<8 | uint32(bytes[6])
        
        // Validate Width/Height before reading
        if len(bytes) < 11 {
            log.Printf("Warning: Truncated first sequence - missing Width/Height for ODS ID %d", id)
            return ObjectData{}, fmt.Errorf("Truncated first sequence - missing Width/Height")
        }
        obj.Width = binary.BigEndian.Uint16(bytes[7:9])
        obj.Height = binary.BigEndian.Uint16(bytes[9:11])
        
        // Validate Width/Height values
        if obj.Width == 0 || obj.Height == 0 {
            log.Printf("Warning: Invalid Width/Height for ODS ID %d (Width=%d, Height=%d)", id, obj.Width, obj.Height)
        }
        if obj.Width > 1920 || obj.Height > 1080 {
            log.Printf("Warning: Unusually large dimensions for ODS ID %d (Width=%d, Height=%d)", id, obj.Width, obj.Height)
        }
        
        obj.Data = bytes[11:]
        obj.BytesRead += uint32(len(bytes[FirstSequenceHeaderSize:]))
    } else {
        obj.Data = bytes[SequenceShortHeaderSize:]
        obj.BytesRead += uint32(len(bytes[SequenceShortHeaderSize:]))
    }

    return obj, nil
}

func (o *ObjectData) MergeSequence(other ObjectData) error {
    if o.Ended {
        log.Printf("Warning: Attempted to merge sequence to already ended ODS ID %d", o.ID)
        return fmt.Errorf("Failed merging sequence - sequence already ended")
    }
    if other.IsFirstSequence() {
        log.Printf("Warning: Attempted to merge first sequence to ODS ID %d", o.ID)
        return fmt.Errorf("Failed merging sequence - can't merge with another sequence start")
    }
    o.Data = append(o.Data, other.Data...)
    o.BytesRead += other.BytesRead
    o.Ended = other.Ended
    if o.Ended && o.BytesRead != o.Length {
        log.Printf("Warning: ODS ID %d ended but read %d bytes, expected %d bytes", o.ID, o.BytesRead, o.Length)
        // Don't return error, allow incomplete sequences to be processed
    }
    return nil
}
