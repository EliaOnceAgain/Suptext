package suptext

import (
    "encoding/binary"
    "fmt"
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
        obj.Length = uint32(bytes[4])<<16 | uint32(bytes[5])<<8 | uint32(bytes[6])
        obj.Width = binary.BigEndian.Uint16(bytes[7:9])
        obj.Height = binary.BigEndian.Uint16(bytes[9:11])
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
        return fmt.Errorf("Failed merging sequence - sequence already ended")
    }
    if other.IsFirstSequence() {
        return fmt.Errorf("Failed merging sequence - can't merge with another sequence start")
    }
    o.Data = append(o.Data, other.Data...)
    o.BytesRead += other.BytesRead
    o.Ended = other.Ended
    if o.Ended && o.BytesRead != o.Length {
        return fmt.Errorf("Failed merging sequence - ended section but read bytes not equal expected length")
    }
    return nil
}
