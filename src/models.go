package suptext

import (
    "encoding/json"
    "fmt"
    "log"
)

const MagicBytes = "PG" // 0x5047
const TimestampAccuracy = 90 // Hz
const SegmentHeaderSize = 13
const JSONIndent = "  "

const (
    PDS uint8 = 0x14 // Palettes definition
    ODS uint8 = 0x15 // Object definition
    PCS uint8 = 0x16 // Presentation composition
    WDS uint8 = 0x17 // Windows definition
    END uint8 = 0x80
)

type SectionData interface {}

type Section struct {
    PTS uint32
    DTS uint32
    Type uint8
    Size uint16
    Data SectionData
}

func (s *Section) Print() {
    // Print Section as json with 2 space indent
    out, err := json.MarshalIndent(s, "", JSONIndent)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s", string(out))
}

