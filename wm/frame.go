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
	x, y           int16
	width, height  uint16
	mouseX, mouseY int16
	mouseResize    bool
	framed         bool

	minWidth, minHeight uint

	client *client
	canvas test.WindowlessCanvas
}

func (f *frame) unFrame() {
	if !f.framed {
		return
	}

	c := f.client.wm.clientForWin(f.client.win)
	f.client.wm.RemoveWindow(c)

	if c == nil {
		return
	}

	if f != nil {
		xproto.ReparentWindow(f.client.wm.x.Conn(), f.client.win, f.client.wm.x.RootWin(), f.x, f.y)
	}
	xproto.UnmapWindow(f.client.wm.x.Conn(), f.client.id)
}

func (f *frame) press(x, y int16) {
	f.mouseX = x
	f.mouseY = y

	relX := x - f.x
	relY := y - f.y
	if relX >= int16(f.width-f.buttonWidth()) && relY >= int16(f.height-f.buttonWidth()) {
		f.mouseResize = true
	} else {
		f.mouseResize = false
	}

	f.client.wm.RaiseToTop(f.client)
}

func (f *frame) release(x, y int16) {
	relX := x - f.x
	relY := y - f.y
	barYMax := int16(f.borderWidth() + f.titleHeight())
	if relY > barYMax {
		return
	}
	if relX > int16(f.borderWidth()) && relX < int16(f.borderWidth()+f.buttonWidth()) {
		f.client.Close()
	} else if relX > int16(f.borderWidth())+int16(theme.Padding())+int16(f.buttonWidth()) &&
		relX < int16(f.borderWidth())+int16(theme.Padding()*2)+int16(f.buttonWidth()*2) {
		if f.client.Maximized() {
			f.client.Unmaximize()
		} else {
			f.client.Maximize()
		}
	} else if relX > int16(f.borderWidth())+int16(theme.Padding()*2)+int16(f.buttonWidth()*2) &&
		relX < int16(f.borderWidth())+int16(theme.Padding()*2)+int16(f.buttonWidth()*3) {
		f.client.Iconify()
	}
}

func (f *frame) drag(x, y int16) {
	deltaX := x - f.mouseX
	deltaY := y - f.mouseY
	f.mouseX = x
	f.mouseY = y
	if deltaX == 0 && deltaY == 0 {
		return
	}
	var moveOnly = true

	if f.mouseResize {
		f.width += uint16(deltaX)
		f.height += uint16(deltaY)

		borderWidth := 2 * f.borderWidth()
		if f.width < uint16(f.minWidth)+borderWidth {
			f.width = uint16(f.minWidth) + borderWidth
		}
		borderHeight := 2*f.borderWidth() + f.titleHeight()
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
	relX := x - f.x
	relY := y - f.y
	cursor := defaultCursor
	if relX > int16(f.borderWidth()) && relX <= int16(f.borderWidth()+f.buttonWidth()) &&
		relY <= int16(f.borderWidth()+f.titleHeight()) {
		cursor = closeCursor
	} else if relX >= int16(f.width-f.buttonWidth()) && relY >= int16(f.height-f.buttonWidth()) {
		cursor = resizeCursor
	}

	err := xproto.ChangeWindowAttributesChecked(f.client.wm.x.Conn(), f.client.id, xproto.CwCursor,
		[]uint32{uint32(cursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (f *frame) updateGeometry(moveOnly bool) {
	if f.framed && !moveOnly {
		borderWidth := 2 * f.borderWidth()
		if f.width < uint16(f.minWidth)+borderWidth {
			f.width = uint16(f.minWidth) + borderWidth
		}
		borderHeight := 2*f.borderWidth() + f.titleHeight()
		if f.height < uint16(f.minHeight)+borderHeight {
			f.height = uint16(f.minHeight) + borderHeight
		}
		err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(f.borderWidth()), uint32(f.borderWidth() + f.titleHeight()),
				uint32(f.width - borderWidth), uint32(f.height - borderHeight)}).Check()
		if err != nil {
			fyne.LogError("Configure Window Error", err)
		}
		f.applyTheme()
	}
	err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (f *frame) maximize() {
	var w = f.client.wm.x.Screen().WidthInPixels
	var h = f.client.wm.x.Screen().HeightInPixels
	var x, y = 0, 0
	f.client.restoreWidth = f.width
	f.client.restoreHeight = f.height
	f.client.restoreX = f.x
	f.client.restoreY = f.y
	f.width = w
	f.height = h
	f.x = int16(x)
	f.y = int16(y)
	f.updateGeometry(false)
}

func (f *frame) unmaximize() {
	f.width = f.client.restoreWidth
	f.height = f.client.restoreHeight
	f.x = f.client.restoreX
	f.y = f.client.restoreY
	f.updateGeometry(false)
}

func (f *frame) applyTheme() {
	if !f.framed {
		return
	}

	depth := f.client.wm.x.Screen().RootDepth
	backR, backG, backB, _ := theme.BackgroundColor().RGBA()
	bgColor := backR<<16 | backG<<8 | backB

	if f.canvas == nil {
		f.canvas = util.NewSoftwareCanvas()
		f.canvas.SetPadded(false)
	}
	scale := desktop.Instance().Root().Canvas().Scale()
	f.canvas.SetScale(scale)
	border := newBorder(f.client)
	f.canvas.SetContent(border)

	scaledSize := fyne.NewSize(int(float32(f.width)/scale), int(float32(f.height)/scale))
	f.canvas.Resize(scaledSize)
	border.Resize(scaledSize)
	img := f.canvas.Capture()

	pid, err := xproto.NewPixmapId(f.client.wm.x.Conn())
	if err != nil {
		fyne.LogError("New Pixmap Error", err)
		return
	}
	xproto.CreatePixmap(f.client.wm.x.Conn(), depth, pid,
		xproto.Drawable(f.client.wm.x.Screen().Root), f.width, f.height)

	draw, _ := xproto.NewGcontextId(f.client.wm.x.Conn())
	xproto.CreateGC(f.client.wm.x.Conn(), draw, xproto.Drawable(pid), xproto.GcForeground, []uint32{bgColor})

	rect := xproto.Rectangle{X: 0, Y: 0, Width: f.width, Height: f.height}
	xproto.PolyFillRectangleChecked(f.client.wm.x.Conn(), xproto.Drawable(pid), draw, []xproto.Rectangle{rect})

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

	xproto.PutImageChecked(f.client.wm.x.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(pid), draw,
		uint16(width), uint16(height), 0, 0, 0, depth, data)

	err = xproto.ChangeWindowAttributesChecked(f.client.wm.x.Conn(), f.client.id,
		xproto.CwBackPixmap, []uint32{uint32(pid)}).Check()
	if err != nil {
		fyne.LogError("Change Attribute Error", err)
		err = xproto.ChangeWindowAttributesChecked(f.client.wm.x.Conn(), f.client.id,
			xproto.CwBackPixmap, []uint32{0}).Check()
		if err != nil {
			fyne.LogError("Change Attribute Error", err)
		}
	}

	xproto.FreePixmap(f.client.wm.x.Conn(), pid)
}

func (f *frame) borderWidth() uint16 {
	return f.client.wm.scaleToPixels(wmTheme.BorderWidth)
}

func (f *frame) buttonWidth() uint16 {
	return f.client.wm.scaleToPixels(wmTheme.ButtonWidth)
}

func (f *frame) titleHeight() uint16 {
	return f.client.wm.scaleToPixels(wmTheme.TitleHeight)
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

	framed := &frame{client: c, x: attrs.X, y: attrs.Y, framed: true}
	r, g, b, _ := theme.BackgroundColor().RGBA()
	values := []uint32{r<<16 | g<<8 | b, xproto.EventMaskStructureNotify | xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect | xproto.EventMaskExposure |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease | xproto.EventMaskButtonMotion |
		xproto.EventMaskKeyPress | xproto.EventMaskPointerMotion | xproto.EventMaskFocusChange}
	err = xproto.CreateWindowChecked(c.wm.x.Conn(), c.wm.x.Screen().RootDepth, f.Id, c.wm.x.RootWin(),
		attrs.X, attrs.Y, attrs.Width+uint16(framed.borderWidth()*2),
		attrs.Height+uint16(framed.borderWidth())*2+framed.titleHeight(),
		0, xproto.WindowClassInputOutput, c.wm.x.Screen().RootVisual,
		xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		fyne.LogError("Create Window Error", err)
		return nil
	}
	c.id = f.Id

	framed.width = attrs.Width + uint16(framed.borderWidth()*2)
	framed.height = attrs.Height + uint16(framed.borderWidth()*2) + uint16(framed.titleHeight())
	framed.applyTheme()

	f.Map()
	xproto.ChangeSaveSet(c.wm.x.Conn(), xproto.SetModeInsert, c.win)
	xproto.ReparentWindow(c.wm.x.Conn(), c.win, c.id, int16(framed.borderWidth()-1), int16(framed.borderWidth()+framed.titleHeight()-1))
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
	return &frame{client: c, x: attrs.X, y: attrs.Y, framed: false}
}
