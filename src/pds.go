package suptext

import (
    "fmt"
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
    // Read palette id
    id := uint8(bytes[0])

    // Validate buffer length
    if len(bytes) % 5 != 2 {
        return PaletteData{ID: id}, fmt.Errorf("Truncated palette definition")
    }

    // Read palette version
    v := uint8(bytes[1])

    // Calculate number of palettes in buffer
    num_palettes := uint16(len(bytes) / 5)
    pds := PaletteData{ID: id, Version: v, NumPalettes: num_palettes}

    // Read the palettes definitions
    offset := 2
    for i := uint16(0); i < num_palettes; i++ {
        entry_id := uint8(bytes[offset])
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
