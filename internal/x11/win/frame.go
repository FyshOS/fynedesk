// +build linux

package win

import (
	"image"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/tools/playground"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xwindow"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/x11"
	wmTheme "fyne.io/fynedesk/theme"
	"fyne.io/fynedesk/wm"
)

type frame struct {
	fynedesk.Geometry
	childWidth, childHeight             uint
	resizeStartWidth, resizeStartHeight uint
	mouseX, mouseY                      int
	resizeStartX, resizeStartY          int
	resizeBottom                        bool
	resizeLeft, resizeRight             bool
	moveOnly                            bool

	borderTop, borderTopRight xproto.Pixmap
	borderTopWidth            uint

	client *client
}

func newFrame(c *client) *frame {
	attrs, err := xproto.GetGeometry(c.wm.Conn(), xproto.Drawable(c.win)).Reply()
	if err != nil {
		fyne.LogError("Get Geometry Error", err)
		return nil
	}

	f, err := xwindow.Generate(c.wm.X())
	if err != nil {
		fyne.LogError("Generate Window Error", err)
		return nil
	}
	g := x11.GeometryFromGetGeometryReply(attrs)
	full := c.Fullscreened()
	decorated := c.Properties().Decorated()
	maximized := c.Maximized()
	screen := fynedesk.Instance().Screens().ScreenForGeometry(g)
	borderWidth := wm.ScaleToPixels(wmTheme.BorderWidth, screen)
	titleHeight := wm.ScaleToPixels(wmTheme.TitleHeight, screen)
	if full || maximized {
		activeHead := fynedesk.Instance().Screens().ScreenForGeometry(g)
		g.X = activeHead.X
		g.Y = activeHead.Y
		if full {
			g.Width = activeHead.Width
			g.Height = activeHead.Height
		} else {
			g.Width, g.Height = fynedesk.Instance().ContentSizePixels(activeHead)
		}
	} else if decorated {
		g.X -= borderWidth
		g.Y -= titleHeight
		if g.X < 0 {
			g.X = 0
		}
		if g.Y < 0 {
			g.Y = 0
		}
		if !maximized {
			g.Width = uint(attrs.Width) + uint(borderWidth*2)
			g.Height = uint(attrs.Height) + uint(borderWidth+titleHeight)
		}
	}
	framed := &frame{Geometry: g, client: c}
	values := []uint32{xproto.EventMaskStructureNotify | xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect | xproto.EventMaskExposure |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease | xproto.EventMaskButtonMotion |
		xproto.EventMaskKeyPress | xproto.EventMaskPointerMotion | xproto.EventMaskFocusChange |
		xproto.EventMaskPropertyChange}
	err = xproto.CreateWindowChecked(c.wm.Conn(), c.wm.X().Screen().RootDepth, f.Id, c.wm.X().RootWin(),
		int16(g.X), int16(g.Y), uint16(g.Width), uint16(g.Height), 0, xproto.WindowClassInputOutput, c.wm.X().Screen().RootVisual,
		xproto.CwEventMask, values).Check()
	if err != nil {
		fyne.LogError("Create Window Error", err)
		return nil
	}
	c.id = f.Id

	if full || !decorated {
		framed.childWidth = g.Width
		framed.childHeight = g.Height
	} else {
		framed.childWidth = g.Width - uint(borderWidth*2)
		framed.childHeight = g.Height - uint(borderWidth-titleHeight)
	}

	var offsetX, offsetY int16 = 0, 0
	if !full && decorated {
		offsetX = int16(borderWidth)
		offsetY = int16(titleHeight)
		xproto.ReparentWindow(c.wm.Conn(), c.win, c.id, int16(borderWidth), int16(titleHeight))
		ewmh.FrameExtentsSet(c.wm.X(), c.win, &ewmh.FrameExtents{Left: int(borderWidth), Right: int(borderWidth),
			Top: int(titleHeight), Bottom: int(borderWidth)})
	} else {
		xproto.ReparentWindow(c.wm.Conn(), c.win, c.id, attrs.X, attrs.Y)
		ewmh.FrameExtentsSet(c.wm.X(), c.win, &ewmh.FrameExtents{Left: 0, Right: 0, Top: 0, Bottom: 0})
	}

	if full || maximized {
		err = xproto.ConfigureWindowChecked(c.wm.Conn(), c.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(offsetX), uint32(offsetY), uint32(framed.childWidth), uint32(framed.childHeight)}).Check()
		if err != nil {
			fyne.LogError("Configure Window Error", err)
		}
	}

	windowStateSet(c.wm.X(), c.win, icccm.StateNormal)
	framed.show()
	framed.applyTheme(true)
	framed.notifyInnerGeometry()

	return framed
}

func (f *frame) addBorder() {
	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))
	x := int(borderWidth)
	y := int(titleHeight)
	g := f.Geometry
	if !f.client.maximized {
		g.X -= x
		g.Y -= y
		if g.X < 0 {
			g.X = 0
		}
		if g.Y < 0 {
			g.Y = 0
		}
		g.Width = f.childWidth + borderWidth*2
		g.Height = f.childHeight + borderWidth + titleHeight
	}
	f.applyTheme(true)

	err := xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(x), uint32(y), uint32(f.childWidth), uint32(f.childHeight)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	err = xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		x11.GeometryToUint32s(g)).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}

	ewmh.FrameExtentsSet(f.client.wm.X(), f.client.win, &ewmh.FrameExtents{Left: int(borderWidth), Right: int(borderWidth), Top: int(titleHeight), Bottom: int(borderWidth)})
	f.notifyInnerGeometry()
}

func (f *frame) applyBorderlessTheme() {
	err := xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(0), uint32(0), uint32(f.Width), uint32(f.Height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (f *frame) applyTheme(force bool) {
	if f.client.Fullscreened() || !f.client.Properties().Decorated() {
		f.applyBorderlessTheme()
		return
	}

	f.checkScale()
	f.decorate(force)
}

func (f *frame) checkScale() {
	titleHeight := x11.TitleHeight(x11.XWin(f.client))
	if f.Height-titleHeight != f.childHeight {
		f.updateGeometry(f.Geometry, true)
		f.notifyInnerGeometry()
	}
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
	err := xproto.PutImageChecked(f.client.wm.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(pid), draw,
		uint16(width), uint16(height), 0, int16(yoff), 0, depth, data).Check()
	if err != nil {
		fyne.LogError("Put image error", err)
	}
}

func (f *frame) createPixmaps(depth byte) error {
	iconPix := x11.TitleHeight(x11.XWin(f.client))
	iconAndBorderPix := iconPix + x11.BorderWidth(x11.XWin(f.client))*2
	f.borderTopWidth = f.Width - iconAndBorderPix

	pid, err := xproto.NewPixmapId(f.client.wm.Conn())
	if err != nil {
		return err
	}

	xproto.CreatePixmap(f.client.wm.Conn(), depth, pid,
		xproto.Drawable(f.client.wm.X().Screen().Root), uint16(f.borderTopWidth), uint16(iconPix))
	f.borderTop = pid

	pid, err = xproto.NewPixmapId(f.client.wm.Conn())
	if err != nil {
		return err
	}

	xproto.CreatePixmap(f.client.wm.Conn(), depth, pid,
		xproto.Drawable(f.client.wm.X().Screen().Root), uint16(iconAndBorderPix), uint16(iconPix))
	f.borderTopRight = pid

	return nil
}

func (f *frame) decorate(force bool) {
	depth := f.client.wm.X().Screen().RootDepth
	refresh := force

	if f.borderTop == 0 || refresh {
		err := f.createPixmaps(depth)
		if err != nil {
			fyne.LogError("New Pixmap Error", err)
			return
		}
		refresh = true
	}

	backR, backG, backB, _ := theme.ButtonColor().RGBA()
	if f.client.Focused() {
		backR, backG, backB, _ = theme.BackgroundColor().RGBA()
	}
	bgColor := backR<<16 | backG<<8 | backB

	drawTop, _ := xproto.NewGcontextId(f.client.wm.Conn())
	xproto.CreateGC(f.client.wm.Conn(), drawTop, xproto.Drawable(f.borderTop), xproto.GcForeground, []uint32{bgColor})
	drawTopRight, _ := xproto.NewGcontextId(f.client.wm.Conn())
	xproto.CreateGC(f.client.wm.Conn(), drawTopRight, xproto.Drawable(f.borderTopRight), xproto.GcForeground, []uint32{bgColor})

	if refresh {
		f.drawDecoration(f.borderTop, drawTop, f.borderTopRight, drawTopRight, depth)
	}

	iconSizePix := x11.TitleHeight(x11.XWin(f.client))
	draw, _ := xproto.NewGcontextId(f.client.wm.Conn())
	xproto.CreateGC(f.client.wm.Conn(), draw, xproto.Drawable(f.client.id), xproto.GcForeground, []uint32{bgColor})
	rect := xproto.Rectangle{X: 0, Y: int16(iconSizePix), Width: uint16(f.Width), Height: uint16(f.Height - iconSizePix)}
	xproto.PolyFillRectangleChecked(f.client.wm.Conn(), xproto.Drawable(f.client.id), draw, []xproto.Rectangle{rect})

	iconAndBorderSizePix := iconSizePix + x11.BorderWidth(x11.XWin(f.client))*2
	if f.borderTopWidth+iconAndBorderSizePix < f.Width {
		rect := xproto.Rectangle{X: int16(f.borderTopWidth), Y: 0,
			Width: uint16(f.Width - f.borderTopWidth - iconAndBorderSizePix), Height: uint16(iconSizePix)}
		xproto.PolyFillRectangleChecked(f.client.wm.Conn(), xproto.Drawable(f.client.id), draw, []xproto.Rectangle{rect})
	}

	xproto.CopyArea(f.client.wm.Conn(), xproto.Drawable(f.borderTop), xproto.Drawable(f.client.id), drawTop,
		0, 0, 0, 0, uint16(f.borderTopWidth), uint16(iconSizePix))
	xproto.CopyArea(f.client.wm.Conn(), xproto.Drawable(f.borderTopRight), xproto.Drawable(f.client.id), drawTopRight,
		0, 0, int16(f.Width-iconAndBorderSizePix), 0, uint16(iconAndBorderSizePix), uint16(iconSizePix))
}

func (f *frame) drawDecoration(pidTop xproto.Pixmap, drawTop xproto.Gcontext, pidTopRight xproto.Pixmap, drawTopRight xproto.Gcontext, depth byte) {
	screen := fynedesk.Instance().Screens().ScreenForWindow(f.client)
	scale := screen.CanvasScale()

	canvas := playground.NewSoftwareCanvas()
	canvas.SetScale(scale)
	canvas.SetPadded(false)
	canMaximize := true
	if windowSizeFixed(f.client.wm.X(), f.client.win) ||
		!windowSizeCanMaximize(f.client.wm.X(), f.client) {
		canMaximize = false
	}
	canvas.SetContent(wm.NewBorder(f.client, f.client.Properties().Icon(), canMaximize))

	heightPix := x11.TitleHeight(x11.XWin(f.client))
	iconBorderPixWidth := heightPix + x11.BorderWidth(x11.XWin(f.client))*2
	widthPix := f.borderTopWidth + iconBorderPixWidth
	canvas.Resize(fyne.NewSize(int(float32(widthPix)/scale)+1, int(wmTheme.TitleHeight)))
	img := canvas.Capture()

	// TODO just copy the label minSize - smallest possible but maybe bigger than window width
	// Draw in pixel rows so we don't overflow count usable by PutImageChecked
	for i := uint(0); i < heightPix; i++ {
		f.copyDecorationPixels(uint32(f.borderTopWidth), 1, 0, uint32(i), img, pidTop, drawTop, depth)
	}

	f.copyDecorationPixels(uint32(iconBorderPixWidth), uint32(heightPix), uint32(f.borderTopWidth), 0, img, pidTopRight, drawTopRight, depth)
}

func (f *frame) getInnerWindowCoordinates(w, h uint) fynedesk.Geometry {
	if f.client.Fullscreened() || !f.client.Properties().Decorated() {
		constrainW, constrainH := w, h
		if !f.client.Properties().Decorated() {
			adjustedW, adjustedH := windowSizeWithIncrement(f.client.wm.X(), f.client.win, f.Width, f.Height)
			constrainW, constrainH = windowSizeConstrain(f.client.wm.X(), f.client.win,
				adjustedW, adjustedH)
		}
		f.Width = constrainW
		f.Height = constrainH
		return fynedesk.NewGeometry(0, 0, constrainW, constrainH)
	}

	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))

	extraWidth := 2 * borderWidth
	extraHeight := borderWidth + titleHeight
	adjustedW, adjustedH := windowSizeWithIncrement(f.client.wm.X(), f.client.win, w-extraWidth, h-extraHeight)
	constrainW, constrainH := windowSizeConstrain(f.client.wm.X(), f.client.win,
		adjustedW, adjustedH)
	f.Width = constrainW + extraWidth
	f.Height = constrainH + extraHeight

	return fynedesk.NewGeometry(int(borderWidth), int(titleHeight), constrainW, constrainH)
}

func (f *frame) hide() {
	f.client.RaiseToTop() // Lets ensure this client is on top of the stack so we can walk backwards to find the next window to focus
	stack := f.client.wm.Windows()
	for i := len(stack) - 1; i >= 0; i-- {
		if !stack[i].Iconic() {
			stack[i].RaiseToTop()
			stack[i].Focus()
		}
	}
	xproto.ReparentWindow(f.client.wm.Conn(), f.client.win, f.client.wm.X().RootWin(), int16(f.X), int16(f.Y))
	xproto.UnmapWindow(f.client.wm.Conn(), f.client.win)
}

func (f *frame) maximizeApply() {
	if windowSizeFixed(f.client.wm.X(), f.client.win) ||
		!windowSizeCanMaximize(f.client.wm.X(), f.client) {
		return
	}
	f.client.restoreGeometry = f.Geometry

	head := fynedesk.Instance().Screens().ScreenForWindow(f.client)
	maxWidth, maxHeight := fynedesk.Instance().ContentSizePixels(head)
	if f.client.Fullscreened() {
		maxWidth = head.Width
		maxHeight = head.Height
	}
	f.updateGeometry(fynedesk.NewGeometry(head.X, head.Y, maxWidth, maxHeight), true)
	f.notifyInnerGeometry()
	f.applyTheme(true)
}

func (f *frame) mouseDrag(x, y int) {
	if f.client.Maximized() || f.client.Fullscreened() {
		return
	}
	moveDeltaX := x - f.mouseX
	moveDeltaY := y - f.mouseY
	f.mouseX = x
	f.mouseY = y
	if moveDeltaX == 0 && moveDeltaY == 0 {
		return
	}

	if f.resizeBottom || f.resizeLeft || f.resizeRight && !windowSizeFixed(f.client.wm.X(), f.client.win) {
		deltaX := x - f.resizeStartX
		deltaY := y - f.resizeStartY
		x := f.X
		width := f.resizeStartWidth
		height := f.resizeStartHeight
		if f.resizeBottom {
			height += uint(deltaY)
		}
		if f.resizeLeft {
			x += moveDeltaX
			width -= uint(deltaX)
		} else if f.resizeRight {
			width += uint(deltaX)
		}
		f.updateGeometry(fynedesk.NewGeometry(x, f.Y, width, height), false)
	} else if f.moveOnly {
		f.updateGeometry(f.Geometry.MovedBy(moveDeltaX, moveDeltaY), false)
	}
}

func (f *frame) mouseMotion(x, y int) {
	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	buttonWidth := x11.ButtonWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))

	relX := x - f.X
	relY := y - f.Y
	cursor := x11.DefaultCursor
	if relY <= int(titleHeight) { // title bar
		if relX > int(borderWidth) && relX <= int(borderWidth+buttonWidth) {
			cursor = x11.CloseCursor
		}
		err := xproto.ChangeWindowAttributesChecked(f.client.wm.Conn(), f.client.id, xproto.CwCursor,
			[]uint32{uint32(cursor)}).Check()
		if err != nil {
			fyne.LogError("Set Cursor Error", err)
		}
		return
	}
	if f.client.Maximized() || f.client.Fullscreened() || windowSizeFixed(f.client.wm.X(), f.client.win) {
		return
	}

	if relY >= int(f.Height-buttonWidth) { // bottom
		if relX < int(buttonWidth) {
			cursor = x11.ResizeBottomLeftCursor
		} else if relX >= int(f.Width-buttonWidth) {
			cursor = x11.ResizeBottomRightCursor
		} else {
			cursor = x11.ResizeBottomCursor
		}
	} else { // center (sides)
		if relX < int(f.Width-buttonWidth) {
			cursor = x11.ResizeLeftCursor
		} else if relX >= int(f.Width-buttonWidth) {
			cursor = x11.ResizeRightCursor
		}
	}

	err := xproto.ChangeWindowAttributesChecked(f.client.wm.Conn(), f.client.id, xproto.CwCursor,
		[]uint32{uint32(cursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (f *frame) mousePress(x, y int) {
	if !f.client.Focused() {
		f.client.RaiseToTop()
		f.client.Focus()
		return
	}
	if f.client.Maximized() || f.client.Fullscreened() {
		return
	}

	buttonWidth := x11.ButtonWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))
	f.mouseX = x
	f.mouseY = y
	f.resizeStartX = x
	f.resizeStartY = y

	relX := x - f.X
	relY := y - f.Y
	f.resizeStartWidth = f.Width
	f.resizeStartHeight = f.Height
	f.resizeBottom = false
	f.resizeLeft = false
	f.resizeRight = false
	f.moveOnly = false

	if relY >= int(titleHeight) && !windowSizeFixed(f.client.wm.X(), f.client.win) {
		if relY >= int(f.Height-buttonWidth) {
			f.resizeBottom = true
		}
		if relX < int(buttonWidth) {
			f.resizeLeft = true
		} else if relX >= int(f.Width-buttonWidth) {
			f.resizeRight = true
		}
	} else if relY < int(titleHeight) {
		f.moveOnly = true
	}

	f.client.wm.RaiseToTop(f.client)
}

func (f *frame) mouseRelease(x, y int) {
	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	buttonWidth := x11.ButtonWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))

	relX := x - f.X
	relY := y - f.Y
	barYMax := int(titleHeight)
	if relY > barYMax {
		return
	}
	if relX >= int(borderWidth) && relX < int(borderWidth+buttonWidth) {
		f.client.Close()
	}
	if relX >= int(borderWidth)+int(theme.Padding())+int(buttonWidth) &&
		relX < int(borderWidth)+int(theme.Padding()*2)+int(buttonWidth*2) {
		if f.client.Maximized() {
			f.client.Unmaximize()
		} else {
			f.client.Maximize()
		}
	} else if relX >= int(borderWidth)+int(theme.Padding()*2)+int(buttonWidth*2) &&
		relX < int(borderWidth)+int(theme.Padding()*2)+int(buttonWidth*3) {
		f.client.Iconify()
	}

	f.updateGeometry(f.Geometry, false)
}

// Notify the child window that it's geometry has changed to update menu positions etc.
// This should be used sparingly as it can impact performance on the child window.
func (f *frame) notifyInnerGeometry() {
	geom := f.getInnerWindowCoordinates(f.Width, f.Height)
	ev := xproto.ConfigureNotifyEvent{Event: f.client.win, Window: f.client.win, AboveSibling: 0,
		X: int16(f.X + geom.X), Y: int16(f.Y + geom.Y), Width: uint16(geom.Width), Height: uint16(geom.Height),
		BorderWidth: uint16(x11.BorderWidth(x11.XWin(f.client))), OverrideRedirect: false}
	xproto.SendEvent(f.client.wm.Conn(), false, f.client.win, xproto.EventMaskStructureNotify, string(ev.Bytes()))
}

func (f *frame) removeBorder() {
	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))

	if !f.client.maximized {
		f.X += int(borderWidth)
		f.Y += int(titleHeight)
		f.Width = f.childWidth
		f.Height = f.childHeight
	}
	f.applyTheme(true)

	err := xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight, x11.GeometryToUint32s(f.Geometry)).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	err = xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{0, 0, uint32(f.childWidth), uint32(f.childHeight)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}

	ewmh.FrameExtentsSet(f.client.wm.X(), f.client.win, &ewmh.FrameExtents{Left: 0, Right: 0, Top: 0, Bottom: 0})
	f.notifyInnerGeometry()
}

func (f *frame) show() {
	c := f.client
	xproto.MapWindow(c.wm.Conn(), c.id)

	xproto.ChangeSaveSet(c.wm.Conn(), xproto.SetModeInsert, c.win)
	xproto.MapWindow(c.wm.Conn(), c.win)
	c.wm.BindKeys(c)
	xproto.GrabButton(f.client.wm.Conn(), true, f.client.id,
		xproto.EventMaskButtonPress, xproto.GrabModeSync, xproto.GrabModeSync,
		f.client.wm.X().RootWin(), xproto.CursorNone, xproto.ButtonIndex1, xproto.ModMaskAny)

	c.RaiseToTop()
	c.Focus()
}

func (f *frame) unFrame() {
	c := f.client
	c.wm.RemoveWindow(c)

	if f != nil {
		xproto.ReparentWindow(c.wm.Conn(), c.win, c.wm.X().RootWin(), int16(f.X), int16(f.Y))
	}
	xproto.ChangeSaveSet(c.wm.Conn(), xproto.SetModeDelete, c.win)
	xproto.UnmapWindow(c.wm.Conn(), c.id)
}

func (f *frame) unmaximizeApply() {
	if windowSizeFixed(f.client.wm.X(), f.client.win) ||
		!windowSizeCanMaximize(f.client.wm.X(), f.client) {
		return
	}
	if f.client.restoreGeometry.Width == 0 && f.client.restoreGeometry.Height == 0 {
		screen := fynedesk.Instance().Screens().ScreenForWindow(f.client)
		f.client.restoreGeometry.Width = screen.Width / 2
		f.client.restoreGeometry.Height = screen.Height / 2
	}
	f.updateGeometry(f.client.restoreGeometry, true)
	f.notifyInnerGeometry()
	f.applyTheme(true)
}

func (f *frame) updateGeometry(g fynedesk.Geometry, force bool) {
	var move, resize bool
	if !force {
		resize = g.Width != f.Width || g.Height != f.Height
		move = g.X != f.X || g.Y != f.Y
		if !move && !resize {
			return
		}
	}

	currentScreen := fynedesk.Instance().Screens().ScreenForWindow(f.client)
	f.X = g.X
	f.Y = g.Y

	childGeom := f.getInnerWindowCoordinates(g.Width, g.Height)
	f.childWidth = childGeom.Width
	f.childHeight = childGeom.Height

	err := xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight, x11.GeometryToUint32s(f.Geometry)).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	err = xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(childGeom.X), uint32(childGeom.Y), uint32(f.childWidth), uint32(f.childHeight)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}

	newScreen := fynedesk.Instance().Screens().ScreenForWindow(f.client)
	if newScreen != currentScreen {
		f.updateScale()
		fynedesk.Instance().Screens().SetActive(newScreen)
	}
}

func (f *frame) updateScale() {
	xproto.FreePixmap(f.client.wm.Conn(), f.borderTop)
	f.borderTop = 0
	xproto.FreePixmap(f.client.wm.Conn(), f.borderTopRight)
	f.borderTopRight = 0

	f.updateGeometry(f.Geometry, true)
	f.applyTheme(true)
}
