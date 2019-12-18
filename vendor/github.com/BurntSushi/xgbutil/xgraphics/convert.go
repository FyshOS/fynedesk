package xgraphics

/*
A set of conversion functions for some image types defined in the Go standard
library. They can be up to 80% faster because the inner loop doesn't use
interfaces. Wow.

Note that these functions assume that the source and destination are precisely
the same size.
*/

import (
	"image"
	"image/color"
)

// convertImage converts any image implementing the image.Image interface to
// an xgraphics.Image type. This is *slow*.
func convertImage(dest *Image, src image.Image) {
	var r, g, b, a uint32
	var x, y, i int

	for x = dest.Rect.Min.X; x < dest.Rect.Max.X; x++ {
		for y = dest.Rect.Min.Y; y < dest.Rect.Max.Y; y++ {
			r, g, b, a = src.At(x, y).RGBA()
			i = dest.PixOffset(x, y)
			dest.Pix[i+0] = uint8(b >> 8)
			dest.Pix[i+1] = uint8(g >> 8)
			dest.Pix[i+2] = uint8(r >> 8)
			dest.Pix[i+3] = uint8(a >> 8)
		}
	}
}

func convertXImage(dest *Image, src *Image) {
	copy(dest.Pix, src.Pix)
}

func convertYCbCr(dest *Image, src *image.YCbCr) {
	var r, g, b uint8
	var x, y, i, yi, ci int

	for x = dest.Rect.Min.X; x < dest.Rect.Max.X; x++ {
		for y = dest.Rect.Min.Y; y < dest.Rect.Max.Y; y++ {
			yi, ci = src.YOffset(x, y), src.COffset(x, y)
			r, g, b = color.YCbCrToRGB(src.Y[yi], src.Cb[ci], src.Cr[ci])
			i = dest.PixOffset(x, y)
			dest.Pix[i+0] = b
			dest.Pix[i+1] = g
			dest.Pix[i+2] = r
			dest.Pix[i+3] = 0xff
		}
	}
}

func convertRGBA(dest *Image, src *image.RGBA) {
	var x, y, i, si int

	for x = dest.Rect.Min.X; x < dest.Rect.Max.X; x++ {
		for y = dest.Rect.Min.Y; y < dest.Rect.Max.Y; y++ {
			si = src.PixOffset(x, y)
			i = dest.PixOffset(x, y)
			dest.Pix[i+0] = src.Pix[si+2]
			dest.Pix[i+1] = src.Pix[si+1]
			dest.Pix[i+2] = src.Pix[si+0]
			dest.Pix[i+3] = src.Pix[si+3]
		}
	}
}

func convertRGBA64(dest *Image, src *image.RGBA64) {
	var x, y, i, si int

	for x = dest.Rect.Min.X; x < dest.Rect.Max.X; x++ {
		for y = dest.Rect.Min.Y; y < dest.Rect.Max.Y; y++ {
			si = src.PixOffset(x, y)
			i = dest.PixOffset(x, y)
			dest.Pix[i+0] = src.Pix[si+4]
			dest.Pix[i+1] = src.Pix[si+2]
			dest.Pix[i+2] = src.Pix[si+0]
			dest.Pix[i+3] = src.Pix[si+6]
		}
	}
}

func convertNRGBA(dest *Image, src *image.NRGBA) {
	var x, y, i, si int
	var a uint16

	for x = dest.Rect.Min.X; x < dest.Rect.Max.X; x++ {
		for y = dest.Rect.Min.Y; y < dest.Rect.Max.Y; y++ {
			si = src.PixOffset(x, y)
			i = dest.PixOffset(x, y)
			a = uint16(src.Pix[si+3])

			dest.Pix[i+0] = uint8((uint16(src.Pix[si+2]) * a) / 0xff)
			dest.Pix[i+1] = uint8((uint16(src.Pix[si+1]) * a) / 0xff)
			dest.Pix[i+2] = uint8((uint16(src.Pix[si+0]) * a) / 0xff)
			dest.Pix[i+3] = src.Pix[si+3]
		}
	}
}

func convertNRGBA64(dest *Image, src *image.NRGBA64) {
	var x, y, i, si int
	var a uint16

	for x = dest.Rect.Min.X; x < dest.Rect.Max.X; x++ {
		for y = dest.Rect.Min.Y; y < dest.Rect.Max.Y; y++ {
			si = src.PixOffset(x, y)
			i = dest.PixOffset(x, y)
			a = uint16(src.Pix[si+6])

			dest.Pix[i+0] = uint8((uint16(src.Pix[si+4]) * a) / 0xff)
			dest.Pix[i+1] = uint8((uint16(src.Pix[si+2]) * a) / 0xff)
			dest.Pix[i+2] = uint8((uint16(src.Pix[si+0]) * a) / 0xff)
			dest.Pix[i+3] = src.Pix[si+6]
		}
	}
}
