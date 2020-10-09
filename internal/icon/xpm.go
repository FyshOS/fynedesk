package icon

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
)

func parseXPM(data []byte) image.Image {
	var colCount, charSize int
	colors := make(map[string]color.Color)
	var img *image.NRGBA

	rowStart := 0
	rowNum := 0
	for i, b := range data {
		if b == '\n' {
			row := string(data[rowStart:i])
			rowStart = i+1
			if row == "" || row[0] != '"' {
				continue
			}
			row = stripQuotes(row)

			if rowNum == 0 {
				w, h, cols, size := parseDimensions(row)
				img = image.NewNRGBA(image.Rectangle{image.Point{}, image.Point{w, h}})
				colCount = cols
				charSize = size
			} else if rowNum <= colCount {
				id, c := parseColor(row)
				if id != "" {
					colors[id] = c
				}
			} else {
				parsePixels(row, charSize, rowNum-colCount-1, colors, img)
			}
			rowNum++
		}
	}


	return img
}

func parseColor(data string) (id string, c color.Color) {
	if len(data) == 0 {
		return
	}
	parts := strings.Fields(data)
	if len(parts) != 3 {
		if len(parts) == 2 && parts[0] == "c" {
			parts = []string{" ", "c", parts[1]}
		} else {
			return
		}
	}

	if parts[1] != "c" {
		return
	}

	return parts[0], stringToColor(parts[2])
}

func parseDimensions(data string) (w, h, i, j int) {
	if len(data) == 0 {
		return
	}
	parts := strings.Split(data, " ")
	if len(parts) != 4 {
		return
	}

	w, _ = strconv.Atoi(parts[0])
	h, _ = strconv.Atoi(parts[1])
	i, _ = strconv.Atoi(parts[2])
	j, _ = strconv.Atoi(parts[3])
	return
}

func parsePixels(row string, charSize int, pixRow int, colors map[string]color.Color, img *image.NRGBA) {
	off := pixRow*img.Stride
	for i := 0; i < len(row); i += charSize {
		id := row[i:i+charSize]
		c, ok := colors[id]
		if !ok {
			c = color.Transparent
		}

		pos := off+(i*4)
		r, g, b, a := c.RGBA()
		img.Pix[pos] = uint8(r)
		img.Pix[pos+1] = uint8(g)
		img.Pix[pos+2] = uint8(b)
		img.Pix[pos+3] = uint8(a)
	}
}

func stringToColor(data string) color.Color {
	if strings.EqualFold("none", data) {
		return color.Transparent
	}

	if data[0] != '#' {
		return color.Transparent // unsupported string like colour name
	}

	c := &color.NRGBA{A: 0xff}
	_, _ = fmt.Sscanf(data, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	return c
}

func stripQuotes(data string) string {
	if len(data) == 0 || data[0] != '"' {
		return data
	}

	end := strings.Index(data[1:], "\"")
	if end == -1 {
		return data[1:]
	}
	return data[1:end+1]
}