// +build linux,!ci

package wm

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"image"

	"fyne.io/fyne"
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

	minWidth, minHeight       uint
	borderTop, borderTopRight xproto.Pixmap
	borderTopWidth            uint16

	client *client
}

func (f *frame) unFrame() {
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
	if !f.client.Focused() {
		f.client.RaiseToTop()
		f.client.Focus()
		return
	}
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

	f.resizeBottom = false
	f.resizeLeft = false
	f.resizeRight = false
	f.updateGeometry(f.x, f.y, f.width, f.height, false)
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
		f.updateGeometry(x, f.y, width, height, false)
	} else {
		f.updateGeometry(f.x+deltaX, f.y+deltaY, f.width, f.height, false)
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

func (f *frame) getInnerWindowCoordinates(x int16, y int16, w uint16, h uint16) (uint32, uint32, uint32, uint32) {
	if f.client.Fullscreened() || !f.client.Decorated() {
		f.width = w
		f.height = h
		return uint32(0), uint32(0), uint32(w), uint32(h)
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

func (f *frame) updateGeometry(x, y int16, w, h uint16, force bool) {
	if !force {
		resize := w != f.width || h != f.height
		move := x != f.x || y != f.y
		if !move && !resize {
			return
		}
	}

	f.x = x
	f.y = y

	var newx, newy, neww, newh uint32
	newx, newy, neww, newh = f.getInnerWindowCoordinates(x, y, w, h)
	f.applyTheme(false)

	err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(newx), uint32(newy), neww, newh}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	err = xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (f *frame) maximizeApply() {
	f.client.restoreWidth = f.width
	f.client.restoreHeight = f.height
	f.client.restoreX = f.x
	f.client.restoreY = f.y

	maxWidth, maxHeight := desktop.Instance().ContentSizePixels()
	if f.client.Fullscreened() {
		maxWidth = uint32(f.client.wm.x.Screen().WidthInPixels)
		maxHeight = uint32(f.client.wm.x.Screen().HeightInPixels)
	}
	f.updateGeometry(0, 0, uint16(maxWidth), uint16(maxHeight), true)
}

func (f *frame) unmaximizeApply() {
	f.updateGeometry(f.client.restoreX, f.client.restoreY, f.client.restoreWidth, f.client.restoreHeight, true)
}

func (f *frame) drawDecoration(pidTop xproto.Pixmap, drawTop xproto.Gcontext, pidTopRight xproto.Pixmap, drawTopRight xproto.Gcontext, depth byte) {
	canvas := playground.NewSoftwareCanvas()
	canvas.SetPadded(false)
	canvas.SetContent(newBorder(f.client))

	scale := desktop.Instance().Root().Canvas().Scale()
	canvas.SetScale(scale)

	heightPix := f.titleHeight()
	iconBorderPixWidth := heightPix + f.borderWidth()*2
	widthPix := f.borderTopWidth + iconBorderPixWidth
	canvas.Resize(fyne.NewSize(int(float32(widthPix)/scale), wmTheme.TitleHeight))
	img := canvas.Capture()

	// TODO just copy the label minSize - smallest possible but maybe bigger than window width
	// Draw in pixel rows so we don't overflow count usable by PutImageChecked
	for i := uint16(0); i < heightPix; i++ {
		f.copyDecorationPixels(uint32(f.borderTopWidth), 1, 0, uint32(i), img, pidTop, drawTop, depth)
	}

	f.copyDecorationPixels(uint32(iconBorderPixWidth), uint32(heightPix), uint32(f.borderTopWidth), 0, img, pidTopRight, drawTopRight, depth)
}

func (f *frame) copyDecorationPixels(width, height, xoff, yoff uint32, img image.Image, pid xproto.Pixmap, draw xproto.Gcontext, depth byte) {
	// DATA is BGRx
	data := make([]byte, width*height*4)
	i := uint32(0)
	for y := uint32(0); y < height; y++ {
		for x := uint32(0); x < width; x++ {
			r, g, b, _ := img.At(int(xoff+x), int(yoff+y)).RGBA()

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

func (f *frame) createPixmaps(depth byte) error {
	iconPix := f.titleHeight()
	iconAndBorderPix := iconPix + f.borderWidth()*2
	f.borderTopWidth = f.width - iconAndBorderPix

	pid, err := xproto.NewPixmapId(f.client.wm.x.Conn())
	if err != nil {
		return err
	}

	xproto.CreatePixmap(f.client.wm.x.Conn(), depth, pid,
		xproto.Drawable(f.client.wm.x.Screen().Root), f.borderTopWidth, iconPix)
	f.borderTop = pid

	pid, err = xproto.NewPixmapId(f.client.wm.x.Conn())
	if err != nil {
		return err
	}

	xproto.CreatePixmap(f.client.wm.x.Conn(), depth, pid,
		xproto.Drawable(f.client.wm.x.Screen().Root), iconAndBorderPix, iconPix)
	f.borderTopRight = pid

	return nil
}

func (f *frame) decorate(force bool) {
	depth := f.client.wm.x.Screen().RootDepth
	refresh := force

	if f.borderTop == 0 {
		err := f.createPixmaps(depth)
		if err != nil {
			fyne.LogError("New Pixmap Error", err)
			return
		}
		refresh = true
	}

	backR, backG, backB, _ := theme.BackgroundColor().RGBA()
	bgColor := backR<<16 | backG<<8 | backB

	drawTop, _ := xproto.NewGcontextId(f.client.wm.x.Conn())
	xproto.CreateGC(f.client.wm.x.Conn(), drawTop, xproto.Drawable(f.borderTop), xproto.GcForeground, []uint32{bgColor})
	drawTopRight, _ := xproto.NewGcontextId(f.client.wm.x.Conn())
	xproto.CreateGC(f.client.wm.x.Conn(), drawTopRight, xproto.Drawable(f.borderTopRight), xproto.GcForeground, []uint32{bgColor})

	if refresh {
		f.drawDecoration(f.borderTop, drawTop, f.borderTopRight, drawTopRight, depth)
	}

	draw, _ := xproto.NewGcontextId(f.client.wm.x.Conn())
	xproto.CreateGC(f.client.wm.x.Conn(), draw, xproto.Drawable(f.client.id), xproto.GcForeground, []uint32{bgColor})
	rect := xproto.Rectangle{X: 0, Y: 0, Width: f.width, Height: f.height}
	xproto.PolyFillRectangleChecked(f.client.wm.x.Conn(), xproto.Drawable(f.client.id), draw, []xproto.Rectangle{rect})

	iconSizePix := f.titleHeight()
	iconAndBorderSizePix := iconSizePix + f.borderWidth()*2
	xproto.CopyArea(f.client.wm.x.Conn(), xproto.Drawable(f.borderTop), xproto.Drawable(f.client.id), drawTop,
		0, 0, 0, 0, f.borderTopWidth, uint16(iconSizePix))
	xproto.CopyArea(f.client.wm.x.Conn(), xproto.Drawable(f.borderTopRight), xproto.Drawable(f.client.id), drawTopRight,
		0, 0, int16(f.width-iconAndBorderSizePix), 0, uint16(iconAndBorderSizePix), uint16(iconSizePix))
}

func (f *frame) applyBorderlessTheme() {
	if !f.client.Decorated() {
		backR, backG, backB, _ := theme.BackgroundColor().RGBA()

		bgColor := backR<<16 | backG<<8 | backB
		draw, _ := xproto.NewGcontextId(f.client.wm.x.Conn())
		xproto.CreateGC(f.client.wm.x.Conn(), draw, xproto.Drawable(f.client.id), xproto.GcForeground, []uint32{bgColor})
		rect := xproto.Rectangle{X: 0, Y: 0, Width: f.width, Height: f.height}
		xproto.PolyFillRectangleChecked(f.client.wm.x.Conn(), xproto.Drawable(f.client.id), draw, []xproto.Rectangle{rect})
	}
	if !f.client.Fullscreened() {
		return
	}
	err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (f *frame) applyTheme(force bool) {
	if !f.client.Fullscreened() && f.client.Decorated() {
		f.decorate(force)
		return
	}
	f.applyBorderlessTheme()
}

func (f *frame) updateTitle() {
	f.applyTheme(true)
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

func (f *frame) hide() {
	xproto.ReparentWindow(f.client.wm.x.Conn(), f.client.win, f.client.wm.x.RootWin(), f.x, f.y)
	xproto.UnmapWindow(f.client.wm.x.Conn(), f.client.win)
}

func (f *frame) addBorder() {
	f.applyTheme(true)
	x := int16(f.borderWidth())
	y := int16(f.titleHeight())
	w := f.width + uint16(f.borderWidth()*2)
	h := f.height + uint16(f.borderWidth()) + f.titleHeight()
	err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(x), uint32(y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	err = xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(w), uint32(h)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	f.width = w
	f.height = h
	ewmh.FrameExtentsSet(f.client.wm.x, f.client.win, &ewmh.FrameExtents{Left: int(f.borderWidth()), Right: int(f.borderWidth()), Top: int(f.titleHeight()), Bottom: int(f.borderWidth())})
}

func (f *frame) removeBorder() {
	f.applyTheme(true)
	err := xproto.ConfigureWindowChecked(f.client.wm.x.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{0, 0, uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	ewmh.FrameExtentsSet(f.client.wm.x, f.client.win, &ewmh.FrameExtents{Left: 0, Right: 0, Top: 0, Bottom: 0})
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
	x, y, w, h := attrs.X, attrs.Y, attrs.Width, attrs.Height
	full := c.Fullscreened()
	decorated := c.Decorated()
	if full {
		x = 0
		y = 0
		w = c.wm.x.Screen().WidthInPixels
		h = c.wm.x.Screen().HeightInPixels
	} else if !decorated {
		x = attrs.X
		y = attrs.Y
		w = attrs.Width
		h = attrs.Height
	}
	framed := &frame{client: c, x: attrs.X, y: attrs.Y}
	if !full && decorated {
		w = attrs.Width + uint16(framed.borderWidth()*2)
		h = attrs.Height + uint16(framed.borderWidth()) + framed.titleHeight()
	}
	values := []uint32{xproto.EventMaskStructureNotify | xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect | xproto.EventMaskExposure |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease | xproto.EventMaskButtonMotion |
		xproto.EventMaskKeyPress | xproto.EventMaskPointerMotion | xproto.EventMaskFocusChange | xproto.EventMaskPropertyChange}
	err = xproto.CreateWindowChecked(c.wm.x.Conn(), c.wm.x.Screen().RootDepth, f.Id, c.wm.x.RootWin(),
		x, y, w, h, 0, xproto.WindowClassInputOutput, c.wm.x.Screen().RootVisual,
		xproto.CwEventMask, values).Check()
	if err != nil {
		fyne.LogError("Create Window Error", err)
		return nil
	}
	c.id = f.Id

	framed.width = w
	framed.height = h
	if !full && decorated {
		xproto.ReparentWindow(c.wm.x.Conn(), c.win, c.id, int16(framed.borderWidth()), int16(framed.titleHeight()))
		ewmh.FrameExtentsSet(c.wm.x, c.win, &ewmh.FrameExtents{Left: int(framed.borderWidth()), Right: int(framed.borderWidth()), Top: int(framed.titleHeight()), Bottom: int(framed.borderWidth())})
	} else {
		xproto.ReparentWindow(c.wm.x.Conn(), c.win, c.id, 0, 0)
		ewmh.FrameExtentsSet(c.wm.x, c.win, &ewmh.FrameExtents{Left: 0, Right: 0, Top: 0, Bottom: 0})
	}
	framed.show()
	framed.applyTheme(true)

	return framed
}
