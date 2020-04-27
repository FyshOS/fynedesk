// +build linux

package wm // import "fyne.io/fynedesk/internal/x11/wm"

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"fyne.io/fyne"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/ui"
	"fyne.io/fynedesk/internal/x11"
	xwin "fyne.io/fynedesk/internal/x11/win"
)

type x11WM struct {
	stack
	x                       *xgbutil.XUtil
	framedExisting          bool
	moveResizing            bool
	moveResizingStartX      int16
	moveResizingStartY      int16
	moveResizingLastX       int16
	moveResizingLastY       int16
	moveResizingStartWidth  uint
	moveResizingStartHeight uint
	moveResizingType        moveResizeType
	screenChangeTimestamp   xproto.Timestamp

	allowedActions []string
	supportedHints []string

	rootIDs      []xproto.Window
	transientMap map[xproto.Window][]xproto.Window
}

type moveResizeType uint32

const (
	moveResizeTopLeft      moveResizeType = 0
	moveResizeTop          moveResizeType = 1
	moveResizeTopRight     moveResizeType = 2
	moveResizeRight        moveResizeType = 3
	moveResizeBottomRight  moveResizeType = 4
	moveResizeBottom       moveResizeType = 5
	moveResizeBottomLeft   moveResizeType = 6
	moveResizeLeft         moveResizeType = 7
	moveResizeMove         moveResizeType = 8
	moveResizeKeyboard     moveResizeType = 9
	moveResizeMoveKeyboard moveResizeType = 10
	moveResizeCancel       moveResizeType = 11

	keyCodeEscape = 9
	keyCodeTab    = 23
	keyCodeCtrl   = 33
	keyCodeReturn = 36
	keyCodeAlt    = 64
	keyCodeSpace  = 65

	keyCodeEnter = 108
	keyCodeLeft  = 113
	keyCodeRight = 114

	keyCodeMute          = 121
	keyCodeSoundDecrease = 122
	keyCodeSoundIncrease = 123

	keyCodeSuper = 133

	keyCodeBrightLess = 232
	keyCodeBrightMore = 233
)

// NewX11WindowManager sets up a new X11 Window Manager to control a desktop in X11.
func NewX11WindowManager(a fyne.App) (fynedesk.WindowManager, error) {
	conn, err := xgbutil.NewConn()
	if err != nil {
		fyne.LogError("Failed to connect to the XServer", err)
		return nil, err
	}

	mgr := &x11WM{x: conn}
	root := conn.RootWin()
	mgr.takeSelectionOwnership()
	mgr.transientMap = make(map[xproto.Window][]xproto.Window)

	eventMask := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskKeyPress |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureRedirect
	if err := xproto.ChangeWindowAttributesChecked(conn.Conn(), root, xproto.CwEventMask,
		[]uint32{uint32(eventMask)}).Check(); err != nil {
		conn.Conn().Close()

		return nil, errors.New("window manager detected, running embedded")
	}

	ewmh.SupportedSet(mgr.x, x11.SupportedHints)
	ewmh.SupportingWmCheckSet(mgr.x, mgr.x.RootWin(), mgr.x.Dummy())
	ewmh.SupportingWmCheckSet(mgr.x, mgr.x.Dummy(), mgr.x.Dummy())
	ewmh.WmNameSet(mgr.x, mgr.x.Dummy(), ui.RootWindowName)
	ewmh.DesktopViewportSet(mgr.x, []ewmh.DesktopViewport{{X: 0, Y: 0}}) // Will always be 0, 0 until virtual desktops are supported
	ewmh.NumberOfDesktopsSet(mgr.x, 1)                                   // Will always be 1 until virtual desktops are supported
	ewmh.CurrentDesktopSet(mgr.x, 0)                                     // Will always be 0 until virtual desktops are supported

	x11.LoadCursors(conn)
	go mgr.runLoop()

	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	go func() {
		for {
			<-listener
			for _, c := range mgr.clients {
				c.(x11.XWin).SettingsChanged()
			}
			mgr.configureRoots(mgr.x.RootWin())
		}
	}()

	return mgr, nil
}

func (x *x11WM) AddStackListener(l fynedesk.StackListener) {
	x.stack.listeners = append(x.stack.listeners, l)
}

func (x *x11WM) BindKeys(win x11.XWin) {
	x.bindKeys(win.ChildID())
}

func (x *x11WM) Blank() {
	go func() {
		time.Sleep(time.Second / 3)
		exec.Command("xset", "-display", os.Getenv("DISPLAY"), "dpms", "force", "off").Start()
	}()
}

func (x *x11WM) Close() {
	log.Println("Disconnecting from X server")

	for _, child := range x.clients {
		child.Close()
	}

	x.x.Conn().Close()
}

func (x *x11WM) X() *xgbutil.XUtil {
	return x.x
}

func (x *x11WM) Conn() *xgb.Conn {
	return x.x.Conn()
}

func (x *x11WM) runLoop() {
	conn := x.x.Conn()

	for {
		ev, err := conn.WaitForEvent()
		if err != nil {
			fyne.LogError("X11 Error:", err)
			continue
		}
		if ev == nil { // disconnected if both are nil
			break
		}
		switch ev := ev.(type) {
		case xproto.ButtonPressEvent:
			x.handleButtonPress(ev)
		case xproto.ButtonReleaseEvent:
			x.handleButtonRelease(ev)
		case xproto.ClientMessageEvent:
			x.handleClientMessage(ev)
		case xproto.ConfigureNotifyEvent:
			x.configureRoots(ev.Window)
		case xproto.ConfigureRequestEvent:
			x.configureWindow(ev.Window, ev)
		case xproto.CreateNotifyEvent:
			x.setInitialWindowAttributes(ev.Window)
		case xproto.DestroyNotifyEvent:
			x.destroyWindow(ev.Window)
		case xproto.EnterNotifyEvent:
			x.handleMouseEnter(ev)
		case xproto.ExposeEvent:
			x.exposeWindow(ev.Window)
		case xproto.FocusInEvent:
			x.handleFocus(ev.Event)
		case xproto.FocusOutEvent:
			x.handleFocus(ev.Event)
		case xproto.KeyPressEvent:
			x.handleKeyPress(ev)
		case xproto.KeyReleaseEvent:
			x.handleKeyRelease(ev)
		case xproto.LeaveNotifyEvent:
			x.handleMouseLeave(ev)
		case xproto.MapRequestEvent:
			x.showWindow(ev.Window, ev.Parent)
		case xproto.MotionNotifyEvent:
			x.handleMouseMotion(ev)
		case xproto.PropertyNotifyEvent:
			x.handlePropertyChange(ev)
		case randr.ScreenChangeNotifyEvent:
			x.handleScreenChange(ev.Timestamp)
		case xproto.UnmapNotifyEvent:
			x.hideWindow(ev.Window)
		case xproto.VisibilityNotifyEvent:
			x.handleVisibilityChange(ev)
		}
	}

	fyne.LogError("X11 connection terminated!", nil)
}

func (x *x11WM) bindKeys(win xproto.Window) {
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMask1, keyCodeSpace, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMask1, keyCodeTab, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMaskShift|xproto.ModMask1, keyCodeTab, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMaskAny, keyCodeBrightLess, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMaskAny, keyCodeBrightMore, xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (x *x11WM) configureRoots(win xproto.Window) {
	if win != x.x.RootWin() || fynedesk.Instance() == nil {
		return
	}
	width, height := 0, 0
	for _, screen := range fynedesk.Instance().Screens().Screens() {
		win := x.WinIDForScreen(screen)
		if win == 0 {
			continue
		}
		if screen.X+screen.Width > width {
			width = screen.X + screen.Width
		}
		if screen.Y+screen.Height > height {
			height = screen.Y + screen.Height
		}
		xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(screen.X), uint32(screen.Y), uint32(screen.Width), uint32(screen.Height)}).Check()
		notifyEv := xproto.ConfigureNotifyEvent{Event: win, Window: win, AboveSibling: 0,
			X: int16(screen.X), Y: int16(screen.Y), Width: uint16(screen.Width), Height: uint16(screen.Height),
			BorderWidth: 0, OverrideRedirect: false}
		xproto.SendEvent(x.x.Conn(), false, win, xproto.EventMaskStructureNotify, string(notifyEv.Bytes()))
	}
	ewmh.DesktopGeometrySet(x.x, &ewmh.DesktopGeometry{Width: width, Height: height})              // The array will grow when virtual desktops are supported
	ewmh.WorkareaSet(x.x, []ewmh.Workarea{{X: 0, Y: 0, Width: uint(width), Height: uint(height)}}) // The array will grow when virtual desktops are supported
}

func (x *x11WM) configureWindow(win xproto.Window, ev xproto.ConfigureRequestEvent) {
	c := x.clientForWin(win)
	xcoord := ev.X
	ycoord := ev.Y
	width := ev.Width
	height := ev.Height

	if c != nil {
		if c.ChildID() == win { // ignore requests from our frame as we must have caused it
			x, y, _, _ := c.Geometry()
			borderWidth := x11.BorderWidth(c)
			titleHeight := x11.TitleHeight(c)

			if c.Properties().Decorated() {
				if !c.Fullscreened() {
					c.NotifyGeometry(x, y, uint(ev.Width+(borderWidth*2)), uint(ev.Height+borderWidth+titleHeight))
				} else {
					c.NotifyGeometry(x, y, uint(ev.Width), uint(ev.Height))
				}
			} else {
				if ev.X == 0 && ev.Y == 0 {
					ev.X = int16(x)
					ev.Y = int16(y)
				}
				c.NotifyGeometry(int(ev.X), int(ev.Y), uint(ev.Width), uint(ev.Height))
			}
		}
		return
	}

	name := x11.WindowName(x.x, win)
	for _, screen := range fynedesk.Instance().Screens().Screens() {
		if !x.isRootTitle(name) || screenNameFromRootTitle(name) != screen.Name {
			continue
		}
		found := false
		for _, id := range x.rootIDs {
			if id == win {
				found = true
			}
		}
		if !found {
			x.rootIDs = append(x.rootIDs, win)
		}
		xcoord = int16(screen.X)
		ycoord = int16(screen.Y)
		width = uint16(screen.Width)
		height = uint16(screen.Height)
		notifyEv := xproto.ConfigureNotifyEvent{Event: win, Window: win, AboveSibling: 0,
			X: int16(screen.X), Y: int16(screen.Y), Width: uint16(screen.Width), Height: uint16(screen.Height),
			BorderWidth: 0, OverrideRedirect: false}
		xproto.SendEvent(x.x.Conn(), false, win, xproto.EventMaskStructureNotify, string(notifyEv.Bytes()))
		break
	}
	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(xcoord), uint32(ycoord), uint32(width), uint32(height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (x *x11WM) destroyWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		for i, id := range x.rootIDs {
			if id == win {
				x.rootIDs = append(x.rootIDs[:i], x.rootIDs[i+1:]...)
			}
		}
		return
	}
	transient := x11.WindowTransientForGet(x.x, win)
	if transient > 0 && transient != win {
		x.transientChildRemove(transient, win)
	} else if transient > 0 && transient == win {
		x.transientLeaderRemove(transient)
	}
	x.RemoveWindow(c)
	windowClientListUpdate(x)
	windowClientListStackingUpdate(x)
}

func (x *x11WM) exposeWindow(win xproto.Window) {
	border := x.clientForWin(win)
	if border != nil {
		border.Expose()
	}
}

func (x *x11WM) frameExisting() {
	tree, err := xproto.QueryTree(x.x.Conn(), x.x.RootWin()).Reply()
	if err != nil {
		fyne.LogError("Query Tree Error", err)
		return
	}

	for _, child := range tree.Children {
		name := x11.WindowName(x.x, child)
		if x.isRootTitle(name) {
			continue
		}
		attrs, err := xproto.GetWindowAttributes(x.x.Conn(), child).Reply()
		if err != nil {
			fyne.LogError("Get Window Attributes Error", err)
			continue
		}
		if attrs.MapState == xproto.MapStateUnmapped {
			continue
		}
		x.setupWindow(child)
	}
}

func (x *x11WM) WinIDForScreen(screen *fynedesk.Screen) xproto.Window {
	screenName := screen.Name
	for _, id := range x.rootIDs {
		name := x11.WindowName(x.x, id)
		if !x.isRootTitle(name) {
			continue
		}
		if screenNameFromRootTitle(name) == screenName {
			return id
		}
	}
	return 0
}

func (x *x11WM) hideWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		return
	}
	xproto.UnmapWindow(x.x.Conn(), c.FrameID())
}

func (x *x11WM) isRootTitle(title string) bool {
	return strings.Index(title, ui.RootWindowName) == 0
}

func screenNameFromRootTitle(title string) string {
	if len(title) <= len(ui.RootWindowName) {
		return ""
	}
	return title[len(ui.RootWindowName):]
}

func (x *x11WM) setInitialWindowAttributes(win xproto.Window) {
	err := xproto.ChangeWindowAttributesChecked(x.x.Conn(), win, xproto.CwCursor,
		[]uint32{uint32(x11.DefaultCursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (x *x11WM) setupWindow(win xproto.Window) {
	if x11.WindowName(x.x, win) == "" {
		x11.WindowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_TASKBAR")
		x11.WindowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_PAGER")
	}
	c := x.clientForWin(win)
	if c != nil {
		return
	}
	c = xwin.NewClient(win, x)
	x.AddWindow(c)
	c.RaiseToTop()
	c.Focus()
	windowClientListUpdate(x)
	windowClientListStackingUpdate(x)
}

func (x *x11WM) showWindow(win xproto.Window, parent xproto.Window) {
	name := x11.WindowName(x.x, win)
	if x.isRootTitle(name) {
		err := xproto.MapWindowChecked(x.x.Conn(), win).Check()
		if err != nil {
			fyne.LogError("Show Window Error", err)
		}
		xproto.ConfigureWindow(x.x.Conn(), win, xproto.ConfigWindowStackMode, []uint32{xproto.StackModeBelow})
		x.bindKeys(win)
		if !x.framedExisting {
			x.framedExisting = true
			go x.frameExisting()
		}
		return
	}

	hints, err := icccm.WmHintsGet(x.x, win)
	if err == nil {
		if hints.Flags&icccm.HintState > 0 && hints.InitialState == icccm.StateWithdrawn { // We don't want to manage windows that are not mapped
			return
		}
	}

	override := windowOverrideGet(x.x, win) // We don't want to manage windows that have an override on the WM
	if override {
		return
	}

	winType := windowTypeGet(x.x, win)
	switch winType[0] {
	case windowTypeUtility, windowTypeDialog, windowTypeNormal:
		break
	default:
		return
	}

	transient := x11.WindowTransientForGet(x.x, win)
	if transient > 0 && transient != win {
		x.transientChildAdd(transient, win)
	}

	x.setupWindow(win)
}

func (x *x11WM) takeSelectionOwnership() {
	name := fmt.Sprintf("WM_S%d", x.x.Conn().DefaultScreen)
	selAtom, err := xprop.Atm(x.x, name)
	if err != nil {
		fyne.LogError("Error getting selection atom", err)
		return
	}
	err = xproto.SetSelectionOwnerChecked(x.x.Conn(), x.x.Dummy(), selAtom, xproto.TimeCurrentTime).Check()
	if err != nil {
		fyne.LogError("Error setting selection owner", err)
		return
	}
	reply, err := xproto.GetSelectionOwner(x.x.Conn(), selAtom).Reply()
	if err != nil {
		fyne.LogError("Error getting selection owner", err)
		return
	}
	if reply.Owner != x.x.Dummy() {
		fyne.LogError("Could not obtain ownership - Another WM is likely running", err)
	}
	manAtom, err := xprop.Atm(x.x, "MANAGER")
	if err != nil {
		fyne.LogError("Error getting manager atom", err)
		return
	}
	cm, err := xevent.NewClientMessage(32, x.x.RootWin(), manAtom,
		xproto.TimeCurrentTime, int(selAtom), int(x.x.Dummy()))
	if err != nil {
		fyne.LogError("Error creating client message", err)
		return
	}
	xproto.SendEvent(x.x.Conn(), false, x.x.RootWin(), xproto.EventMaskStructureNotify,
		string(cm.Bytes()))
}
