package suptext

import (
    "encoding/binary"
    "fmt"
    "log"
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

// Default screen bounds (HD 1920x1080)
const DefaultScreenWidth = 1920
const DefaultScreenHeight = 1080

func NewWindowsData(bytes []byte) (WindowsData, error) {
    return NewWindowsDataWithBounds(bytes, DefaultScreenWidth, DefaultScreenHeight)
}

func NewWindowsDataWithBounds(bytes []byte, screenWidth, screenHeight uint16) (WindowsData, error) {
    // Validate minimum buffer length
    if len(bytes) < 1 {
        return WindowsData{}, fmt.Errorf("Truncated windows definition - missing header")
    }

    // Read number of windows
    NumWindows := uint8(bytes[0])
    wds := WindowsData{NumWindows: NumWindows}

    // Validate buffer length
    required_size := 1 + (NumWindows * WindowDefinitionSize)
    if len(bytes) < int(required_size) {
        log.Printf("Warning: Truncated windows definition (got %d bytes, need %d)", len(bytes), required_size)
        return wds, fmt.Errorf("Truncated windows definition")
    }

    // Read the windows definitions
    offset := 1
    for i := uint8(1); i <= NumWindows; i++ {
        if offset+WindowDefinitionSize > len(bytes) {
            log.Printf("Warning: Truncated window definition %d", i)
            break
        }
        
        window := WindowDefinition {
            WinID: uint8(bytes[offset]),
            Hpos: binary.BigEndian.Uint16(bytes[offset + 1: offset + 3]),
            Vpos: binary.BigEndian.Uint16(bytes[offset + 3: offset + 5]),
            Width: binary.BigEndian.Uint16(bytes[offset + 5: offset + 7]),
            Height: binary.BigEndian.Uint16(bytes[offset + 7: offset + 9]),
        }
        
        // Validate window bounds using provided screen dimensions
        if err := ValidateWindowBounds(window, screenWidth, screenHeight); err != nil {
            log.Printf("Warning: Window ID %d exceeds screen bounds: %v", window.WinID, err)
            // Continue processing instead of failing
        }
        
        wds.Windows = append(wds.Windows, window)
        offset += WindowDefinitionSize
    }

    return wds, nil
}

// ValidateWindowBounds checks if a window exceeds screen bounds
func ValidateWindowBounds(window WindowDefinition, screenWidth, screenHeight uint16) error {
    // Check if window position is within bounds
    if window.Hpos >= screenWidth {
        return fmt.Errorf("Hpos %d exceeds screen width %d", window.Hpos, screenWidth)
    }
    if window.Vpos >= screenHeight {
        return fmt.Errorf("Vpos %d exceeds screen height %d", window.Vpos, screenHeight)
    }
    
    // Check if window extends beyond screen bounds
    if window.Hpos+window.Width > screenWidth {
        return fmt.Errorf("Window extends beyond screen width (Hpos %d + Width %d > %d)", window.Hpos, window.Width, screenWidth)
    }
    if window.Vpos+window.Height > screenHeight {
        return fmt.Errorf("Window extends beyond screen height (Vpos %d + Height %d > %d)", window.Vpos, window.Height, screenHeight)
    }
    
    // Check for zero dimensions
    if window.Width == 0 || window.Height == 0 {
        return fmt.Errorf("Window has zero dimensions (Width=%d, Height=%d)", window.Width, window.Height)
    }
    
    return nil
}
