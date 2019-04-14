// +build linux,!ci

package wm

import (
	"log"

	"fyne.io/fyne/theme"
	"github.com/fyne-io/desktop"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type frame struct {
	id, win        xproto.Window
	x, y           int16
	width, height  uint16
	mouseX, mouseY int16

	framed bool
	wm     *x11WM
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

	f.wm.RaiseToTop(f)
}

func (f *frame) release(x, y int16) {
	if x <= 25 && y <= 19 {
		f.Close()
	}
}

func (f *frame) drag(x, y int16) {
	deltaX := x - f.mouseX
	deltaY := y - f.mouseY

	if x >= int16(f.width)-25 && y >= int16(f.height)-25 {
		f.width = f.width + uint16(deltaX)
		f.height = f.height + uint16(deltaY)
		f.mouseX = x
		f.mouseY = y

		if f.framed {
			err := xproto.ConfigureWindowChecked(f.wm.x.Conn(), f.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
				[]uint32{uint32(borderWidth), uint32(borderWidth + titleHeight), uint32(f.width - 2*borderWidth), uint32(f.height - 2*borderWidth - titleHeight)}).Check()
			if err != nil {
				log.Println("ConfigureWindow Err", err)
			}
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

	f.ApplyTheme()
}

func (f *frame) motion(x, y int16) {
	cursor := defaultCursor
	if x <= 25 && y <= 19 {
		cursor = closeCursor
	} else if x >= int16(f.width)-25 && y >= int16(f.height)-25 {
		cursor = resizeCursor
	}

	err := xproto.ChangeWindowAttributesChecked(f.wm.x.Conn(), f.id, xproto.CwCursor,
		[]uint32{uint32(cursor)}).Check()
	if err != nil {
		log.Println("SetCursor Err", err)
	}
}

func (f *frame) raiseToTop() {
	topID := f.wm.rootID
	if len(f.wm.frames) >= 1 {
		topID = f.wm.frames[0].(*frame).id
	}

	f.Focus()
	if f.id == topID {
		return
	}

	err := xproto.ConfigureWindowChecked(f.wm.x.Conn(), f.id, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(topID), uint32(xproto.StackModeAbove)}).Check()
	if err != nil {
		log.Println("Restack Err", err)
	}

	f.ApplyTheme()
}

func (f *frame) ApplyTheme() {
	if !f.framed {
		return
	}

	r, g, b, _ := theme.BackgroundColor().RGBA()
	values := []uint32{r<<16 | g<<8 | b}

	err := xproto.ChangeWindowAttributesChecked(f.wm.x.Conn(), f.id,
		xproto.CwBackPixel, values).Check()
	if err != nil {
		log.Println("ChangeAttribute Err", err)
	}

	err = xproto.ClearAreaChecked(f.wm.x.Conn(), false, f.id, 0, 0, 0, 0).Check()
	if err != nil {
		log.Println("ClearArea Err", err)
	}

	rf, gf, bf, _ := theme.ButtonColor().RGBA()
	rect := xproto.Rectangle{X: 0, Y: 0, Width: 25, Height: 19}
	values = []uint32{rf<<16 | gf<<8 | bf}
	gc, err := xproto.NewGcontextId(f.wm.x.Conn())
	err = xproto.CreateGCChecked(f.wm.x.Conn(), gc, xproto.Drawable(f.id), xproto.GcForeground, values).Check()
	if err != nil {
		log.Println("CreateGraphics Err", err)
	}
	err = xproto.PolyFillRectangleChecked(f.wm.x.Conn(), xproto.Drawable(f.id), gc, []xproto.Rectangle{rect}).Check()
	if err != nil {
		log.Println("PolyRectangle Err", err)
	}
	xproto.FreeGC(f.wm.x.Conn(), gc)

	fid, err := xproto.NewFontId(f.wm.x.Conn())
	err = xproto.OpenFontChecked(f.wm.x.Conn(), fid, 5, "fixed").Check()
	if err != nil {
		log.Println("OpenFont Err", err)
	}

	prop, _ := xprop.GetProperty(f.wm.x, f.win, "WM_NAME")
	title := ""
	if prop != nil {
		title = string(prop.Value)
	}

	rf, gf, bf, _ = theme.TextColor().RGBA()
	values = []uint32{rf<<16 | gf<<8 | bf, r<<16 | g<<8 | b, uint32(fid)}
	gc, err = xproto.NewGcontextId(f.wm.x.Conn())
	err = xproto.CreateGCChecked(f.wm.x.Conn(), gc, xproto.Drawable(f.id), xproto.GcForeground|xproto.GcBackground|xproto.GcFont, values).Check()
	if err != nil {
		log.Println("CreateGraphics Err", err)
	}

	xproto.CloseFont(f.wm.x.Conn(), fid)

	err = xproto.ImageText8Checked(f.wm.x.Conn(), byte(len(title)), xproto.Drawable(f.id), gc, 29, 14, title).Check()
	if err != nil {
		log.Println("PolyText8 Err", err)
	}

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
	values := []uint32{r<<16 | g<<8 | b, xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify | xproto.EventMaskSubstructureRedirect |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease | xproto.EventMaskButtonMotion |
		xproto.EventMaskPointerMotion | xproto.EventMaskFocusChange}
	err = xproto.CreateWindowChecked(wm.x.Conn(), wm.x.Screen().RootDepth, fr.Id, wm.x.RootWin(),
		attrs.X, attrs.Y, attrs.Width+borderWidth*2, attrs.Height+borderWidth*2+titleHeight, 0, xproto.WindowClassInputOutput,
		wm.x.Screen().RootVisual, xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		log.Println("CreateWindow Err", err)
		return nil
	}

	framed := &frame{id: fr.Id, win: win, x: attrs.X, y: attrs.Y, width: attrs.Width + borderWidth*2, height: attrs.Height + borderWidth*2 + titleHeight,
		wm: wm, framed: true}

	fr.Map()
	xproto.ChangeSaveSet(wm.x.Conn(), xproto.SetModeInsert, win)
	xproto.ReparentWindow(wm.x.Conn(), win, fr.Id, borderWidth-1, borderWidth+titleHeight-1)
	xproto.MapWindow(wm.x.Conn(), win)

	framed.ApplyTheme()
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
