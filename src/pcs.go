package suptext

import (
    "encoding/binary"
    "fmt"
    "log"
)

const CompositionObjectSize = 8
const CompositionObjectExtendedSize = 16
const PresentationCompositionSize = 11

type PresentationCompositionData struct {
    Width uint16
    Height uint16
    Framerate uint8
    Num uint16
    State uint8
    PaletteUpdate uint8
    PaletteID uint8
    NumComps uint8
    Comps []CompositionObject
}

type CompositionObject struct {
    ObjID uint16
    WinID uint8
    Cropped uint8
    Hpos uint16
    Vpos uint16
    HCropPos uint16
    VCropPos uint16
    CropWidth uint16
    CropHeight uint16
}

func NewPresentationData(bytes []byte) (PresentationCompositionData, error) {
    // Validate minimum buffer length
    if len(bytes) < PresentationCompositionSize {
        log.Printf("Warning: Truncated presentation composition header (need %d bytes, got %d)", PresentationCompositionSize, len(bytes))
        return PresentationCompositionData{}, fmt.Errorf("Truncated presentation composition header")
    }

    section := PresentationCompositionData{
        Width:          binary.BigEndian.Uint16(bytes[:2]),
        Height:         binary.BigEndian.Uint16(bytes[2:4]),
        Framerate:      uint8(bytes[4]),
        Num:            binary.BigEndian.Uint16(bytes[5:7]),
        State:          uint8(bytes[7]),
        PaletteUpdate:  uint8(bytes[8]),
        PaletteID:      uint8(bytes[9]),
        NumComps:       uint8(bytes[10]),
    }

    // Skip section data header
    offset := PresentationCompositionSize

    // Read composition objects
    for i := uint8(1); i <= section.NumComps; i++ {
        // Check if we have enough bytes remaining
        if offset >= len(bytes) {
            log.Printf("Warning: Truncated composition objects list (expected %d objects, only read %d before buffer end)", section.NumComps, i-1)
            break
        }
        
        comp, err := NewCompositionObject(bytes[offset:])
        if err != nil {
            log.Printf("Warning: Failed to parse composition object %d: %v", i, err)
            // Continue processing remaining objects instead of failing completely
            break
        }
        section.Comps = append(section.Comps, comp)
        // Advance to the beginning of next object
        if 0 == comp.Cropped {
            offset += CompositionObjectSize
        } else {
            offset += CompositionObjectExtendedSize
        }
    }

    return section, nil
}

func NewCompositionObject(bytes []byte) (CompositionObject, error) {
    // Not enough bytes
    if len(bytes) < 8 {
        log.Printf("Warning: Truncated composition object - need at least 8 bytes, got %d", len(bytes))
        return CompositionObject{}, fmt.Errorf("Truncated composition object")
    }
    // Read composition fields
    composition := CompositionObject{
        ObjID:      binary.BigEndian.Uint16(bytes[:2]),
        WinID:      uint8(bytes[2]),
        Cropped:    uint8(bytes[3]),
        Hpos:       binary.BigEndian.Uint16(bytes[4:6]),
        Vpos:       binary.BigEndian.Uint16(bytes[6:8]),
    }
    // Not cropped so no extension
    if 0 == composition.Cropped {
        return composition, nil
    }
    // Not enough bytes for extension
    if len(bytes) < 16 {
        log.Printf("Warning: Truncated composition object extension for ObjID %d (need 16 bytes, got %d). Extension fields will be zero.", composition.ObjID, len(bytes))
        return composition, nil
    }
    // Read extension fields (offset 8 for extension data)
    composition.HCropPos    = binary.BigEndian.Uint16(bytes[8:10])
    composition.VCropPos    = binary.BigEndian.Uint16(bytes[10:12])
    composition.CropWidth   = binary.BigEndian.Uint16(bytes[12:14])
    composition.CropHeight  = binary.BigEndian.Uint16(bytes[14:16])
    return composition, nil
}
