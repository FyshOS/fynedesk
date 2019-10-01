// +build linux,!ci

package wm

import (
	"image"

	"fyne.io/fyne"
	"fyne.io/fyne/test"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/tools/playground"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"

	"fyne.io/desktop"
	wmTheme "fyne.io/desktop/theme"
)

type frame struct {
	x, y                    int16
	width, height           uint16
	mouseX, mouseY          int16
	resizeBottom            bool
	resizeLeft, resizeRight bool
	framed                  bool

	minWidth, minHeight uint

	client *client
	canvas test.WindowlessCanvas
}

func (f *frame) unFrame() {
	if !f.framed {
		return
	}

	cli := f.client.wm.clientForWin(f.client.win)
	if cli == nil {
		return
	}
	c := cli.(*client)
	c.wm.RemoveWindow(c)

	if f != nil {
		xproto.ReparentWindow(c.wm.x.Conn(), c.win, c.wm.x.RootWin(), f.x, f.y)
	}
	xproto.ChangeSaveSet(c.wm.x.Conn(), xproto.SetModeDelete, c.win)
	xproto.UnmapWindow(c.wm.x.Conn(), c.id)
}

func (f *frame) press(x, y int16) {
	f.mouseX = x
	f.mouseY = y

	relX := x - f.x
	relY := y - f.y
	f.resizeBottom = false
	f.resizeLeft = false
	f.resizeRight = false
	if relY >= int16(f.titleHeight()) {
		if relY >= int16(f.height-f.buttonWidth()) {
			f.resizeBottom = true
		}
		if relX < int16(f.buttonWidth()) {
			f.resizeLeft = true
		} else if relX >= int16(f.width-f.buttonWidth()) {
			f.resizeRight = true
		}
	}

	f.client.wm.RaiseToTop(f.client)
}

func (f *frame) release(x, y int16) {
	relX := x - f.x
	relY := y - f.y
	barYMax := int16(f.titleHeight())
	if relY > barYMax {
		return
	}
	if relX >= int16(f.borderWidth()) && relX < int16(f.borderWidth()+f.buttonWidth()) {
		f.client.Close()
	} else if relX >= int16(f.borderWidth())+int16(theme.Padding())+int16(f.buttonWidth()) &&
		relX < int16(f.borderWidth())+int16(theme.Padding()*2)+int16(f.buttonWidth()*2) {
		if f.client.Maximized() {
			f.client.Unmaximize()
		} else {
			f.client.Maximize()
		}
	} else if relX >= int16(f.borderWidth())+int16(theme.Padding()*2)+int16(f.buttonWidth()*2) &&
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

	if f.resizeBottom || f.resizeLeft || f.resizeRight {
		x := f.x
		width := f.width
		height := f.height
		if f.resizeBottom {
			height += uint16(deltaY)
		}
		if f.resizeLeft {
			x += int16(deltaX)
			width -= uint16(deltaX)
		} else if f.resizeRight {
			width += uint16(deltaX)
		}
		f.updateGeometry(x, f.y, width, height)
	} else {
		f.updateGeometry(f.x+deltaX, f.y+deltaY, f.width, f.height)
	}
}

func (f *frame) motion(x, y int16) {
	relX := x - f.x
	relY := y - f.y
	cursor := defaultCursor
	if relY <= int16(f.titleHeight()) { // title bar
		if relX > int16(f.borderWidth()) && relX <= int16(f.borderWidth()+f.buttonWidth()) {
			cursor = closeCursor
		}
	} else if relY >= int16(f.height-f.buttonWidth()) { // bottom
		if relX < int16(f.buttonWidth()) {
			cursor = resizeBottomLeftCursor
		} else if relX >= int16(f.width-f.buttonWidth()) {
			cursor = resizeBottomRightCursor
		} else {
			cursor = resizeBottomCursor
		}
	} else { // center (sides)
		if relX < int16(f.width-f.buttonWidth()) {
			cursor = resizeLeftCursor
		} else if relX >= int16(f.width-f.buttonWidth()) {
			cursor = resizeRightCursor
		}
	}

	err := xproto.ChangeWindowAttributesChecked(f.client.wm.x.Conn(), f.client.id, xproto.CwCursor,
		[]uint32{uint32(cursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (f *frame) getInnerWindowCoordinates(x int16, y int16, w uint16, h uint16, decorated bool) (uint32, uint32, uint32, uint32) {
	if !decorated {
		f.width = w
		f.height = h
		return uint32(x), uint32(y), uint32(w), uint32(h)
	}
	borderWidth := 2 * f.borderWidth()
	if w < uint16(f.minWidth)+borderWidth {
		w = uint16(f.minWidth) + borderWidth
	}
	borderHeight := f.borderWidth() + f.titleHeight()
	if h < uint16(f.minHeight)+borderHeight {
		h = uint16(f.minHeight) + borderHeight
	}
	f.width = w
	f.height = h

	return uint32(f.borderWidth()), uint32(f.titleHeight()),
		uint32(f.width - borderWidth), uint32(f.height - borderHeight)
}

func (f *frame) getGeometry() (int16, int16, uint16, uint16) {
	return f.x, f.y, f.width, f.height
}

func (f *frame) updateGeometry(x, y int16, w, h uint16) {
	resize := w != f.width || h != f.height
	move := x != f.x || y != f.y
	if !move && !resize {
		return
	}
	f.x = x
	f.y = y
	if f.framed && resize {
		var newx, newy, neww, newh uint32
		newx, newy, neww, newh = f.getInnerWindowCoordinates(x, y, w, h, !f.client.Fullscreened())
		f.applyTheme()
		err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{newx, newy, neww, newh}).Check()
		if err != nil {
			fyne.LogError("Configure Window Error", err)
		}
	} else {
		f.width = w
		f.height = h
	}
	err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (f *frame) iconifyApply() {
	xproto.ReparentWindow(f.client.wm.x.Conn(), f.client.win, f.client.wm.x.RootWin(), f.x, f.y)
	xproto.UnmapWindow(f.client.wm.x.Conn(), f.client.win)
}

func (f *frame) uniconifyApply() {
	xproto.MapWindow(f.client.wm.x.Conn(), f.client.win)
}

func (f *frame) maximizeApply() {
	var w = f.client.wm.x.Screen().WidthInPixels
	var h = f.client.wm.x.Screen().HeightInPixels
	f.client.restoreWidth = f.width
	f.client.restoreHeight = f.height
	f.client.restoreX = f.x
	f.client.restoreY = f.y
	f.updateGeometry(0, 0, w, h)
}

func (f *frame) unmaximizeApply() {
	f.updateGeometry(f.client.restoreX, f.client.restoreY, f.client.restoreWidth, f.client.restoreHeight)
}

func (f *frame) drawDecoration(pid xproto.Pixmap, draw xproto.Gcontext, depth byte) {
	if f.canvas == nil {
		f.canvas = playground.NewSoftwareCanvas()
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

	// Draw in two passes so we don't overflow count usable by PutImageChecked
	mid := uint32(f.titleHeight() / 2)
	f.copyDecorationPixels(0, mid, img, pid, draw, depth)
	f.copyDecorationPixels(mid, uint32(f.titleHeight())-mid, img, pid, draw, depth)
}

func (f *frame) copyDecorationPixels(yoff, height uint32, img image.Image, pid xproto.Pixmap, draw xproto.Gcontext, depth byte) {
	// DATA is BGRx
	width, height := uint32(f.width), uint32(height)
	data := make([]byte, width*height*4)
	i := uint32(0)
	for y := uint32(0); y < height; y++ {
		for x := uint32(0); x < width; x++ {
			r, g, b, _ := img.At(int(x), int(yoff+y)).RGBA()

			data[i] = byte(b)
			data[i+1] = byte(g)
			data[i+2] = byte(r)
			data[i+3] = 0

			i += 4
		}
	}
	err := xproto.PutImageChecked(f.client.wm.x.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(pid), draw,
		uint16(width), uint16(height), 0, int16(yoff), 0, depth, data).Check()
	if err != nil {
		fyne.LogError("Put image error", err)
	}
}

func (f *frame) decorate() {
	pid, err := xproto.NewPixmapId(f.client.wm.x.Conn())
	if err != nil {
		fyne.LogError("New Pixmap Error", err)
		return
	}

	depth := f.client.wm.x.Screen().RootDepth
	backR, backG, backB, _ := theme.BackgroundColor().RGBA()
	bgColor := backR<<16 | backG<<8 | backB

	xproto.CreatePixmap(f.client.wm.x.Conn(), depth, pid,
		xproto.Drawable(f.client.wm.x.Screen().Root), f.width, f.height)
	draw, _ := xproto.NewGcontextId(f.client.wm.x.Conn())
	xproto.CreateGC(f.client.wm.x.Conn(), draw, xproto.Drawable(pid), xproto.GcForeground, []uint32{bgColor})
	rect := xproto.Rectangle{X: 0, Y: 0, Width: f.width, Height: f.height}
	xproto.PolyFillRectangleChecked(f.client.wm.x.Conn(), xproto.Drawable(pid), draw, []xproto.Rectangle{rect})
	f.drawDecoration(pid, draw, depth)
	err = xproto.ChangeWindowAttributesChecked(f.client.wm.x.Conn(), f.client.id,
		xproto.CwBackPixmap, []uint32{uint32(pid)}).Check()
	if err != nil {
		fyne.LogError("Change Attribute Error", err)
	}
	xproto.FreePixmap(f.client.wm.x.Conn(), pid)
}

func (f *frame) applyTheme() {
	if !f.framed {
		return
	}

	if !f.client.Fullscreened() {
		f.decorate()
	} else {
		err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
		if err != nil {
			fyne.LogError("Configure Window Error", err)
		}
	}
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

func (f *frame) show() {
	c := f.client
	xproto.MapWindow(c.wm.x.Conn(), c.id)

	xproto.ChangeSaveSet(c.wm.x.Conn(), xproto.SetModeInsert, c.win)
	xproto.MapWindow(c.wm.x.Conn(), c.win)
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
		attrs.Height+uint16(framed.borderWidth())+framed.titleHeight(),
		0, xproto.WindowClassInputOutput, c.wm.x.Screen().RootVisual,
		xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		fyne.LogError("Create Window Error", err)
		return nil
	}
	c.id = f.Id

	framed.width = attrs.Width + uint16(framed.borderWidth()*2)
	framed.height = attrs.Height + uint16(framed.borderWidth()) + uint16(framed.titleHeight())
	framed.applyTheme()

	xproto.ReparentWindow(c.wm.x.Conn(), c.win, c.id, int16(framed.borderWidth()-1), int16(framed.titleHeight()-1))
	framed.show()

	return framed
}

func newFrameBorderless(c *client) *frame {
	attrs, err := xproto.GetGeometry(c.wm.x.Conn(), xproto.Drawable(c.win)).Reply()
	if err != nil {
		fyne.LogError("Get Geometry Error", err)
		return nil
	}

	c.id = c.win
	unframed := &frame{client: c, x: attrs.X, y: attrs.Y,
		width: attrs.Width, height: attrs.Height, framed: false}
	unframed.show()

	return unframed
}
