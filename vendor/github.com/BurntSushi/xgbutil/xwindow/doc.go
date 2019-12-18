/*
Package xwindow defines a window type that provides easy access to common
window operations while hiding many of the more obscure X parameters. Examples
of such window operations include, but are not limited to, creating a window,
mapping a window, moving/resizing a window and getting the geometry of a
top-level client window including the window manager's decorations.

New and Generate functions are provided as constructors. New should be used
when you already have a window id, and it cannot fail. Generate should be used
when you need to allocate a new window identifier. Since allocating a new
window identifier can fail, an error could be returned.

Note that methods starting with 'WM' should only be used with a window manager
running that supports the EWMH specification. You should otherwise try to use
the corresponding methods without the 'WM' prefix.

A quick example

To create a window with a blue background that is 500 pixels wide and 200
pixels tall and map the window, use something like:

	win, err := xwindow.Generate(X)
	if err != nil {
		log.Fatal(err)
	}
	win.Create(X.RootWin(), 0, 0, 500, 200, xproto.CwBackPixel, 0x0000ff)
	win.Map()

You may also want to use CreateChecked instead of Create if you want to see if
there was an error when creating a window.

More examples

The xwindow package is used in many of the examples in the examples directory
of the xgbutil package. Of particular interest is window-name-sizes, which
prints the name and size of each top-level client window. (The geometry of the
window is found using DecorGeometry.)
*/
package xwindow
