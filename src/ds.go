package suptext

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
	"github.com/otiai10/gosseract/v2"
)

type DisplaySet struct {
    PCS Section
    WDS Section
    PDS Section
    ODS []Section
    END Section
}

func (d *DisplaySet) StartTS() string {
    return FormatMilliseconds(d.PCS.PTS)
}

func (d *DisplaySet) IsEpochStart() bool {
    if d.PCS.Data == nil {
        return false
    }
    pcsData, ok := d.PCS.Data.(PresentationCompositionData)
    if !ok {
        return false
    }
    state := pcsData.State
    return state == 0x40 || state == 0x80
}

func (d *DisplaySet) IsEpochEnd() bool {
    return d.PCS.Data.(PresentationCompositionData).State == 0
}

func (d *DisplaySet) GetActiveCompositionObjects() []CompositionObject {
    if d.PCS.Data == nil {
        return []CompositionObject{}
    }
    pcsData, ok := d.PCS.Data.(PresentationCompositionData)
    if !ok {
        return []CompositionObject{}
    }
    return pcsData.Comps
}

// ValidateWindowCompositionLinkage checks that all composition objects reference valid windows
func (d *DisplaySet) ValidateWindowCompositionLinkage() error {
    if d.PCS.Data == nil || d.WDS.Data == nil {
        return nil // Can't validate without both PCS and WDS
    }
    
    pcsData, ok := d.PCS.Data.(PresentationCompositionData)
    if !ok {
        return nil
    }
    
    wdsData, ok := d.WDS.Data.(WindowsData)
    if !ok {
        return nil
    }
    
    // Build a map of valid window IDs
    validWinIDs := make(map[uint8]bool)
    for _, window := range wdsData.Windows {
        validWinIDs[window.WinID] = true
    }
    
    // Check each composition object references a valid window
    for _, comp := range pcsData.Comps {
        if !validWinIDs[comp.WinID] {
            log.Printf("Warning: Composition object ObjID %d references invalid WinID %d", comp.ObjID, comp.WinID)
        }
    }
    
    return nil
}

// GetScreenDimensions returns screen dimensions from PCS, or defaults
func (d *DisplaySet) GetScreenDimensions() (uint16, uint16) {
    if d.PCS.Data == nil {
        return DefaultScreenWidth, DefaultScreenHeight
    }
    
    pcsData, ok := d.PCS.Data.(PresentationCompositionData)
    if !ok {
        return DefaultScreenWidth, DefaultScreenHeight
    }
    
    // Use PCS dimensions if valid, otherwise defaults
    if pcsData.Width > 0 && pcsData.Height > 0 {
        return pcsData.Width, pcsData.Height
    }
    
    return DefaultScreenWidth, DefaultScreenHeight
}

func (d *DisplaySet) AppendSRT(ocr *gosseract.Client, f *os.File, i uint, ets string) error {
    text, err := d.OCR(ocr)
    if err != nil {
        return err
    }
    sts := d.StartTS()
    srt := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", i, sts, ets, text)
    if _, err := f.WriteString(srt); err != nil {
        log.Fatalf("Failed to write to file: %v", err)
    }
    return nil
}

func (d *DisplaySet) OCR(ocr *gosseract.Client) (string, error) {
    var text, ocr_result string

    // Get active composition objects to filter ODS processing
    activeComps := d.GetActiveCompositionObjects()
    activeObjIDs := make(map[uint16]bool)
    for _, comp := range activeComps {
        activeObjIDs[comp.ObjID] = true
    }

    // If no active composition objects, log warning but continue processing all ODS
    if len(activeComps) == 0 {
        log.Printf("Warning: No active composition objects in DisplaySet at PTS %d - processing all ODS", d.PCS.PTS)
    }

    for _, ods := range d.ODS {
        objData, ok := ods.Data.(ObjectData)
        if !ok {
            log.Printf("Warning: Invalid ODS data type in DisplaySet")
            continue
        }
        if !objData.Ended {
            log.Printf("Warning: Skipping incomplete ODS sequence - ObjID %d at PTS %d (sequence not ended)", objData.ID, ods.PTS)
            continue
        }
        
        // Only process ODS that are referenced by active composition objects
        // This avoids processing ODS for windows that aren't displayed
        if len(activeComps) > 0 && !activeObjIDs[objData.ID] {
            log.Printf("Warning: Skipping ODS ID %d - not referenced by any active composition object", objData.ID)
            continue
        }
        
        // Validate Width/Height before decoding
        if objData.Width == 0 || objData.Height == 0 {
            log.Printf("Warning: Skipping ODS ID %d - invalid dimensions (Width=%d, Height=%d)", objData.ID, objData.Width, objData.Height)
            continue
        }
        
        // Only validate against screen dimensions if we have real PCS dimensions (not defaults)
        // This avoids false warnings for 4K/UHD content when PCS data is missing
        if d.PCS.Data != nil {
            pcsData, ok := d.PCS.Data.(PresentationCompositionData)
            if ok && pcsData.Width > 0 && pcsData.Height > 0 {
                if objData.Width > pcsData.Width || objData.Height > pcsData.Height {
                    log.Printf("Warning: ODS ID %d dimensions (Width=%d, Height=%d) exceed screen dimensions (Width=%d, Height=%d)", objData.ID, objData.Width, objData.Height, pcsData.Width, pcsData.Height)
                    // Continue processing instead of skipping - allow oversized ODS for compatibility
                }
            }
        }
        
        // Decode image data
        img_encoded := objData.Data
        img_decoded, err := RLEDecode(img_encoded)
        if err != nil {
            log.Printf("Warning: Failed to decode RLE for ODS ID %d: %v", objData.ID, err)
            continue
        }
        // Create image
        if d.PDS.Data == nil {
            log.Printf("Warning: Missing PDS data for ODS ID %d", objData.ID)
            continue
        }
        paletteData, ok := d.PDS.Data.(PaletteData)
        if !ok {
            log.Printf("Warning: Invalid PDS data type for ODS ID %d", objData.ID)
            continue
        }
        img, err := CreateImage(img_decoded, paletteData.Palettes)
        if err != nil {
            log.Printf("Warning: Failed to create image for ODS ID %d: %v", objData.ID, err)
            continue
        }
        // Get image bytes
        img_bytes, err := GetImageBytesJPEG(img)
        if err != nil {
            log.Printf("Warning: Failed to encode JPEG for ODS ID %d: %v", objData.ID, err)
            continue
        }
        // OCR Image
        ocr_result, err = RunOCR(ocr, img_bytes)
        if err != nil {
            log.Printf("Warning: OCR failed for ODS ID %d: %v", objData.ID, err)
            continue
        }
        // Concatenate strings
        if text != "" {
            text = text + "\n" + ocr_result
        } else {
            text = ocr_result
        }
    }
	return text, nil
}

func (d *DisplaySet) Print() {
    // Print DisplaySet as json with 2 space indent
    out, err := json.MarshalIndent(d, "", JSONIndent)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s", string(out))
}
