package suptext

import (
    "bufio"
    "encoding/binary"
    "fmt"
    "io"
    "time"
)

func ReadPGS(r *bufio.Reader) (PGS, error) {
    pgs := PGS{}
    ds := DisplaySet{}
    var running_ods *ObjectData

    for {
        // Read section header bytes
        bytes := make([]byte, SegmentHeaderSize)
        if _, err := io.ReadFull(r, bytes); err != nil {
            if err != io.EOF {
                return pgs, err
            } else {
                break
            }
        }
        // Parse section header
        section, err := NewSection(bytes)
        if err != nil {
            return pgs, err
        }
        // If section has no data, add as is
        if 0 == section.Size {
            ds.END = section
            pgs.Sections = append(pgs.Sections, ds)
            ds = DisplaySet{}
            continue
        }
        // Read section data bytes
        section_data := make([]byte, section.Size)
        if _, err := io.ReadFull(r, section_data); err != nil {
            return pgs, err
        }
        // Parse section data
        if section.Type == PCS {
            data, err := NewPresentationData(section_data)
            if err != nil {
                return pgs, err
            }
            section.Data = data
            ds.PCS = section
        } else if section.Type == WDS {
            data, err := NewWindowsData(section_data)
            if err != nil {
                return pgs, err
            }
            section.Data = data
            ds.WDS = section
        } else if section.Type == PDS {
            data, err := NewPaletteData(section_data)
            if err != nil {
                return pgs, err
            }
            section.Data = data
            ds.PDS = section
        } else if section.Type == ODS {
            data, err := NewObjectData(section_data)
            if err != nil {
                return pgs, err
            }
            // Merge to previous ODS if it wasn't ended
            if running_ods != nil {
                err = running_ods.MergeSequence(data)
                if err != nil {
                   return pgs, err
                }
                // If ODS now ended, add it to DS
                if running_ods.Ended {
                    section.Data = *running_ods
                    ds.ODS = append(ds.ODS, section)
                    running_ods = nil
                }
            } else if !data.Ended {
                running_ods = &data
            } else {
                section.Data = data
                ds.ODS = append(ds.ODS, section)
            }
        } else {
            return pgs, fmt.Errorf("Segment type not supported: 0x%x", section.Type)
        }
    }

    return pgs, nil
}

func NewSection(bytes []byte) (Section, error) {
    if string(bytes[:2]) != MagicBytes {
        return Section{}, fmt.Errorf("Invalid header magic '%s' != '%s'", string(bytes[:2]), MagicBytes)
    }

    header := Section{
        PTS: binary.BigEndian.Uint32(bytes[2:6]) / TimestampAccuracy,
        DTS: binary.BigEndian.Uint32(bytes[6:10]) / TimestampAccuracy,
        Type: uint8(bytes[10]),
        Size: binary.BigEndian.Uint16(bytes[11:13]),
    }

    return header, nil
}

func FormatMilliseconds(ts uint32) string {
    duration := time.Duration(ts) * time.Millisecond
    h := int(duration.Hours())
    m := int(duration.Minutes()) % 60
    s := int(duration.Seconds()) % 60
    ms := int(duration.Milliseconds()) % 1000
    return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
