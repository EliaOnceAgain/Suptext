package suptext

import (
    "encoding/binary"
    "fmt"
)

const WindowDefinitionSize = 9

type WindowDefinition struct {
    WinID uint8
    Hpos uint16
    Vpos uint16
    Width uint16
    Height uint16
}

type WindowsData struct {
    NumWindows uint8
    Windows []WindowDefinition
}

func NewWindowsData(bytes []byte) (WindowsData, error) {
    // Read number of windows
    NumWindows := uint8(bytes[0])
    wds := WindowsData{NumWindows: NumWindows}

    // Validate buffer length
    if NumWindows * WindowDefinitionSize > uint8(len(bytes) - 1) {
        return wds, fmt.Errorf("Truncated windows definition")
    }

    // Read the windows definitions
    offset := 1
    for i := uint8(1); i <= NumWindows; i++ {
        window := WindowDefinition {
            WinID: uint8(bytes[offset]),
            Hpos: binary.BigEndian.Uint16(bytes[offset + 1: offset + 3]),
            Vpos: binary.BigEndian.Uint16(bytes[offset + 3: offset + 5]),
            Width: binary.BigEndian.Uint16(bytes[offset + 5: offset + 7]),
            Height: binary.BigEndian.Uint16(bytes[offset + 7: offset + 9]),
        }
        wds.Windows = append(wds.Windows, window)
        offset += WindowDefinitionSize
    }

    return wds, nil
}
