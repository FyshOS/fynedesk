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
	cli                         *client
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

func (fr *frame) unFrame() {
	if !fr.framed {
		return
	}

	cli := fr.cli.wm.clientForWin(fr.cli.win)
	fr.cli.wm.RemoveWindow(cli)

	if cli == nil {
		return
	}

	if fr != nil {
		xproto.ReparentWindow(fr.cli.wm.x.Conn(), fr.cli.win, fr.cli.wm.x.RootWin(), fr.x, fr.y)
	}
	xproto.UnmapWindow(fr.cli.wm.x.Conn(), fr.cli.id)
}

func (fr *frame) press(x, y int16) {
	fr.mouseX = x
	fr.mouseY = y
	if x >= int16(fr.width-fr.cli.wm.buttonWidth()) && y >= int16(fr.height-fr.cli.wm.buttonWidth()) {
		fr.mouseResize = true
	} else {
		fr.mouseResize = false
	}

	fr.cli.wm.RaiseToTop(fr.cli)
}

func (fr *frame) release(x, y int16) {
	if x > int16(fr.cli.wm.borderWidth()) && x < int16(fr.cli.wm.borderWidth()+fr.cli.wm.buttonWidth()) &&
		y < int16(fr.cli.wm.borderWidth()+fr.cli.wm.titleHeight()) {
		fr.cli.Close()
	} else if x > int16(fr.cli.wm.borderWidth())+int16(theme.Padding())+int16(fr.cli.wm.buttonWidth()) &&
		x < int16(fr.cli.wm.borderWidth())+int16(theme.Padding()*2)+int16(fr.cli.wm.buttonWidth()*2) &&
		y < int16(fr.cli.wm.borderWidth())+int16(fr.cli.wm.titleHeight()) {
		if fr.cli.Maximized() {
			fr.cli.Unmaximize()
		} else {
			fr.cli.Maximize()
		}
	} else if x > int16(fr.cli.wm.borderWidth())+int16(theme.Padding()*2)+int16(fr.cli.wm.buttonWidth()*2) &&
		x < int16(fr.cli.wm.borderWidth())+int16(theme.Padding()*2)+int16(fr.cli.wm.buttonWidth()*3) &&
		y < int16(fr.cli.wm.borderWidth())+int16(fr.cli.wm.titleHeight()) {
		fr.cli.Iconify()
	}
}

func (fr *frame) drag(x, y int16) {
	deltaX := x - fr.mouseX
	deltaY := y - fr.mouseY
	var moveOnly = true

	if fr.mouseResize {
		fr.width = fr.width + uint16(deltaX)
		fr.height = fr.height + uint16(deltaY)
		fr.mouseX = x
		fr.mouseY = y

		borderWidth := 2 * fr.cli.wm.borderWidth()
		if fr.width < uint16(fr.minWidth)+borderWidth {
			fr.width = uint16(fr.minWidth) + borderWidth
		}
		borderHeight := 2*fr.cli.wm.borderWidth() + fr.cli.wm.titleHeight()
		if fr.height < uint16(fr.minHeight)+borderHeight {
			fr.height = uint16(fr.minHeight) + borderHeight
		}
		moveOnly = false
	} else {
		fr.x += deltaX
		fr.y += deltaY
	}
	fr.updateGeometry(uint16(fr.x), uint16(fr.y), fr.width, fr.height, moveOnly)
}

func (fr *frame) motion(x, y int16) {
	cursor := defaultCursor
	if x > int16(fr.cli.wm.borderWidth()) && x <= int16(fr.cli.wm.borderWidth()+fr.cli.wm.buttonWidth()) &&
		y <= int16(fr.cli.wm.borderWidth()+fr.cli.wm.titleHeight()) {
		cursor = closeCursor
	} else if x >= int16(fr.width-fr.cli.wm.buttonWidth()) && y >= int16(fr.height-fr.cli.wm.buttonWidth()) {
		cursor = resizeCursor
	}

	err := xproto.ChangeWindowAttributesChecked(fr.cli.wm.x.Conn(), fr.cli.id, xproto.CwCursor,
		[]uint32{uint32(cursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (fr *frame) updateGeometry(x uint16, y uint16, w uint16, h uint16, moveOnly bool) {
	if fr.framed && !moveOnly {
		borderWidth := 2 * fr.cli.wm.borderWidth()
		if w < uint16(fr.minWidth)+borderWidth {
			w = uint16(fr.minWidth) + borderWidth
		}
		borderHeight := 2*fr.cli.wm.borderWidth() + fr.cli.wm.titleHeight()
		if h < uint16(fr.minHeight)+borderHeight {
			h = uint16(fr.minHeight) + borderHeight
		}
		err := xproto.ConfigureWindowChecked(fr.cli.wm.x.Conn(), fr.cli.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(fr.cli.wm.borderWidth()), uint32(fr.cli.wm.borderWidth() + fr.cli.wm.titleHeight()),
				uint32(w - borderWidth), uint32(h - borderHeight)}).Check()
		if err != nil {
			fyne.LogError("Configure Window Error", err)
		}
		fr.applyTheme()
	}
	err := xproto.ConfigureWindowChecked(fr.cli.wm.x.Conn(), fr.cli.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(x), uint32(y), uint32(w), uint32(h)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (fr *frame) maximizeFrame() {
	var w = fr.cli.wm.x.Screen().WidthInPixels
	var h = fr.cli.wm.x.Screen().HeightInPixels
	var x, y = 0, 0
	fr.restoreWidth = fr.width
	fr.restoreHeight = fr.height
	fr.restoreX = fr.x
	fr.restoreY = fr.y
	fr.width = w
	fr.height = h
	fr.x = int16(x)
	fr.y = int16(y)
	fr.updateGeometry(uint16(x), uint16(y), w, h, false)
}

func (fr *frame) unmaximizeFrame() {
	fr.width = fr.restoreWidth
	fr.height = fr.restoreHeight
	fr.x = fr.restoreX
	fr.y = fr.restoreY
	fr.updateGeometry(uint16(fr.x), uint16(fr.y), fr.width, fr.height, false)
}

func (fr *frame) applyTheme() {
	if !fr.framed {
		return
	}

	depth := fr.cli.wm.x.Screen().RootDepth
	backR, backG, backB, _ := theme.BackgroundColor().RGBA()
	bgColor := backR<<16 | backG<<8 | backB

	scale := float32(2)
	if fr.canvas == nil {
		fr.canvas = util.NewSoftwareCanvas()
		fr.canvas.SetPadded(false)
	}
	scale = desktop.Instance().Root().Canvas().Scale()
	fr.canvas.SetScale(scale)
	border := newBorder(fr.cli)
	border.Resize(fr.canvas.Size())
	fr.canvas.SetContent(border)

	fr.canvas.Resize(fyne.NewSize(int(float32(fr.width)/scale), int(float32(fr.height)/scale)))
	img := fr.canvas.Capture()

	pid, err := xproto.NewPixmapId(fr.cli.wm.x.Conn())
	if err != nil {
		fyne.LogError("New Pixmap Error", err)
		return
	}
	xproto.CreatePixmap(fr.cli.wm.x.Conn(), depth, pid,
		xproto.Drawable(fr.cli.wm.x.Screen().Root), fr.width, fr.height)

	draw, _ := xproto.NewGcontextId(fr.cli.wm.x.Conn())
	xproto.CreateGC(fr.cli.wm.x.Conn(), draw, xproto.Drawable(pid), xproto.GcForeground, []uint32{bgColor})

	rect := xproto.Rectangle{X: 0, Y: 0, Width: fr.width, Height: fr.height}
	xproto.PolyFillRectangleChecked(fr.cli.wm.x.Conn(), xproto.Drawable(pid), draw, []xproto.Rectangle{rect})

	// DATA is BGRx
	width, height := uint32(fr.width), uint32(float32(wmTheme.TitleHeight)*scale+float32(wmTheme.BorderWidth)*scale)
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

	xproto.PutImageChecked(fr.cli.wm.x.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(pid), draw,
		uint16(width), uint16(height), 0, 0, 0, depth, data)

	err = xproto.ChangeWindowAttributesChecked(fr.cli.wm.x.Conn(), fr.cli.id,
		xproto.CwBackPixmap, []uint32{uint32(pid)}).Check()
	if err != nil {
		fyne.LogError("Change Attribute Error", err)
		err = xproto.ChangeWindowAttributesChecked(fr.cli.wm.x.Conn(), fr.cli.id,
			xproto.CwBackPixmap, []uint32{0}).Check()
		if err != nil {
			fyne.LogError("Change Attribute Error", err)
		}
	}

	xproto.FreePixmap(fr.cli.wm.x.Conn(), pid)
}

func newFrame(cli *client) *frame {
	attrs, err := xproto.GetGeometry(cli.wm.x.Conn(), xproto.Drawable(cli.win)).Reply()
	if err != nil {
		fyne.LogError("Get Geometry Error", err)
		return nil
	}

	fr, err := xwindow.Generate(cli.wm.x)
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
	err = xproto.CreateWindowChecked(cli.wm.x.Conn(), cli.wm.x.Screen().RootDepth, fr.Id, cli.wm.x.RootWin(),
		attrs.X, attrs.Y, attrs.Width+uint16(float32(cli.wm.borderWidth())/scale)*2, attrs.Height+uint16(float32(cli.wm.borderWidth())/scale)*2+cli.wm.titleHeight(), 0, xproto.WindowClassInputOutput,
		cli.wm.x.Screen().RootVisual, xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		fyne.LogError("Create Window Error", err)
		return nil
	}
	cli.id = fr.Id
	framed := &frame{cli: cli, x: attrs.X, y: attrs.Y, width: attrs.Width + uint16(float32(cli.wm.borderWidth())*2/scale),
		height: attrs.Height + uint16(float32(cli.wm.borderWidth())/scale*2) + uint16(float32(cli.wm.titleHeight())/scale), framed: true}
	framed.applyTheme()

	fr.Map()
	xproto.ChangeSaveSet(cli.wm.x.Conn(), xproto.SetModeInsert, cli.win)
	xproto.ReparentWindow(cli.wm.x.Conn(), cli.win, cli.id, int16(cli.wm.borderWidth()-1), int16(cli.wm.borderWidth()+cli.wm.titleHeight()-1))
	xproto.MapWindow(cli.wm.x.Conn(), cli.win)

	return framed
}

func newFrameBorderless(cli *client) *frame {
	attrs, err := xproto.GetGeometry(cli.wm.x.Conn(), xproto.Drawable(cli.win)).Reply()
	if err != nil {
		fyne.LogError("Get Geometry Error", err)
		return nil
	}

	cli.id = cli.win
	return &frame{cli: cli, x: attrs.X, y: attrs.Y, framed: false}
}
