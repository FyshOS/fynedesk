// +build linux,!ci

package wm

import (
	"errors"
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/fyne-io/desktop"
)

const (
	borderWidth = 5
	titleHeight = 15
)

type x11WM struct {
	x      *xgbutil.XUtil
	root   fyne.Window
	rootID xproto.Window
	loaded bool

	frames map[xproto.Window]xproto.Window
}

func (x *x11WM) Close() {
	log.Println("Disconnecting from X server")

	for child := range x.frames {
		x.unframe(child)
	}

	x.x.Conn().Close()
}

func (x *x11WM) SetRoot(win fyne.Window) {
	x.root = win
}

// NewX11WindowManager sets up a new X11 Window Manager to control a desktop in X11.
func NewX11WindowManager(a fyne.App) (desktop.WindowManager, error) {
	conn, err := xgbutil.NewConn()
	if err != nil {
		log.Println("Failed to connect to the XServer", err)
		return nil, err
	}

	mgr := &x11WM{x: conn}
	mgr.frames = make(map[xproto.Window]xproto.Window)
	mgr.frameExisting()

	root := conn.RootWin()
	eventMask := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskVisibilityChange |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect
	if err := xproto.ChangeWindowAttributesChecked(conn.Conn(), root, xproto.CwEventMask,
		[]uint32{uint32(eventMask)}).Check(); err != nil {
		conn.Conn().Close()

		return nil, errors.New("Window manager detected, running embedded")
	}

	go mgr.runLoop()

	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	go func() {
		for {
			<-listener
			for _, frame := range mgr.frames {
				mgr.ApplyTheme(frame)
			}
		}
	}()

	return mgr, nil
}

func (x *x11WM) runLoop() {
	conn := x.x.Conn()

	for {
		ev, err := conn.WaitForEvent()
		if err != nil {
			log.Println("X11 Error:", err)
			continue
		}

		if ev != nil {
			switch ev := ev.(type) {
			case xproto.MapRequestEvent:
				x.showWindow(ev.Window)
			case xproto.UnmapNotifyEvent:
				x.hideWindow(ev.Window)
			case xproto.ConfigureRequestEvent:
				x.configureWindow(ev.Window, ev)
			case xproto.PropertyNotifyEvent:
				if ev.Atom == xproto.AtomWmIconName {
					log.Println("Hints", ev.State)
				}
			}
		}
	}
}

func (x *x11WM) configureWindow(win xproto.Window, ev xproto.ConfigureRequestEvent) {
	framed := x.frames[win] != 0
	if framed {
		err := xproto.ConfigureWindowChecked(x.x.Conn(), x.frames[win], xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(ev.X), uint32(ev.Y), uint32(ev.Width + borderWidth*2), uint32(ev.Height + borderWidth*2 + titleHeight)}).Check()

		if err != nil {
			log.Println("ConfigureFrame Err", err)
		}
	}
	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(ev.X), uint32(ev.Y), uint32(ev.Width), uint32(ev.Height)}).Check()
	if err != nil {
		log.Println("ConfigureWindow Err", err)
	}

	prop, _ := xprop.GetProperty(x.x, win, "WM_NAME")
	if !framed && prop != nil && x.root != nil && string(prop.Value) == x.root.Title() {
		if x.loaded {
			return
		}
		x.rootID = win
		x.loaded = true

		for _, frame := range x.frames {
			err = xproto.ConfigureWindowChecked(x.x.Conn(), frame, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
				[]uint32{uint32(x.rootID), uint32(xproto.StackModeAbove)}).Check()
			if err != nil {
				log.Println("Restack Err", err)
			}
		}
	}
}

func (x *x11WM) showWindow(win xproto.Window) {
	framed := x.frames[win] != 0
	prop, err := xprop.GetProperty(x.x, win, "WM_NAME")
	if err != nil {
		log.Println("GetProperty Err", err)
	}
	if framed || string(prop.Value) == x.root.Title() {
		err := xproto.MapWindowChecked(x.x.Conn(), win).Check()
		if err != nil {
			log.Println("ShowWindow Err", err)
		}

		return
	}
	if x.rootID == 0 {
		return
	}

	x.frame(win)
}

func (x *x11WM) ApplyTheme(frame xproto.Window) {
	r, g, b, _ := theme.BackgroundColor().RGBA()
	log.Println("rgb", r, g, b)
	r8, g8, b8 := uint8(r), uint8(g), uint8(b)
	values := []uint32{uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)}
	log.Println("v", values[0])

	err := xproto.ChangeWindowAttributesChecked(x.x.Conn(), frame,
		xproto.CwBackPixel, values).Check()
	if err != nil {
		log.Println("ChangeAttribute Err", err)
	}

	attrs, err := xproto.GetGeometry(x.x.Conn(), xproto.Drawable(frame)).Reply()
	if err != nil {
		log.Println("GetGeometry Err", err)
		return
	}
	err = xproto.ClearAreaChecked(x.x.Conn(), true, frame, 0, 0, attrs.Width, attrs.Height).Check()
	if err != nil {
		log.Println("ClearArea Err", err)
	}
}

func (x *x11WM) frame(win xproto.Window) {
	attrs, err := xproto.GetGeometry(x.x.Conn(), xproto.Drawable(win)).Reply()
	if err != nil {
		log.Println("GetGeometry Err", err)
		return
	}

	frame, err := xwindow.Generate(x.x)
	if err != nil {
		log.Println("GenerateWindow Err", err)
		return
	}

	r, g, b, _ := theme.BackgroundColor().RGBA()
	values := []uint32{r<<16 | g<<8 | b, xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify | xproto.EventMaskSubstructureRedirect |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease |
		xproto.EventMaskFocusChange}
	err = xproto.CreateWindowChecked(x.x.Conn(), x.x.Screen().RootDepth, frame.Id, x.x.RootWin(),
		0, 0, attrs.Width+borderWidth*2, attrs.Height+borderWidth*2+titleHeight, 0, xproto.WindowClassInputOutput,
		x.x.Screen().RootVisual, xproto.CwBackPixel|xproto.CwEventMask, values).Check()
	if err != nil {
		log.Println("CreateWindow Err", err)
		return
	}

	x.frames[win] = frame.Id

	frame.Map()
	xproto.ReparentWindow(x.x.Conn(), win, frame.Id, borderWidth-1, borderWidth+titleHeight-1)
	xproto.MapWindow(x.x.Conn(), win)
	err = xproto.ConfigureWindowChecked(x.x.Conn(), frame.Id, xproto.ConfigWindowSibling|xproto.ConfigWindowStackMode,
		[]uint32{uint32(x.rootID), uint32(xproto.StackModeTopIf)}).Check()
	if err != nil {
		log.Println("Restack Err", err)
	}

	xproto.SetInputFocus(x.x.Conn(), 0, win, 0)
}

func (x *x11WM) unframe(win xproto.Window) {
	frame := x.frames[win]
	x.frames[win] = 0

	if frame == 0 {
		return
	}
	attrs, err := xproto.GetGeometry(x.x.Conn(), xproto.Drawable(frame)).Reply()
	if err != nil {
		log.Println("GetGeometry Err", err)
		return
	}

	xproto.ReparentWindow(x.x.Conn(), win, x.x.RootWin(), attrs.X, attrs.Y)

	xproto.UnmapWindow(x.x.Conn(), frame)
}

func (x *x11WM) frameExisting() {
	tree, err := xproto.QueryTree(x.x.Conn(), x.x.RootWin()).Reply()
	if err != nil {
		log.Println("QueryTree Err", err)
		return
	}

	for _, child := range tree.Children {
		prop, _ := xprop.GetProperty(x.x, child, "WM_NAME")
		if prop == nil || x.root != nil && string(prop.Value) == x.root.Title() {
			continue
		}

		attrs, err := xproto.GetWindowAttributes(x.x.Conn(), child).Reply()
		if err != nil {
			log.Println("GetWindowAttributes Err", err)
			continue
		}
		if attrs.MapState == xproto.MapStateUnmapped {
			continue
		}

		x.frame(child)
	}
}

func (x *x11WM) hideWindow(win xproto.Window) {
	if x.frames[win] != 0 {
		frame := x.frames[win]
		xproto.UnmapWindow(x.x.Conn(), frame)
	}
}
