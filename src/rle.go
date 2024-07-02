package suptext

import (
    "bytes"
    "fmt"
    "image"
	"image/color"
	"image/jpeg"
	"github.com/otiai10/gosseract/v2"
)

func CreateImage(pixels [][]uint8, palettes [256]PaletteDefinition) (*image.RGBA, error) {
    height := len(pixels)
	width := len(pixels[0])
	img := image.NewRGBA(image.Rect(0, 0, width, height))

    // Validate dimensions
	if height == 0 || width == 0 {
		return img, fmt.Errorf("Failed creating image: empty matrix")
	}

    // Fill colors
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			p := palettes[pixels[row][col]]
			c := color.NYCbCrA{
			    YCbCr: color.YCbCr{
			        Y: p.Y,
			        Cb: p.Cb,
			        Cr: p.Cr,
                },
			    A: p.A,
			}
			img.Set(col, row, c)
		}
	}

	return img, nil
}

func GetImageBytesJPEG(img *image.RGBA) ([]byte, error) {
    var b bytes.Buffer
	err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 100})
	if err != nil {
	    return nil, err
	}
	return b.Bytes(), err
}

func RunOCR(ocr *gosseract.Client, img []byte) (string, error) {
    ocr.SetImageFromBytes(img)
    return ocr.Text()
}

func RLEDecode(bytes []byte) ([][]uint8, error) {
    var img [][]uint8
    var line []uint8

    for i := 0; i < len(bytes); {
        // Single pixel specific color
        if bytes[i] != 0 {
            line = append(line, uint8(bytes[i]))
            i += 1
            continue
        }
        // 0, 0 = line end
        if bytes[i + 1] == 0 {
            img = append(img, line)
            line = nil
            i += 2
            continue
        }
        // Defaults
        color := uint8(0)
        count := uint16(bytes[i + 1] & 0x3f)
        // Flags
        has_color := bytes[i + 1] & 0x80 != 0
        is_14bit_count := bytes[i + 1] & 0x40 != 0
        // Adapt defaults according to flags
        if !is_14bit_count && !has_color {
            i += 2
        } else if is_14bit_count && !has_color {
            count = count<<8 | uint16(bytes[i + 2])
            i += 3
        } else if !is_14bit_count && has_color {
            color = uint8(bytes[i + 2])
            i += 3
        } else {
            count = count<<8 | uint16(bytes[i + 2])
            color = uint8(bytes[i + 3])
            i += 4
        }
        // Append `count` times `color` to line
        colors := make([]uint8, count)
        for j := uint16(0); j < count; j++ {
            colors[j] = color
        }
        line = append(line, colors...)
    }

    if len(line) != 0 {
        return img, fmt.Errorf("Leftover values in unterminated line")
	}

    return img, nil
}
