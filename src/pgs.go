package suptext

import (
    "fmt"
    "encoding/json"
    "log"
    "os"
    "github.com/otiai10/gosseract/v2"
)

type PGS struct {
    Sections []DisplaySet
}

func (p *PGS) PrintPGS() {
    // Print PGS as json with 2 space indent
    out, err := json.MarshalIndent(p, "", JSONIndent)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s", string(out))
}

func (p *PGS) PrintDisplaySet(i uint) {
    p.Sections[i].Print()
}

func (p *PGS) PrintSectionType(t uint8) {
    for _, section := range p.Sections {
        if section.PCS.Type == t {
            section.PCS.Print()
        } else if section.WDS.Type == t {
            section.WDS.Print()
        } else if section.PDS.Type == t {
            section.PDS.Print()
        } else if section.ODS.Type == t {
            section.ODS.Print()
        } else if section.END.Type == t {
            section.END.Print()
        }
    }
}

func (p *PGS) GetSectionEndTimestamp(start_section int) string {
    for i := start_section + 1; i < len(p.Sections); i++ {
        if p.Sections[i].IsEpochStart() || p.Sections[i].IsEpochEnd(){
            return p.Sections[i].StartTS()
        }
    }
    return FormatMilliseconds(p.Sections[start_section].PCS.PTS + 10000)
}

func (p *PGS) ToSRT(fout *os.File) error {
    client := gosseract.NewClient()
    defer client.Close()

    var srt_section_id uint
    srt_section_id = 1
    for i, ds := range p.Sections {
        if !ds.IsEpochStart() {
            continue
        }
        ets := p.GetSectionEndTimestamp(i)
        ds.AppendSRT(client, fout, srt_section_id, ets)
        srt_section_id++
    }

    return nil
}
