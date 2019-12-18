package xgraphics

/*
xgraphics/image.go contains an implementation of the draw.Image interface.

RGBA could feasibly be used, but the representation of image data is dependent
upon the configuration of the X server.

For the time being, I'm hard-coding a lot of that configuration for the common
case. Namely:

Byte order: least significant byte first
Depth: 24
Bits per pixel: 32

This will have to be fixed for this to be truly compatible with any X server.

Most of the code is based heavily on the implementation of common images in
the Go standard library.

Manipulating images isn't something I've had much experience with, so if it
seems like I'm doing something stupid, I probably am.
*/

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"os"

	"github.com/BurntSushi/graphics-go/graphics"
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xwindow"
)

// Model for the BGRA color type.
var BGRAModel color.Model = color.ModelFunc(bgraModel)

type Image struct {
	// X images must be tied to an X connection.
	X *xgbutil.XUtil

	// X images must also be tied to a pixmap (its drawing surface).
	// Calls to 'XDraw' will draw data to this pixmap.
	// Calls to 'XPaint' will tell X to show the pixmap on some window.
	Pixmap xproto.Pixmap

	// Pix holds the image's pixels in BGRA order, so that they don't need
	// to be swapped for every PutImage request.
	Pix []uint8

	// Stride corresponds to the number of elements in Pix between two pixels
	// that are vertically adjacent.
	Stride int

	// The geometry of the image.
	Rect image.Rectangle

	// Whether this is a sub-image or not.
	// This is useful to know when sending data or setting surfaces.
	// Namely, sub-images cannot be set as surfaces and sub-images, when
	// being drawn, only have its pixels sent to X instead of the whole image.
	Subimg bool
}

// New returns a new instance of Image with colors initialized to black
// for the geometry given.
// New will also create an X pixmap. When you are no longer using this
// image, you should call Destroy so that the X pixmap can be freed on the
// X server.
// If 'X' is nil, then a new connection will be made. This is usually a bad
// idea, particularly if you're making a lot of small images, but can be
// used to achieve true parallelism. (Particularly useful when painting large
// images.)
func New(X *xgbutil.XUtil, r image.Rectangle) *Image {
	var err error
	if X == nil {
		X, err = xgbutil.NewConn()
		if err != nil {
			xgbutil.Logger.Panicf("Could not create a new connection when "+
				"creating a new xgraphics.Image value because: %s", err)
		}
	}

	return &Image{
		X:      X,
		Pixmap: 0,
		Pix:    make([]uint8, 4*r.Dx()*r.Dy()),
		Stride: 4 * r.Dx(),
		Rect:   r,
		Subimg: false,
	}
}

// Destroy frees the pixmap resource being used by this image.
// It should be called whenever the image will no longer be drawn or painted.
func (im *Image) Destroy() {
	if im.Pixmap != 0 {
		xproto.FreePixmap(im.X.Conn(), im.Pixmap)
		im.Pixmap = 0
	}
}

// Scale will scale the image to the size provided.
// Note that this will destroy the current pixmap associated with this image.
// After scaling, XSurfaceSet will need to be called for each window that
// this image is painted to. (And obviously, XDraw and XPaint will need to
// be called again.)
func (im *Image) Scale(width, height int) *Image {
	dimg := New(im.X, image.Rect(0, 0, width, height))
	graphics.Scale(dimg, im)
	im.Destroy()

	return dimg
}

// WritePng encodes the image to w as a png.
func (im *Image) WritePng(w io.Writer) error {
	return png.Encode(w, im)
}

// SavePng writes the Image to a file with name as a png.
func (im *Image) SavePng(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	return im.WritePng(file)
}

// ColorModel returns the color.Model used by the Image struct.
func (im *Image) ColorModel() color.Model {
	return BGRAModel
}

// Bounds returns the rectangle representing the geometry of Image.
func (im *Image) Bounds() image.Rectangle {
	return im.Rect
}

// At returns the color at the specified pixel.
func (im *Image) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(im.Rect)) {
		return BGRA{}
	}
	i := im.PixOffset(x, y)
	return BGRA{
		B: im.Pix[i],
		G: im.Pix[i+1],
		R: im.Pix[i+2],
		A: im.Pix[i+3],
	}
}

// Set satisfies the draw.Image interface by allowing the color of a pixel
// at (x, y) to be changed.
func (im *Image) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(im.Rect)) {
		return
	}

	i := im.PixOffset(x, y)
	cc := BGRAModel.Convert(c).(BGRA)
	im.Pix[i] = cc.B
	im.Pix[i+1] = cc.G
	im.Pix[i+2] = cc.R
	im.Pix[i+3] = cc.A
}

// SetBGRA is like set, but without the type assertion.
func (im *Image) SetBGRA(x, y int, c BGRA) {
	if !(image.Point{x, y}.In(im.Rect)) {
		return
	}

	i := im.PixOffset(x, y)
	im.Pix[i] = c.B
	im.Pix[i+1] = c.G
	im.Pix[i+2] = c.R
	im.Pix[i+3] = c.A
}

// For transforms every pixel color to the color returned by 'each' given
// an (x, y) position.
func (im *Image) For(each func(x, y int) BGRA) {
	for x := im.Rect.Min.X; x < im.Rect.Max.X; x++ {
		for y := im.Rect.Min.Y; y < im.Rect.Max.Y; y++ {
			im.SetBGRA(x, y, each(x, y))
		}
	}
}

// ForExp is like For, but bypasses image.Color types.
// (So it should be faster.)
func (im *Image) ForExp(each func(x, y int) (r, g, b, a uint8)) {
	var x, y, i int
	var r, g, b, a uint8
	for x = im.Rect.Min.X; x < im.Rect.Max.X; x++ {
		for y = im.Rect.Min.Y; y < im.Rect.Max.Y; y++ {
			i = im.PixOffset(x, y)
			r, g, b, a = each(x, y)

			im.Pix[i+0] = b
			im.Pix[i+1] = g
			im.Pix[i+2] = r
			im.Pix[i+3] = a
		}
	}
}

// SubImage provides a sub image of Image without copying image data.
// N.B. The standard library defines a similar function, but returns an
// image.Image. Here, we return xgraphics.Image so that we can use the extra
// methods defined by xgraphics on it.
//
// This method is cheap to call. It should be used to update only specific
// regions of an X pixmap to avoid sending an entire image to the X server when
// only a piece of it is updated.
//
// Note that if the intersection of `r` and `im` is empty, `nil` is returned.
func (im *Image) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(im.Rect)
	if r.Empty() {
		return nil
	}

	i := im.PixOffset(r.Min.X, r.Min.Y)
	return &Image{
		X:      im.X,
		Pixmap: im.Pixmap,
		Pix:    im.Pix[i:],
		Stride: im.Stride,
		Rect:   r,
		Subimg: true,
	}
}

// PixOffset returns the index of the frst element of the Pix data that
// corresponds to the pixel at (x, y).
func (im *Image) PixOffset(x, y int) int {
	return (y-im.Rect.Min.Y)*im.Stride + (x-im.Rect.Min.X)*4
}

// Window is a convenience function for painting the provided
// Image value to a new window, destroying the pixmap created by that image,
// and returning the window value.
// The window is sized to the dimensions of the image.
func (im *Image) Window(parent xproto.Window) *xwindow.Window {
	win := xwindow.Must(xwindow.Create(im.X, parent))
	win.Resize(im.Bounds().Dx(), im.Bounds().Dy())

	im.XSurfaceSet(win.Id)
	im.XDraw()
	im.XPaint(win.Id)
	im.Destroy()

	return win
}

// BGRA is the representation of color for each pixel in an X pixmap.
// BUG(burntsushi): This is hard-coded when it shouldn't be.
type BGRA struct {
	B, G, R, A uint8
}

// RGBA satisfies the color.Color interface.
func (c BGRA) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8

	g = uint32(c.G)
	g |= g << 8

	b = uint32(c.B)
	b |= b << 8

	a = uint32(c.A)
	a |= a << 8

	return
}

// bgraModel converts from any color to a BGRA color type.
func bgraModel(c color.Color) color.Color {
	if _, ok := c.(BGRA); ok {
		return c
	}

	r, g, b, a := c.RGBA()
	return BGRA{
		B: uint8(b >> 8),
		G: uint8(g >> 8),
		R: uint8(r >> 8),
		A: uint8(a >> 8),
	}
}
