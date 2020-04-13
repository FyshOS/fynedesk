package icns

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

var jpeg2000header = []byte{0x00, 0x00, 0x00, 0x0c, 0x6a, 0x50, 0x20, 0x20}

// Decode finds the largest icon listed in the icns file and returns it,
// ignoring all other sizes. The format returned will be whatever the icon data
// is, typically jpeg or png.
func Decode(r io.Reader) (image.Image, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	icnsHeader := data[0:4]
	if string(icnsHeader) != "icns" {
		return nil, fmt.Errorf("invalid header for icns file")
	}
	fileSize := binary.BigEndian.Uint32(data[4:8])
	icons := []iconReader{}
	read := uint32(8)
	for read < fileSize {
		next := data[read : read+4]
		read += 4
		switch string(next) {
		case "TOC ":
			tocSize := binary.BigEndian.Uint32(data[read : read+4])
			read += tocSize-4 // size includes header and size fields
			continue
		case "icnV":
			read += 4
			continue
		}

		dataSize := binary.BigEndian.Uint32(data[read : read+4])
		read += 4
		if dataSize == 0 {
			continue // no content, we're not interested
		}

		iconData := data[read : read+dataSize-8]
		read += dataSize-8 // size includes header and size fields

		if isOsType(string(next)) {
			if bytes.Equal(iconData[:8], jpeg2000header) {
				continue // skipping JPEG2000
			}

			icons = append(icons, iconReader{
				OsType: osTypeFromID(string(next)),
				r:      bytes.NewBuffer(iconData),
			})
		}
	}
	if len(icons) == 0 {
		return nil, fmt.Errorf("no icons found")
	}
	var biggest iconReader
	for _, icon := range icons {
		if icon.Size > biggest.Size {
			biggest = icon
		}
	}
	img, _, err := image.Decode(biggest.r)
	if err != nil {
		return nil, errors.Wrap(err, "decoding largest image")
	}
	return img, nil
}

type iconReader struct {
	OsType
	r io.Reader
}

func isOsType(ID string) bool {
	_, ok := getTypeFromID(ID)
	return ok
}

func init() {
	image.RegisterFormat("icns", "icns", Decode, nil)
}
