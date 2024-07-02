package suptext

import (
    "encoding/binary"
    "fmt"
)

const ObjectMinSize = 7

type ObjectData struct {
    ID uint16
    Version uint8
    Last uint8
    Length uint32 // uint24
    Width uint16
    Height uint16
    Data []byte `json:"-"`
}

func NewObjectData(bytes []byte) (ObjectData, error) {
    // Validate buffer length
    if len(bytes) < ObjectMinSize {
        return ObjectData{}, fmt.Errorf("Truncated object definition")
    }

    // Read basic object fields
    id := binary.BigEndian.Uint16(bytes[:2])
    v := uint8(bytes[2])
    last := uint8(bytes[3])
    length := uint32(bytes[4])<<16 | uint32(bytes[5])<<8 | uint32(bytes[6])
    obj := ObjectData{
        ID: id,
        Version: v,
        Last: last,
        Length: length,
    }

    // Validate buffer length for object data
    if uint32(len(bytes[ObjectMinSize:])) != length {
        return obj, fmt.Errorf("Truncated object payload %d != %d", uint32(len(bytes[ObjectMinSize:])), length)
    }

    // Set image fields
    obj.Width = binary.BigEndian.Uint16(bytes[7:9])
    obj.Height = binary.BigEndian.Uint16(bytes[9:11])
    obj.Data = bytes[11:]

    return obj, nil
}
