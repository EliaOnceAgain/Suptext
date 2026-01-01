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

    for _, ods := range d.ODS {
        objData, ok := ods.Data.(ObjectData)
        if !ok {
            log.Printf("Warning: Invalid ODS data type in DisplaySet")
            continue
        }
        if !objData.Ended {
            continue
        }
        
        // Validate Width/Height before decoding
        if objData.Width == 0 || objData.Height == 0 {
            log.Printf("Warning: Skipping ODS ID %d - invalid dimensions (Width=%d, Height=%d)", objData.ID, objData.Width, objData.Height)
            continue
        }
        if objData.Width > 1920 || objData.Height > 1080 {
            log.Printf("Warning: Skipping ODS ID %d - dimensions exceed reasonable bounds (Width=%d, Height=%d)", objData.ID, objData.Width, objData.Height)
            continue
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
