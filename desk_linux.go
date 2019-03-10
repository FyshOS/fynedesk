// +build !ci

package desktop

import (
	"log"
	"os"

	"fyne.io/fyne"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xwindow"
)

var embed bool

func isEmbedded() bool {
	env := os.Getenv("WAYLAND_DISPLAY")
	if env != "" {
		embed = true
	}

	return embed
}

// newDesktopWindow will return a new window based the current environment.
// When running in an existing desktop then load a window.
// Otherwise let's return a desktop root!
func newDesktopWindow(a fyne.App) fyne.Window {
	if isEmbedded() {
		return createWindow(a)
	}

	conn, err := xgbutil.NewConn()
	if err != nil {
		log.Println("Failed to connect to the XServer", err)
		return nil
	}

	root := conn.RootWin()
	disp := xproto.Setup(conn.Conn()).Roots[0]
	log.Println("Default Root", disp.WidthInPixels, "x", disp.HeightInPixels)

	eventMask := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskVisibilityChange |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect

	win := xwindow.New(conn, root)
	if err := win.Listen(eventMask); err != nil {
		conn.Conn().Close()
		log.Println("Window manager detected, running in Embedded mode")

		embed = true
		return createWindow(a)
	}
	win.Destroy()

	go func() {
		for {
			ev, err := conn.Conn().WaitForEvent()
			if err != nil {
				log.Println("X11 Error:", err)
				continue
			}

			if ev != nil {
				switch ev := ev.(type) {
				case xproto.MapRequestEvent:
					xproto.MapWindow(conn.Conn(), ev.Window)
					xproto.SetInputFocus(conn.Conn(), 0, ev.Window, 0)
				case xproto.ConfigureRequestEvent:
					xproto.ConfigureWindow(conn.Conn(), ev.Window, xproto.ConfigWindowX|xproto.ConfigWindowY|
						xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
						[]uint32{uint32(ev.X), uint32(ev.Y), uint32(ev.Width), uint32(ev.Height)})
				}
			}
		}
	}()

	desk = createWindow(a)
	desk.SetFullScreen(true)

	return desk
}
