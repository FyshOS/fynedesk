package wm

import (
	"log"

	"fyne.io/fyne/theme"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type frame struct {
	id, win xproto.Window
	wm      *x11WM
}

func (f *frame) unframe() {
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
	f.unframe()

	err := xproto.DestroyWindowChecked(f.wm.x.Conn(), f.win).Check()
	if err != nil {
		log.Println("Close Err", err)
	}
}

func (f *frame) tapped(x, y int16) {
	if x <= 25 && y <= 18 {
		f.close()
	}
}

func (f *frame) ApplyTheme() {
	r, g, b, _ := theme.BackgroundColor().RGBA()
	r8, g8, b8 := uint8(r), uint8(g), uint8(b)
	values := []uint32{uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)}

	err := xproto.ChangeWindowAttributesChecked(f.wm.x.Conn(), f.id,
		xproto.CwBackPixel, values).Check()
	if err != nil {
		log.Println("ChangeAttribute Err", err)
	}

	err = xproto.ClearAreaChecked(f.wm.x.Conn(), false, f.id, 0, 0, 0, 0).Check()
	if err != nil {
		log.Println("ClearArea Err", err)
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
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease |
		xproto.EventMaskFocusChange}
	err = xproto.CreateWindowChecked(wm.x.Conn(), wm.x.Screen().RootDepth, fr.Id, wm.x.RootWin(),
		attrs.X, attrs.Y, attrs.Width+borderWidth*2, attrs.Height+borderWidth*2+titleHeight, 0, xproto.WindowClassInputOutput,
		wm.x.Screen().RootVisual, xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		log.Println("CreateWindow Err", err)
		return nil
	}

	framed := &frame{fr.Id, win, wm}

	fr.Map()
	xproto.ReparentWindow(wm.x.Conn(), win, fr.Id, borderWidth-1, borderWidth+titleHeight-1)
	xproto.MapWindow(wm.x.Conn(), win)
	err = xproto.ConfigureWindowChecked(wm.x.Conn(), fr.Id, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(wm.rootID), uint32(xproto.StackModeTopIf)}).Check()
	if err != nil {
		log.Println("Restack Err", err)
	}

	xproto.SetInputFocus(wm.x.Conn(), 0, win, 0)
	return framed
}
