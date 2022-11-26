package xpm

import (
	"image"
	"io"
)

func init() {
	image.RegisterFormat("xpm", "/* XPM */", Decode, nil)
	image.RegisterFormat("xpm", "static char", Decode, nil)
}

func Decode(r io.Reader) (image.Image, error) {
	return parseXPM(r)
}
