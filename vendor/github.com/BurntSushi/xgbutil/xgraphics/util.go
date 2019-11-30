package xgraphics

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/BurntSushi/graphics-go/graphics"
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
)

/*
xgraphics/util.go contains a variety of image manipulation functions that
are not specific to xgraphics.Image.
*/

// Scale is a simple wrapper around graphics.Scale.
func Scale(img image.Image, width, height int) draw.Image {
	dimg := image.NewRGBA(image.Rect(0, 0, width, height))
	graphics.Scale(dimg, img)

	return dimg
}

// Alpha will modify the alpha channel of the image such that:
// existingAlpha = existingAlpha * (givenAlpha / 100.0)
func Alpha(dest *Image, alpha int) {
	r := dest.Bounds()

	var a, x, y, i int
	for x = r.Min.X; x < r.Max.X; x++ {
		for y = r.Min.Y; y < r.Max.Y; y++ {
			i = dest.PixOffset(x, y)
			a = int(dest.Pix[i+3])
			dest.Pix[i+3] = uint8((a * alpha) / 100)
		}
	}
}

// Blend alpha blends the src image (starting at the spt Point) into the
// dest image.
// If you're blending into a solid background color, use BlendBgColor
// instead. (It's more efficient.)
// Blend does not (currently) blend with the destination's alpha channel,
// only the source's alpha channel.
func Blend(dest *Image, src image.Image, sp image.Point) {
	rsrc, dsrc := src.Bounds(), dest.Bounds()
	_, smxx, _, smxy := rsrc.Min.X, rsrc.Max.X, rsrc.Min.Y, rsrc.Max.Y
	dmnx, dmxx, dmny, dmxy := dsrc.Min.X, dsrc.Max.X, dsrc.Min.Y, dsrc.Max.Y

	var sx, dx, sy, dy int
	var sr, sg, sb, sa uint32
	var bgra BGRA
	var alpha float64
	for sx, dx = sp.X, dmnx; sx < smxx && dx < dmxx; sx, dx = sx+1, dx+1 {
		for sy, dy = sp.Y, dmny; sy < smxy && dy < dmxy; sy, dy = sy+1, dy+1 {
			sr, sg, sb, sa = src.At(sx, sy).RGBA()
			bgra = dest.At(dx, dy).(BGRA)
			alpha = float64(uint8(sa)) / 255.0

			dest.SetBGRA(dx, dy, BGRA{
				blend(uint8(bgra.B), uint8(sb), alpha),
				blend(uint8(bgra.G), uint8(sg), alpha),
				blend(uint8(bgra.R), uint8(sr), alpha),
				0xff,
			})
		}
	}
}

// BlendBgColor blends the Image (receiver) into the background color
// specified. This is more efficient than creating a background image and
// blending with Blend.
func BlendBgColor(dest *Image, c color.Color) {
	r := dest.Bounds()
	cr32, cg32, cb32, _ := c.RGBA()
	cr, cg, cb := uint8(cr32), uint8(cg32), uint8(cb32)

	var bgra BGRA
	var alpha float64
	for x := r.Min.X; x < r.Max.X; x++ {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			bgra = dest.At(x, y).(BGRA)
			alpha = float64(bgra.A) / 255.0
			dest.SetBGRA(x, y, BGRA{
				B: blend(cb, bgra.B, alpha),
				G: blend(cg, bgra.G, alpha),
				R: blend(cr, bgra.R, alpha),
				A: 0xff,
			})
		}
	}
}

// Blend returns the blended alpha color for src and dest colors.
// This assumes that the destination has alpha = 1.
func BlendBGRA(dest, src BGRA) BGRA {
	alpha := float64(src.A) / 255.0
	return BGRA{
		B: blend(dest.B, src.B, alpha),
		G: blend(dest.G, src.G, alpha),
		R: blend(dest.R, src.R, alpha),
		A: 0xff,
	}
}

// blend calculates the value of a color given some alpha value in [0, 1]
// and a source and destination color. Note that this assumes that the
// destination is fully opaque (has an alpha value of 1).
func blend(d, s uint8, alpha float64) uint8 {
	return uint8(float64(s)*alpha + float64(d)*(1-alpha))
}

// FreePixmap is a convenience function for destroying a pixmap resource
// on the X server.
// If you're using an xgraphics.Image value, then its Destroy method will call
// this for you.
func FreePixmap(X *xgbutil.XUtil, pixid xproto.Pixmap) {
	xproto.FreePixmap(X.Conn(), pixid)
}

// FindIcon takes a window id and attempts to return an xgraphics.Image of
// that window's icon.
// It will first try to look for an icon in _NET_WM_ICON that is closest to
// the size specified.
// If there are no icons in _NET_WM_ICON, then WM_HINTS will be checked for
// an icon.
// If an icon is found from either one and doesn't match the size
// specified, it will be scaled to that size.
// If the width and height are 0, then the largest icon will be returned with
// no scaling.
// If an icon is not found, an error is returned.
func FindIcon(X *xgbutil.XUtil, wid xproto.Window,
	width, height int) (*Image, error) {

	var ewmhErr, icccmErr error

	// First try to get a EWMH style icon.
	icon, ewmhErr := findIconEwmh(X, wid, width, height)
	if ewmhErr != nil { // now look for an icccm-style icon
		icon, icccmErr = findIconIcccm(X, wid)
		if icccmErr != nil {
			return nil, fmt.Errorf("Neither a EWMH-style or ICCCM-style icon "+
				"could be found for window id %x because: %s *AND* %s",
				wid, ewmhErr, icccmErr)
		}
	}

	// We should have a valid xgraphics.Image if we're here.
	// If the size doesn't match what's preferred, scale it.
	if width != 0 && height != 0 {
		if icon.Bounds().Dx() != width || icon.Bounds().Dy() != height {
			icon = icon.Scale(width, height)
		}
	}
	return icon, nil
}

// findIconEwmh helps FindIcon by trying to return an ewmh-style icon that is
// closest to the preferred size specified.
func findIconEwmh(X *xgbutil.XUtil, wid xproto.Window,
	width, height int) (*Image, error) {

	icons, err := ewmh.WmIconGet(X, wid)
	if err != nil {
		return nil, err
	}

	icon := FindBestEwmhIcon(width, height, icons)
	if icon == nil {
		return nil, fmt.Errorf("Could not find any _NET_WM_ICON icon.")
	}

	return NewEwmhIcon(X, icon), nil
}

// findIconIcccm helps FindIcon by trying to return an icccm-style icon.
func findIconIcccm(X *xgbutil.XUtil, wid xproto.Window) (*Image, error) {
	hints, err := icccm.WmHintsGet(X, wid)
	if err != nil {
		return nil, err
	}

	// Only continue if the WM_HINTS flags say an icon is specified and
	// if at least one of icon pixmap or icon mask is non-zero.
	if hints.Flags&icccm.HintIconPixmap == 0 ||
		(hints.IconPixmap == 0 && hints.IconMask == 0) {

		return nil, fmt.Errorf("No icon found in WM_HINTS.")
	}

	return NewIcccmIcon(X, hints.IconPixmap, hints.IconMask)
}

// FindBestEwmhIcon takes width/height dimensions and a slice of *ewmh.WmIcon
// and finds the best matching icon of the bunch. We always prefer bigger.
// If no icons are bigger than the preferred dimensions, use the biggest
// available. Otherwise, use the smallest icon that is greater than or equal
// to the preferred dimensions. The preferred dimensions is essentially
// what you'll likely scale the resulting icon to.
// If width and height are 0, then the largest icon found will be returned.
func FindBestEwmhIcon(width, height int, icons []ewmh.WmIcon) *ewmh.WmIcon {
	// nada nada limonada
	if len(icons) == 0 {
		return nil
	}

	parea := width * height // preferred size
	best := -1

	// If zero area, set it to the largest possible.
	if parea == 0 {
		parea = math.MaxInt32
	}

	var bestArea, iconArea int

	for i, icon := range icons {
		// the first valid icon we've seen; use it!
		if best == -1 {
			best = i
			continue
		}

		// load areas for comparison
		bestArea = int(icons[best].Width * icons[best].Height)
		iconArea = int(icon.Width * icon.Height)

		// We don't always want to accept bigger icons if our best is
		// already bigger. But we always want something bigger if our best
		// is insufficient.
		if (iconArea >= parea && iconArea <= bestArea) ||
			(bestArea < parea && iconArea > bestArea) {
			best = i
		}
	}

	if best > -1 {
		return &icons[best]
	}
	return nil
}
