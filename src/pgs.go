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

func (p *PGS) GetSectionEndTimestamp(startSection int) string {
    currentPTS := p.Sections[startSection].PCS.PTS

    // Search for the next epoch start or end
    for i := startSection + 1; i < len(p.Sections); i++ {
        nextPCS := p.Sections[i].PCS.Data.(PresentationCompositionData)
        if p.Sections[i].IsEpochStart() || p.Sections[i].IsEpochEnd() || nextPCS.State != p.Sections[startSection].PCS.Data.(PresentationCompositionData).State {
            return FormatMilliseconds(p.Sections[i].PCS.PTS)
        }
    }

    // No future epoch/end found, fallback: last known PTS + safe buffer (e.g., 5s)
    safeEnd := currentPTS + 5000 // milliseconds
    return FormatMilliseconds(safeEnd)
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
