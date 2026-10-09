//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package composit

import (
	"image"
	"image/color"
	"math"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

// capturePixmap reads the contents of an X pixmap into a Go RGBA image.
// The result is premultiplied-alpha RGBA: when argb is true the pixmap's alpha
// is applied to the colour channels, otherwise alpha is forced to 0xff (fully
// opaque). This image choice is optimised for the Fyne render pipeline.
func capturePixmap(conn *xgb.Conn, drawable xproto.Drawable, width, height uint16, argb bool, buf *image.RGBA) *image.RGBA {
	w, h := int(width), int(height)
	if w == 0 || h == 0 {
		return nil
	}

	img := buf
	if img == nil || img.Rect.Dx() != w || img.Rect.Dy() != h || img.Stride != w*4 {
		img = image.NewRGBA(image.Rect(0, 0, w, h))
	}
	if !captureRegion(conn, drawable, img.Rect, argb, img) {
		return nil
	}
	return img
}

// captureRegion reads the pixels in region r of an X pixmap into the same
// region of img, converting them as capturePixmap does. The rest of img is
// left as it was. It returns false if the pixels could not be read.
func captureRegion(conn *xgb.Conn, drawable xproto.Drawable, r image.Rectangle, argb bool, img *image.RGBA) bool {
	reply, err := xproto.GetImage(conn, xproto.ImageFormatZPixmap, drawable,
		int16(r.Min.X), int16(r.Min.Y), uint16(r.Dx()), uint16(r.Dy()), math.MaxUint32).Reply()
	if err != nil || reply == nil {
		return false
	}

	// X11 ZPixmap format: 4 bytes per pixel in BGRx or BGRA order
	if len(reply.Data) < r.Dx()*r.Dy()*4 {
		return false
	}
	convertZPixmapRegion(img, r, reply.Data, argb)
	return true
}

// convertZPixmapRegion converts X11 ZPixmap data holding the rows of region r
// into that region of img, as convertZPixmap does.
func convertZPixmapRegion(img *image.RGBA, r image.Rectangle, data []byte, argb bool) {
	row := r.Dx() * 4
	for y := r.Min.Y; y < r.Max.Y; y++ {
		off := img.PixOffset(r.Min.X, y)
		convertZPixmap(img.Pix[off:off+row], data[:row], argb)
		data = data[row:]
	}
}

// copyRegion copies the pixels in region r of src into the same region of dst,
// an image with the same bounds.
func copyRegion(dst, src *image.RGBA, r image.Rectangle) {
	row := r.Dx() * 4
	for y := r.Min.Y; y < r.Max.Y; y++ {
		off := dst.PixOffset(r.Min.X, y)
		copy(dst.Pix[off:off+row], src.Pix[off:off+row])
	}
}

// damagedRegion returns where in the pixmap of a window with the given border
// width a damage notification's area lies. Damage is reported from the origin
// of the window, which the pixmap surrounds with its border.
func damagedRegion(area xproto.Rectangle, border uint16) image.Rectangle {
	x, y := int(area.X)+int(border), int(area.Y)+int(border)
	return image.Rect(x, y, x+int(area.Width), y+int(area.Height))
}

// convertZPixmap converts X11 ZPixmap data (4 bytes/pixel in BGRx or BGRA
// order) into premultiplied-alpha RGBA byte order (R,G,B,A), writing into dst.
// Only the leading min(len(dst), len(src)) bytes (rounded down to a whole
// pixel) are touched.
//
// When argb is false the alpha byte is forced to 0xff (fully opaque), so the
// colour channels are an unchanged BGR->RGB swap. When argb is true the colour
// channels are premultiplied by alpha - so the painter can upload the RGBA pixels
// directly with no per-frame conversion or allocation.
func convertZPixmap(dst, src []byte, argb bool) {
	n := len(dst)
	if len(src) < n {
		n = len(src)
	}
	n -= n % 4
	dst = dst[:n:n]
	src = src[:n:n]

	if argb {
		for i := 0; i < n; i += 4 {
			d := dst[i : i+4 : i+4]
			s := src[i : i+4 : i+4]
			a := uint32(s[3])
			d[0] = uint8((uint32(s[2])*a + 127) / 255) // R
			d[1] = uint8((uint32(s[1])*a + 127) / 255) // G
			d[2] = uint8((uint32(s[0])*a + 127) / 255) // B
			d[3] = s[3]                                // A
		}
		return
	}
	for i := 0; i < n; i += 4 {
		d := dst[i : i+4 : i+4]
		s := src[i : i+4 : i+4]
		d[0], d[1], d[2], d[3] = s[2], s[1], s[0], 0xff
	}
}

// roundCorners clears pixels outside a circular arc of the given radius
// at each corner of the image, making them fully transparent.
func roundCorners(img *image.RGBA, radius int) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if radius <= 0 || w < radius*2 || h < radius*2 {
		return
	}

	transparent := color.RGBA{}
	r2 := float64(radius) * float64(radius)
	for dy := 0; dy < radius; dy++ {
		for dx := 0; dx < radius; dx++ {
			cx := float64(radius-dx) - 0.5
			cy := float64(radius-dy) - 0.5
			if cx*cx+cy*cy > r2 {
				img.SetRGBA(dx, dy, transparent)         // top-left
				img.SetRGBA(w-1-dx, dy, transparent)     // top-right
				img.SetRGBA(dx, h-1-dy, transparent)     // bottom-left
				img.SetRGBA(w-1-dx, h-1-dy, transparent) // bottom-right
			}
		}
	}
}
