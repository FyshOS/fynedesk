package xpm

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"
)

func parseXPM(data io.Reader) (image.Image, error) {
	var colCount, charSize int
	colors := make(map[string]color.Color)
	var img *image.NRGBA

	rowNum := 0
	scan := bufio.NewScanner(data)
	for scan.Scan() {
		row := scan.Text()
		if row == "" || row[0] != '"' {
			continue
		}
		row = stripQuotes(row)

		if rowNum == 0 {
			w, h, cols, size, err := parseDimensions(row)
			if err != nil {
				return nil, err
			}
			img = image.NewNRGBA(image.Rectangle{image.Point{}, image.Point{w, h}})
			colCount = cols
			charSize = size
		} else if rowNum <= colCount {
			id, c, err := parseColor(row, charSize)
			if err != nil {
				return nil, err
			}

			if id != "" {
				colors[id] = c
			}
		} else {
			parsePixels(row, charSize, rowNum-colCount-1, colors, img)
		}
		rowNum++
	}
	return img, scan.Err()
}

func parseColor(data string, charSize int) (id string, c color.Color, err error) {
	if len(data) == 0 {
		return
	}
	parts := strings.Fields(data)
	if len(parts) == 2 && parts[0] == "c" {
		parts = []string{" ", "c", parts[1]}
	} else if len(parts) != 3 {
		return
	} else if parts[1] != "c" {
		return
	}

	color, err := stringToColor(parts[2])
	return data[:charSize], color, err
}

func parseDimensions(data string) (w, h, i, j int, err error) {
	if len(data) == 0 {
		return
	}
	parts := strings.Split(data, " ")
	if len(parts) != 4 {
		return
	}

	w, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	h, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	i, err = strconv.Atoi(parts[2])
	if err != nil {
		return
	}
	j, err = strconv.Atoi(parts[3])
	return
}

func parsePixels(row string, charSize int, pixRow int, colors map[string]color.Color, img *image.NRGBA) {
	off := pixRow * img.Stride
	chPos := 0
	for i := 0; i < img.Stride/4; i++ {
		id := row[chPos : chPos+charSize]
		c, ok := colors[id]
		if !ok {
			c = color.Transparent
		}

		pos := off + (i * 4)
		r, g, b, a := c.RGBA()
		img.Pix[pos] = uint8(r)
		img.Pix[pos+1] = uint8(g)
		img.Pix[pos+2] = uint8(b)
		img.Pix[pos+3] = uint8(a)
		chPos += charSize
	}
}

func stringToColor(data string) (color.Color, error) {
	if strings.EqualFold("none", data) {
		return color.Transparent, nil
	}

	if data[0] != '#' {
		return color.Transparent, nil // unsupported string like colour name
	}

	c := &color.NRGBA{A: 0xff}
	_, err := fmt.Sscanf(data, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	return c, err
}

func stripQuotes(data string) string {
	if len(data) == 0 || data[0] != '"' {
		return data
	}

	end := strings.Index(data[1:], "\"")
	if end == -1 {
		return data[1:]
	}
	return data[1 : end+1]
}
