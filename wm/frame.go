// +build linux,!ci

package wm

import (
	"log"

	"fyne.io/fyne/theme"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type frame struct {
	id, win        xproto.Window
	x, y           int16
	mouseX, mouseY int16

	wm *x11WM
}

func (f *frame) unFrame() {
	frame := f.wm.frames[f.win]
	delete(f.wm.frames, f.win)

	if frame == nil {
		return
	}
	attrs, err := xproto.GetGeometry(f.wm.x.Conn(), xproto.Drawable(frame.id)).Reply()
	if err != nil {
		log.Println("GetGeometry Err", err)
		return
	}

	xproto.ReparentWindow(f.wm.x.Conn(), f.win, f.wm.x.RootWin(), attrs.X, attrs.Y)

	xproto.UnmapWindow(f.wm.x.Conn(), f.id)
}

func (f *frame) close() {
	err := xproto.DestroyWindowChecked(f.wm.x.Conn(), f.win).Check()
	if err != nil {
		log.Println("Close Err", err)
	}

	// TODO if top pick next down - requires real stack handling
}

func (f *frame) press(x, y int16) {
	f.mouseX = x
	f.mouseY = y

	f.stackTop()
}

func (f *frame) release(x, y int16) {
	if x <= 25 && y <= 19 {
		f.close()
	}
}

func (f *frame) motion(x, y int16) {
	deltaX := x - f.mouseX
	deltaY := y - f.mouseY

	f.x += deltaX
	f.y += deltaY

	err := xproto.ConfigureWindowChecked(f.wm.x.Conn(), f.id, xproto.ConfigWindowX|xproto.ConfigWindowY,
		[]uint32{uint32(f.x), uint32(f.y)}).Check()
	if err != nil {
		log.Println("ConfigureWindow Err", err)
	}
}

func (f *frame) stackTop() {
	err := xproto.ConfigureWindowChecked(f.wm.x.Conn(), f.id, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(f.wm.topID), uint32(xproto.StackModeAbove)}).Check()
	if err != nil {
		log.Println("Restack Err", err)
	}

	f.wm.topID = f.id
	xproto.SetInputFocus(f.wm.x.Conn(), 0, f.win, 0)

	f.ApplyTheme()
}

func (f *frame) ApplyTheme() {
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
		xproto.EventMaskFocusChange}
	err = xproto.CreateWindowChecked(wm.x.Conn(), wm.x.Screen().RootDepth, fr.Id, wm.x.RootWin(),
		attrs.X, attrs.Y, attrs.Width+borderWidth*2, attrs.Height+borderWidth*2+titleHeight, 0, xproto.WindowClassInputOutput,
		wm.x.Screen().RootVisual, xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		log.Println("CreateWindow Err", err)
		return nil
	}

	framed := &frame{id: fr.Id, win: win, x: attrs.X, y: attrs.Y, wm: wm}

	fr.Map()
	xproto.ChangeSaveSet(wm.x.Conn(), xproto.SetModeInsert, win)
	xproto.ReparentWindow(wm.x.Conn(), win, fr.Id, borderWidth-1, borderWidth+titleHeight-1)
	xproto.MapWindow(wm.x.Conn(), win)
	err = xproto.ConfigureWindowChecked(wm.x.Conn(), fr.Id, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(wm.rootID), uint32(xproto.StackModeTopIf)}).Check()
	if err != nil {
		log.Println("Restack Err", err)
	}

	framed.ApplyTheme()
	return framed
}
