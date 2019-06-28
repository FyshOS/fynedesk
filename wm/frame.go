// +build linux,!ci

package wm

import (
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/fyne-io/desktop"
	"github.com/fyne-io/desktop/driver"
	wmTheme "github.com/fyne-io/desktop/theme"
)

type frame struct {
	id, win        xproto.Window
	x, y           int16
	width, height  uint16
	mouseX, mouseY int16
	mouseResize    bool

	framed bool
	wm     *x11WM
	canvas driver.WindowlessCanvas
}

func (s *stack) frameForWin(id xproto.Window) desktop.Window {
	for _, w := range s.frames {
		if w.(*frame).id == id || w.(*frame).win == id {
			return w
		}
	}

	return nil
}

func (f *frame) unFrame() {
	if !f.framed {
		return
	}

	frame := f.wm.frameForWin(f.win)
	f.wm.RemoveWindow(frame)

	if frame == nil {
		return
	}

	xproto.ReparentWindow(f.wm.x.Conn(), f.win, f.wm.x.RootWin(), f.x, f.y)
	xproto.UnmapWindow(f.wm.x.Conn(), f.id)
}

func (f *frame) Decorated() bool {
	return f.framed
}

func (f *frame) Close() {
	winProtos, err := icccm.WmProtocolsGet(f.wm.x, f.win)
	if err != nil {
		log.Println("GetProtocols err", err)
	}

	askNicely := false
	for _, proto := range winProtos {
		if proto == "WM_DELETE_WINDOW" {
			askNicely = true
		}
	}

	if !askNicely {
		err := xproto.DestroyWindowChecked(f.wm.x.Conn(), f.win).Check()
		if err != nil {
			log.Println("Close Err", err)
		}

		return
	}

	protocols, err := xprop.Atm(f.wm.x, "WM_PROTOCOLS")
	if err != nil {
		return
	}

	delWin, err := xprop.Atm(f.wm.x, "WM_DELETE_WINDOW")
	if err != nil {
		return
	}
	cm, err := xevent.NewClientMessage(32, f.win, protocols,
		int(delWin))
	err = xproto.SendEventChecked(f.wm.x.Conn(), false, f.win, 0,
		string(cm.Bytes())).Check()
	if err != nil {
		log.Println("WinDelete Err", err)
	}
}

func (f *frame) Focus() {
	xproto.SetInputFocus(f.wm.x.Conn(), 0, f.win, 0)
}

func (f *frame) press(x, y int16) {
	f.mouseX = x
	f.mouseY = y
	if x >= int16(f.width-f.wm.buttonWidth()) && y >= int16(f.height-f.wm.buttonWidth()) {
		f.mouseResize = true
	} else {
		f.mouseResize = false
	}

	f.wm.RaiseToTop(f)
}

func (f *frame) release(x, y int16) {
	if x > int16(f.wm.borderWidth()) && x < int16(f.wm.borderWidth()+f.wm.buttonWidth()) &&
		y < int16(f.wm.borderWidth()+f.wm.titleHeight()) {
		f.Close()
	}
}

func (f *frame) drag(x, y int16) {
	deltaX := x - f.mouseX
	deltaY := y - f.mouseY

	if f.mouseResize {
		f.width = f.width + uint16(deltaX)
		f.height = f.height + uint16(deltaY)
		f.mouseX = x
		f.mouseY = y

		if f.framed {
			err := xproto.ConfigureWindowChecked(f.wm.x.Conn(), f.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
				[]uint32{uint32(f.wm.borderWidth()), uint32(f.wm.borderWidth() + f.wm.titleHeight()), uint32(f.width - 2*f.wm.borderWidth()),
					uint32(f.height - 2*f.wm.borderWidth() - f.wm.titleHeight())}).Check()
			if err != nil {
				log.Println("ConfigureWindow Err", err)
			}

			f.ApplyTheme()
		}
	} else {
		f.x += deltaX
		f.y += deltaY
	}

	err := xproto.ConfigureWindowChecked(f.wm.x.Conn(), f.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		log.Println("ConfigureWindow Err", err)
	}
}

func (f *frame) motion(x, y int16) {
	cursor := defaultCursor
	if x > int16(f.wm.borderWidth()) && x <= int16(f.wm.borderWidth()+f.wm.buttonWidth()) &&
		y <= int16(f.wm.borderWidth()+f.wm.titleHeight()) {
		cursor = closeCursor
	} else if x >= int16(f.width-f.wm.buttonWidth()) && y >= int16(f.height-f.wm.buttonWidth()) {
		cursor = resizeCursor
	}

	err := xproto.ChangeWindowAttributesChecked(f.wm.x.Conn(), f.id, xproto.CwCursor,
		[]uint32{uint32(cursor)}).Check()
	if err != nil {
		log.Println("SetCursor Err", err)
	}
}

func (f *frame) RaiseAbove(win desktop.Window) {
	topID := f.wm.rootID
	if win != nil {
		topID = win.(*frame).id
	}

	f.Focus()
	if f.id == topID {
		return
	}

	f.wm.raiseWinAboveID(f.id, topID)
}

func (x *x11WM) raiseWinAboveID(win, top xproto.Window) {
	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(top), uint32(xproto.StackModeAbove)}).Check()
	if err != nil {
		log.Println("Restack Err", err)
	}
}

func (f *frame) ApplyTheme() {
	if !f.framed {
		return
	}

	depth := f.wm.x.Screen().RootDepth
	backR, backG, backB, _ := theme.BackgroundColor().RGBA()
	bgColor := backR<<16 | backG<<8 | backB
	prop, _ := xprop.GetProperty(f.wm.x, f.win, "WM_NAME")
	title := ""
	if prop != nil {
		title = string(prop.Value)
	}

	if f.canvas == nil {
		f.canvas = driver.NewSoftwareCanvas(widget.NewLabel(title))
	}
	scale := float32(1) // TODO detect like the gl driver
	f.canvas.SetScale(scale)
	f.canvas.SetContent(newBorder(title))
	f.canvas.ApplyTheme()

	f.canvas.Resize(fyne.NewSize(int(float32(f.width)*scale), int(float32(f.height)*scale)))
	img := f.canvas.Capture()

	pid, err := xproto.NewPixmapId(f.wm.x.Conn())
	if err != nil {
		log.Println("NewPixmap Err", err)
		return
	}
	xproto.CreatePixmap(f.wm.x.Conn(), depth, pid,
		xproto.Drawable(f.wm.x.Screen().Root), f.width, f.height)

	draw, _ := xproto.NewGcontextId(f.wm.x.Conn())
	xproto.CreateGC(f.wm.x.Conn(), draw, xproto.Drawable(pid), xproto.GcForeground, []uint32{bgColor})

	rect := xproto.Rectangle{X: 0, Y: 0, Width: f.width, Height: f.height}
	xproto.PolyFillRectangleChecked(f.wm.x.Conn(), xproto.Drawable(pid), draw, []xproto.Rectangle{rect})

	// DATA is BGRx
	width, height := uint32(f.width), uint32(wmTheme.TitleHeight+wmTheme.BorderWidth)
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

	xproto.PutImageChecked(f.wm.x.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(pid), draw,
		uint16(width), uint16(height), 0, 0, 0, depth, data)

	err = xproto.ChangeWindowAttributesChecked(f.wm.x.Conn(), f.id,
		xproto.CwBackPixmap, []uint32{uint32(pid)}).Check()
	if err != nil {
		log.Println("ChangeAttribute Err", err)
		err = xproto.ChangeWindowAttributesChecked(f.wm.x.Conn(), f.id,
			xproto.CwBackPixmap, []uint32{0}).Check()
		log.Println(err)
	}

	xproto.FreePixmap(f.wm.x.Conn(), pid)
}

func newFrame(win xproto.Window, wm *x11WM) *frame {
	attrs, err := xproto.GetGeometry(wm.x.Conn(), xproto.Drawable(win)).Reply()
	if err != nil {
		log.Println("GetGeometry Err", err)
		return nil
	}

	fr, err := xwindow.Generate(wm.x)
	if err != nil {
		log.Println("GenerateWindow Err", err)
		return nil
	}

	r, g, b, _ := theme.BackgroundColor().RGBA()
	values := []uint32{r<<16 | g<<8 | b, xproto.EventMaskStructureNotify | xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect | xproto.EventMaskExposure |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease | xproto.EventMaskButtonMotion |
		xproto.EventMaskKeyPress | xproto.EventMaskPointerMotion | xproto.EventMaskFocusChange}
	err = xproto.CreateWindowChecked(wm.x.Conn(), wm.x.Screen().RootDepth, fr.Id, wm.x.RootWin(),
		attrs.X, attrs.Y, attrs.Width+wm.borderWidth()*2, attrs.Height+wm.borderWidth()*2+wm.titleHeight(), 0, xproto.WindowClassInputOutput,
		wm.x.Screen().RootVisual, xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		log.Println("CreateWindow Err", err)
		return nil
	}

	framed := &frame{id: fr.Id, win: win, x: attrs.X, y: attrs.Y, width: attrs.Width + wm.borderWidth()*2,
		height: attrs.Height + wm.borderWidth()*2 + wm.titleHeight(), wm: wm, framed: true}
	framed.ApplyTheme()

	fr.Map()
	xproto.ChangeSaveSet(wm.x.Conn(), xproto.SetModeInsert, win)
	xproto.ReparentWindow(wm.x.Conn(), win, fr.Id, int16(wm.borderWidth()-1), int16(wm.borderWidth()+wm.titleHeight()-1))
	xproto.MapWindow(wm.x.Conn(), win)

	return framed
}

func newFrameBorderless(win xproto.Window, wm *x11WM) *frame {
	attrs, err := xproto.GetGeometry(wm.x.Conn(), xproto.Drawable(win)).Reply()
	if err != nil {
		log.Println("GetGeometry Err", err)
		return nil
	}

	return &frame{id: win, win: win, x: attrs.X, y: attrs.Y, wm: wm, framed: false}
}
