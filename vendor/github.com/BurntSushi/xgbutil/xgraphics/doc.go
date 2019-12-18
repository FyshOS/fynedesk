/*
Package xgraphics defines an X image type and provides convenience functions
for reading and writing X pixmaps and bitmaps. It is a work-in-progress, and
while it works for some common X server configurations, it does not work for
all X server configurations. Package xgraphics also provides some support for
drawing text on to images using freetype-go, scaling images using graphics-go,
simple alpha blending, finding EWMH and ICCCM window icons and efficiently
drawing any image into an X pixmap. (Where "efficient" means being able to
specify sub-regions of images to draw, so that the entire image isn't sent to
X.) If more elaborate image routines are required, I recommend using draw2d.
(The xgraphics.Image type satisfies the draw.Image interface, which allows it
to work with draw2d.)

In general, xgraphics paints pixmaps to windows using using the BackPixmap
approach. (Setting the background pixmap of the window to the pixmap containing
your image, and clearing the window's background when the pixmap is updated.)
It also provides experimental support for another mechanism: copying the
contents of your image's pixmap directly to the window. (This requires
responding to expose events to redraw the pixmap.) The former approach requires
less book-keeping, but supposedly has some issues with some video cards. The
latter approach is probably more reliable, but requires more book-keeping.

Note that while text drawing functions are provided, it is not necessary to use
them to write text on images. Namely, there is nothing X specific about them.
They are strictly for convenience.

A quick example

This is a simple example the converts any value satisfying the image.Image
interface into an *xgraphics.Image value, and creates a new window with that
image painted in the window. (The XShow function probably doesn't have any
practical applications outside serving as an example, but can be useful for
debugging what an image looks like.)

	imgFile, err := os.Open(imgPath)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Fatal(err)
	}

	ximg := xgraphics.NewConvert(XUtilValue, img)
	ximg.XShow()

A complete working example named 'show-image' that's similar to this can be
found in the examples directory of the xgbutil package. More involved examples,
'show-window-icons' and 'pointer-painting', are also provided.

Portability

The xgraphics package *assumes* a particular kind of X server configuration.
Namely, this configuration specifies bits per pixel, image byte order, bitmap
bit order, scanline padding and unit length, image depth and so on. Handling
all of the possible values for each configuration option will greatly inflate
the code, but is on the TODO list.

I am undecided (perhaps because I haven't thought about it too much) about
whether to hide these configuration details behind multiple xgraphics.Image
types or hiding everything inside one xgraphics.Image type. I lean toward the
latter because the former requires a large number of types (and therefore a lot
of code duplication). One design decision that I've already made is that images
should be converted to the format used by the X server (xgraphics currently
assumes this is BGRx) once when the image is created. Without this, an
xgraphics.Image type wouldn't be required, and images would have to be
converted to the X image format every time an image is drawn into a pixmap.
This results in a lot of overhead. Moreover, Go's interfaces allow an
xgraphics.Image type to work anywhere that an image.Image or a draw.Image value
is expected.

The obvious down-side to this approach is that optimizations made in image
drawing routines in other libraries won't be able to apply to xgraphics.Image
values (since the optimizations are probably hard-coded for image types
declared in Go's standard library). This isn't well suited to the process of
creating some canvas to draw on, and using another library to draw on the
canvas. (At least, it won't be as fast as possible.) I can't think of any way
around this, other than having the library add an optimization step
specifically for xgraphics.Image values. Of course, the other approach is to
convert image formats only when drawing to X and completely subvert the
xgraphics.Image type, but this seems worse than unoptimized image drawing
routines. (Unfortunately, both things need to be fast.)

If your X server is not configured to what the xgraphics package expects,
messages will be emitted to stderr when a new xgraphics.Image value is created.
If you see any of these messages, please report them to xgbutil's project page:
https://github.com/BurntSushi/xgbutil.
*/
package xgraphics
