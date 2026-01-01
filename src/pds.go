package suptext

import (
    "fmt"
    "log"
)

const PaletteSize = 5

type PaletteDefinition struct {
    Y uint8 // Luminance
    Cr uint8 // ColorDiffRed
    Cb uint8 // ColorDiffBlue
    A uint8 // Transparency
}

type PaletteData struct {
    ID uint8
    Version uint8
    NumPalettes uint16
    Palettes [256]PaletteDefinition `json:"-"`
}

func NewPaletteData(bytes []byte) (PaletteData, error) {
    // Validate minimum buffer length (header: ID + Version)
    if len(bytes) < 2 {
        return PaletteData{}, fmt.Errorf("Truncated palette definition - missing header")
    }

    // Read palette id
    id := uint8(bytes[0])

    // Read palette version
    v := uint8(bytes[1])

    // Validate buffer length (header + palette entries must be multiple of PaletteSize)
    // Header is 2 bytes, each palette entry is 5 bytes
    data_length := len(bytes) - 2
    if data_length%PaletteSize != 0 {
        log.Printf("Warning: Truncated palette definition for palette ID %d (buffer length: %d, expected multiple of %d + 2)", id, len(bytes), PaletteSize)
        return PaletteData{ID: id, Version: v}, fmt.Errorf("Truncated palette definition")
    }

    // Calculate number of palettes in buffer (corrected: subtract header)
    num_palettes := uint16(data_length / PaletteSize)
    pds := PaletteData{ID: id, Version: v, NumPalettes: num_palettes}

    // Read the palettes definitions
    offset := 2
    for i := uint16(0); i < num_palettes; i++ {
        if offset+PaletteSize > len(bytes) {
            log.Printf("Warning: Truncated palette entry %d for palette ID %d", i, id)
            break
        }
        entry_id := uint8(bytes[offset])
        
        // Validate palette ID is in range 0-255
        if entry_id > 255 {
            log.Printf("Warning: Invalid palette entry ID %d (must be 0-255) for palette ID %d", entry_id, id)
            offset += PaletteSize
            continue
        }
        
        pds.Palettes[entry_id] = PaletteDefinition{
            Y: uint8(bytes[offset + 1]),
            Cr: uint8(bytes[offset + 2]),
            Cb: uint8(bytes[offset + 3]),
            A: uint8(bytes[offset + 4]),
        }
        offset += PaletteSize
    }

    return pds, nil
}
