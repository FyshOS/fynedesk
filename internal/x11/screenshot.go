// +build linux openbsd freebsd netbsd

package x11

import (
	"image"
	"math"

	"fyne.io/fyne/v2"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

// CaptureWindow allows x11 code to get the image representation of a screen area.
// The window specified will be captured according to its bounds.
func CaptureWindow(conn *xgb.Conn, win xproto.Window) *image.NRGBA {
	geom, err := xproto.GetGeometry(conn, xproto.Drawable(win)).Reply()
	if err != nil {
		fyne.LogError("Unable to get screen geometry", err)
		return nil
	}
	pix, err := xproto.GetImage(conn, xproto.ImageFormatZPixmap, xproto.Drawable(win),
		0, 0, geom.Width, geom.Height, math.MaxUint32).Reply()
	if err != nil {
		fyne.LogError("Error capturing window content", err)
		return nil
	}

	img := image.NewNRGBA(image.Rect(0, 0, int(geom.Width), int(geom.Height)))
	i := 0
	for y := 0; y < int(geom.Height); y++ {
		for x := 0; x < int(geom.Width); x++ {
			copyPixel(pix.Data, img.Pix, i)
			i += 4
		}
	}
	return img
}

func copyPixel(in []byte, out []uint8, i int) {
	b := in[i]
	g := in[i+1]
	r := in[i+2]
	out[i] = r
	out[i+1] = g
	out[i+2] = b
	out[i+3] = 0xff // some colour maps / depths don't include alpha
}
