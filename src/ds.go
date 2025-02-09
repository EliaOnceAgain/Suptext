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
    return d.PCS.Data.(PresentationCompositionData).State == 0x80
}

func (d *DisplaySet) IsEpochEnd() bool {
    return d.PCS.Data.(PresentationCompositionData).State == 0
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
        // Decode image data
        img_encoded := ods.Data.(ObjectData).Data
        img_decoded, err := RLEDecode(img_encoded)
        if err != nil {
            return text, err
        }
        // Create image
        img, err := CreateImage(img_decoded, d.PDS.Data.(PaletteData).Palettes)
        if err != nil {
            return text, err
        }
        // Get image bytes
        img_bytes, err := GetImageBytesJPEG(img)
        if err != nil {
            return text, err
        }
        // OCR Image
        ocr_result, err = RunOCR(ocr, img_bytes)
        if err != nil {
            return text, err
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
