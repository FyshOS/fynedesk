// +build linux,!ci

package wm

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app/util"
	"fyne.io/fyne/test"
	"fyne.io/fyne/theme"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"

	"fyne.io/desktop"
	wmTheme "fyne.io/desktop/theme"
)

type frame struct {
	c                           *client
	x, y                        int16
	restoreX, restoreY          int16
	width, height               uint16
	restoreWidth, restoreHeight uint16
	mouseX, mouseY              int16
	mouseResize                 bool

	framed              bool
	minWidth, minHeight uint

	canvas test.WindowlessCanvas
}

func (f *frame) unFrame() {
	if !f.framed {
		return
	}

	cli := f.c.wm.clientForWin(f.c.win)
	f.c.wm.RemoveWindow(cli)

	if cli == nil {
		return
	}

	if f != nil {
		xproto.ReparentWindow(f.c.wm.x.Conn(), f.c.win, f.c.wm.x.RootWin(), f.x, f.y)
	}
	xproto.UnmapWindow(f.c.wm.x.Conn(), f.c.id)
}

func (f *frame) press(x, y int16) {
	f.mouseX = x
	f.mouseY = y
	if x >= int16(f.width-f.c.wm.buttonWidth()) && y >= int16(f.height-f.c.wm.buttonWidth()) {
		f.mouseResize = true
	} else {
		f.mouseResize = false
	}

	f.c.wm.RaiseToTop(f.c)
}

func (f *frame) release(x, y int16) {
	ycoord := int16(f.c.wm.borderWidth() + f.c.wm.titleHeight())
	if x > int16(f.c.wm.borderWidth()) && x < int16(f.c.wm.borderWidth()+f.c.wm.buttonWidth()) &&
		y < ycoord {
		f.c.Close()
	} else if x > int16(f.c.wm.borderWidth())+int16(theme.Padding())+int16(f.c.wm.buttonWidth()) &&
		x < int16(f.c.wm.borderWidth())+int16(theme.Padding()*2)+int16(f.c.wm.buttonWidth()*2) &&
		y < ycoord {
		if f.c.Maximized() {
			f.c.Unmaximize()
		} else {
			f.c.Maximize()
		}
	} else if x > int16(f.c.wm.borderWidth())+int16(theme.Padding()*2)+int16(f.c.wm.buttonWidth()*2) &&
		x < int16(f.c.wm.borderWidth())+int16(theme.Padding()*2)+int16(f.c.wm.buttonWidth()*3) &&
		y < ycoord {
		f.c.Iconify()
	}
}

func (f *frame) drag(x, y int16) {
	deltaX := x - f.mouseX
	deltaY := y - f.mouseY
	var moveOnly = true

	if f.mouseResize {
		f.width = f.width + uint16(deltaX)
		f.height = f.height + uint16(deltaY)
		f.mouseX = x
		f.mouseY = y

		borderWidth := 2 * f.c.wm.borderWidth()
		if f.width < uint16(f.minWidth)+borderWidth {
			f.width = uint16(f.minWidth) + borderWidth
		}
		borderHeight := 2*f.c.wm.borderWidth() + f.c.wm.titleHeight()
		if f.height < uint16(f.minHeight)+borderHeight {
			f.height = uint16(f.minHeight) + borderHeight
		}
		moveOnly = false
	} else {
		f.x += deltaX
		f.y += deltaY
	}
	f.updateGeometry(moveOnly)
}

func (f *frame) motion(x, y int16) {
	cursor := defaultCursor
	if x > int16(f.c.wm.borderWidth()) && x <= int16(f.c.wm.borderWidth()+f.c.wm.buttonWidth()) &&
		y <= int16(f.c.wm.borderWidth()+f.c.wm.titleHeight()) {
		cursor = closeCursor
	} else if x >= int16(f.width-f.c.wm.buttonWidth()) && y >= int16(f.height-f.c.wm.buttonWidth()) {
		cursor = resizeCursor
	}

	err := xproto.ChangeWindowAttributesChecked(f.c.wm.x.Conn(), f.c.id, xproto.CwCursor,
		[]uint32{uint32(cursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (f *frame) updateGeometry(moveOnly bool) {
	if f.framed && !moveOnly {
		borderWidth := 2 * f.c.wm.borderWidth()
		if f.width < uint16(f.minWidth)+borderWidth {
			f.width = uint16(f.minWidth) + borderWidth
		}
		borderHeight := 2*f.c.wm.borderWidth() + f.c.wm.titleHeight()
		if f.height < uint16(f.minHeight)+borderHeight {
			f.height = uint16(f.minHeight) + borderHeight
		}
		err := xproto.ConfigureWindowChecked(f.c.wm.x.Conn(), f.c.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(f.c.wm.borderWidth()), uint32(f.c.wm.borderWidth() + f.c.wm.titleHeight()),
				uint32(f.width - borderWidth), uint32(f.height - borderHeight)}).Check()
		if err != nil {
			fyne.LogError("Configure Window Error", err)
		}
		f.applyTheme()
	}
	err := xproto.ConfigureWindowChecked(f.c.wm.x.Conn(), f.c.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (f *frame) maximize() {
	var w = f.c.wm.x.Screen().WidthInPixels
	var h = f.c.wm.x.Screen().HeightInPixels
	var x, y = 0, 0
	f.restoreWidth = f.width
	f.restoreHeight = f.height
	f.restoreX = f.x
	f.restoreY = f.y
	f.width = w
	f.height = h
	f.x = int16(x)
	f.y = int16(y)
	f.updateGeometry(false)
}

func (f *frame) unmaximize() {
	f.width = f.restoreWidth
	f.height = f.restoreHeight
	f.x = f.restoreX
	f.y = f.restoreY
	f.updateGeometry(false)
}

func (f *frame) applyTheme() {
	if !f.framed {
		return
	}

	depth := f.c.wm.x.Screen().RootDepth
	backR, backG, backB, _ := theme.BackgroundColor().RGBA()
	bgColor := backR<<16 | backG<<8 | backB

	if f.canvas == nil {
		f.canvas = util.NewSoftwareCanvas()
		f.canvas.SetPadded(false)
	}
	scale := desktop.Instance().Root().Canvas().Scale()
	f.canvas.SetScale(scale)
	border := newBorder(f.c)
	f.canvas.SetContent(border)

	scaledSize := fyne.NewSize(int(float32(f.width)/scale), int(float32(f.height)/scale))
	f.canvas.Resize(scaledSize)
	border.Resize(scaledSize)
	img := f.canvas.Capture()

	pid, err := xproto.NewPixmapId(f.c.wm.x.Conn())
	if err != nil {
		fyne.LogError("New Pixmap Error", err)
		return
	}
	xproto.CreatePixmap(f.c.wm.x.Conn(), depth, pid,
		xproto.Drawable(f.c.wm.x.Screen().Root), f.width, f.height)

	draw, _ := xproto.NewGcontextId(f.c.wm.x.Conn())
	xproto.CreateGC(f.c.wm.x.Conn(), draw, xproto.Drawable(pid), xproto.GcForeground, []uint32{bgColor})

	rect := xproto.Rectangle{X: 0, Y: 0, Width: f.width, Height: f.height}
	xproto.PolyFillRectangleChecked(f.c.wm.x.Conn(), xproto.Drawable(pid), draw, []xproto.Rectangle{rect})

	// DATA is BGRx
	width, height := uint32(f.width), uint32(float32(wmTheme.TitleHeight)*scale+float32(wmTheme.BorderWidth)*scale)
	data := make([]byte, width*height*4)
	i := uint32(0)
	for y := uint32(0); y < height; y++ {
		for x := uint32(0); x < width; x++ {
			r, g, b, _ := img.At(int(x), int(y)).RGBA()

			data[i] = byte(b)
			data[i+1] = byte(g)
			data[i+2] = byte(r)
			data[i+3] = 0

			i += 4
		}

	}

	xproto.PutImageChecked(f.c.wm.x.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(pid), draw,
		uint16(width), uint16(height), 0, 0, 0, depth, data)

	err = xproto.ChangeWindowAttributesChecked(f.c.wm.x.Conn(), f.c.id,
		xproto.CwBackPixmap, []uint32{uint32(pid)}).Check()
	if err != nil {
		fyne.LogError("Change Attribute Error", err)
		err = xproto.ChangeWindowAttributesChecked(f.c.wm.x.Conn(), f.c.id,
			xproto.CwBackPixmap, []uint32{0}).Check()
		if err != nil {
			fyne.LogError("Change Attribute Error", err)
		}
	}

	xproto.FreePixmap(f.c.wm.x.Conn(), pid)
}

func newFrame(c *client) *frame {
	attrs, err := xproto.GetGeometry(c.wm.x.Conn(), xproto.Drawable(c.win)).Reply()
	if err != nil {
		fyne.LogError("Get Geometry Error", err)
		return nil
	}

	f, err := xwindow.Generate(c.wm.x)
	if err != nil {
		fyne.LogError("Generate Window Error", err)
		return nil
	}

	scale := desktop.Instance().Root().Canvas().Scale()
	r, g, b, _ := theme.BackgroundColor().RGBA()
	values := []uint32{r<<16 | g<<8 | b, xproto.EventMaskStructureNotify | xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect | xproto.EventMaskExposure |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease | xproto.EventMaskButtonMotion |
		xproto.EventMaskKeyPress | xproto.EventMaskPointerMotion | xproto.EventMaskFocusChange}
	err = xproto.CreateWindowChecked(c.wm.x.Conn(), c.wm.x.Screen().RootDepth, f.Id, c.wm.x.RootWin(),
		attrs.X, attrs.Y, attrs.Width+uint16(float32(c.wm.borderWidth())/scale)*2, attrs.Height+uint16(float32(c.wm.borderWidth())/scale)*2+c.wm.titleHeight(), 0, xproto.WindowClassInputOutput,
		c.wm.x.Screen().RootVisual, xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		fyne.LogError("Create Window Error", err)
		return nil
	}
	c.id = f.Id
	framed := &frame{c: c, x: attrs.X, y: attrs.Y, width: attrs.Width + uint16(float32(c.wm.borderWidth())*2/scale),
		height: attrs.Height + uint16(float32(c.wm.borderWidth())/scale*2) + uint16(float32(c.wm.titleHeight())/scale), framed: true}
	framed.applyTheme()

	f.Map()
	xproto.ChangeSaveSet(c.wm.x.Conn(), xproto.SetModeInsert, c.win)
	xproto.ReparentWindow(c.wm.x.Conn(), c.win, c.id, int16(c.wm.borderWidth()-1), int16(c.wm.borderWidth()+c.wm.titleHeight()-1))
	xproto.MapWindow(c.wm.x.Conn(), c.win)

	return framed
}

func newFrameBorderless(c *client) *frame {
	attrs, err := xproto.GetGeometry(c.wm.x.Conn(), xproto.Drawable(c.win)).Reply()
	if err != nil {
		fyne.LogError("Get Geometry Error", err)
		return nil
	}

	c.id = c.win
	return &frame{c: c, x: attrs.X, y: attrs.Y, framed: false}
}
