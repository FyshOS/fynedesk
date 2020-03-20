// +build linux

package wm // import "fyne.io/desktop/wm"

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"fyne.io/desktop"
	"fyne.io/desktop/internal/ui"

	"fyne.io/fyne"
)

type x11WM struct {
	stack
	x                     *xgbutil.XUtil
	framedExisting        bool
	moveResizing          bool
	moveResizingLastX     int16
	moveResizingLastY     int16
	moveResizingType      moveResizeType
	screenChangeTimestamp xproto.Timestamp

	allowedActions []string
	supportedHints []string

	rootIDs []xproto.Window
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
	keyCodeReturn = 36
	keyCodeAlt    = 64
	keyCodeSpace  = 65

	keyCodeEnter = 108
	keyCodeLeft  = 113
	keyCodeRight = 114
)

var focusedWin xproto.Window

// NewX11WindowManager sets up a new X11 Window Manager to control a desktop in X11.
func NewX11WindowManager(a fyne.App) (desktop.WindowManager, error) {
	conn, err := xgbutil.NewConn()
	if err != nil {
		fyne.LogError("Failed to connect to the XServer", err)
		return nil, err
	}

	mgr := &x11WM{x: conn}
	root := conn.RootWin()
	mgr.takeSelectionOwnership()

	eventMask := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskKeyPress |
		xproto.EventMaskVisibilityChange |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect
	if err := xproto.ChangeWindowAttributesChecked(conn.Conn(), root, xproto.CwEventMask,
		[]uint32{uint32(eventMask)}).Check(); err != nil {
		conn.Conn().Close()

		return nil, errors.New("window manager detected, running embedded")
	}

	mgr.allowedActions = []string{
		"_NET_WM_ACTION_MOVE",
		"_NET_WM_ACTION_RESIZE",
		"_NET_WM_ACTION_MINIMIZE",
		"_NET_WM_ACTION_MAXIMIZE_HORZ",
		"_NET_WM_ACTION_MAXIMIZE_VERT",
		"_NET_WM_ACTION_CLOSE",
		"_NET_WM_ACTION_FULLSCREEN",
	}

	mgr.supportedHints = append(mgr.supportedHints, mgr.allowedActions...)
	mgr.supportedHints = append(mgr.supportedHints, "_NET_SUPPORTED",
		"_NET_CLIENT_LIST",
		"_NET_CLIENT_LIST_STACKING",
		"_NET_WM_STATE",
		"_NET_WM_STATE_MAXIMIZED_VERT",
		"_NET_WM_STATE_MAXIMIZED_HORZ",
		"_NET_WM_STATE_SKIP_TASKBAR",
		"_NET_WM_STATE_SKIP_PAGER",
		"_NET_WM_STATE_HIDDEN",
		"_NET_WM_STATE_FULLSCREEN",
		"_NET_FRAME_EXTENTS",
		"_NET_WM_MOVERESIZE",
		"_NET_WM_NAME",
		"_NET_WM_FULLSCREEN_MONITORS",
		"_NET_MOVERESIZE_WINDOW")

	ewmh.SupportedSet(mgr.x, mgr.supportedHints)
	ewmh.SupportingWmCheckSet(mgr.x, mgr.x.RootWin(), mgr.x.Dummy())
	ewmh.SupportingWmCheckSet(mgr.x, mgr.x.Dummy(), mgr.x.Dummy())
	ewmh.WmNameSet(mgr.x, mgr.x.Dummy(), ui.RootWindowName)

	loadCursors(conn)
	go mgr.runLoop()

	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	go func() {
		for {
			<-listener
			for _, c := range mgr.clients {
				c.(*client).frame.updateScale()
			}
			mgr.configureRoots(mgr.x.RootWin())
		}
	}()

	return mgr, nil
}

func (x *x11WM) AddStackListener(l desktop.StackListener) {
	x.stack.listeners = append(x.stack.listeners, l)
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
		child.(*client).frame.unFrame()
	}

	x.x.Conn().Close()
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
		case xproto.CreateNotifyEvent:
			x.setInitialWindowAttributes(ev.Window)
		case xproto.ConfigureNotifyEvent:
			x.configureRoots(ev.Window)
		case xproto.ConfigureRequestEvent:
			x.configureWindow(ev.Window, ev)
		case xproto.DestroyNotifyEvent:
			x.destroyWindow(ev.Window)
		case xproto.EnterNotifyEvent:
			x.handleMouseEnter(ev)
		case xproto.ExposeEvent:
			x.exposeWindow(ev.Window)
		case xproto.KeyPressEvent:
			x.handleKeyPress(ev)
		case xproto.KeyReleaseEvent:
			x.handleKeyRelease(ev)
		case xproto.LeaveNotifyEvent:
			x.handleMouseLeave(ev)
		case xproto.MapRequestEvent:
			x.showWindow(ev.Window)
		case xproto.MotionNotifyEvent:
			x.handleMouseMotion(ev)
		case xproto.PropertyNotifyEvent:
			x.handlePropertyChange(ev)
		case randr.ScreenChangeNotifyEvent:
			x.handleScreenChange(ev.Timestamp)
		case xproto.UnmapNotifyEvent:
			x.hideWindow(ev.Window)
		}
	}

	fyne.LogError("X11 connection terminated!", nil)
}

func (x *x11WM) bindKeys(win xproto.Window) {
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMask1, keyCodeSpace, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMask1, keyCodeTab, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMaskShift|xproto.ModMask1, keyCodeTab, xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (x *x11WM) configureRoots(win xproto.Window) {
	if win != x.x.RootWin() || desktop.Instance() == nil {
		return
	}
	for _, screen := range desktop.Instance().Screens().Screens() {
		win := x.getWindowFromScreenName(screen.Name)
		if win == 0 {
			continue
		}
		xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(screen.X), uint32(screen.Y), uint32(screen.Width), uint32(screen.Height)}).Check()
		notifyEv := xproto.ConfigureNotifyEvent{Event: win, Window: win, AboveSibling: 0,
			X: int16(screen.X), Y: int16(screen.Y), Width: uint16(screen.Width), Height: uint16(screen.Height),
			BorderWidth: 0, OverrideRedirect: false}
		xproto.SendEvent(x.x.Conn(), false, win, xproto.EventMaskStructureNotify, string(notifyEv.Bytes()))
	}
}

func (x *x11WM) configureWindow(win xproto.Window, ev xproto.ConfigureRequestEvent) {
	c := x.clientForWin(win)
	xcoord := ev.X
	ycoord := ev.Y
	width := ev.Width
	height := ev.Height

	if c != nil {
		f := c.(*client).frame
		if f != nil && c.(*client).win == win { // ignore requests from our frame as we must have caused it
			f.minWidth, f.minHeight = windowMinSize(x.x, win)
			if c.Decorated() {
				if !c.Fullscreened() {
					c.(*client).setWindowGeometry(f.x, f.y, ev.Width+(f.borderWidth()*2), ev.Height+(f.borderWidth()+f.titleHeight()))
				} else {
					c.(*client).setWindowGeometry(f.x, f.y, ev.Width, ev.Height)
				}
			} else {
				if ev.X == 0 && ev.Y == 0 {
					ev.X = f.x
					ev.Y = f.y
				}
				c.(*client).setWindowGeometry(ev.X, ev.Y, ev.Width, ev.Height)
			}
		}
		return
	}

	name := windowName(x.x, win)
	for _, screen := range desktop.Instance().Screens().Screens() {
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
	x.RemoveWindow(c)
	windowClientListUpdate(x)
	windowClientListStackingUpdate(x)
}

func (x *x11WM) exposeWindow(win xproto.Window) {
	border := x.clientForWin(win)
	if border != nil && border.(*client).frame != nil {
		border.(*client).frame.applyTheme(false)
	}
}

func (x *x11WM) frameExisting() {
	tree, err := xproto.QueryTree(x.x.Conn(), x.x.RootWin()).Reply()
	if err != nil {
		fyne.LogError("Query Tree Error", err)
		return
	}

	for _, child := range tree.Children {
		name := windowName(x.x, child)
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

func (x *x11WM) getWindowFromScreenName(screenName string) xproto.Window {
	for _, id := range x.rootIDs {
		name := windowName(x.x, id)
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
	xproto.UnmapWindow(x.x.Conn(), c.(*client).id)
}

func (x *x11WM) isRootTitle(title string) bool {
	return strings.Index(title, ui.RootWindowName) == 0
}

func (x *x11WM) scaleToPixels(i int, screen *desktop.Screen) uint16 {
	return uint16(float32(i) * screen.CanvasScale())
}

func screenNameFromRootTitle(title string) string {
	if len(title) <= len(ui.RootWindowName) {
		return ""
	}
	return title[len(ui.RootWindowName):]
}

func (x *x11WM) setInitialWindowAttributes(win xproto.Window) {
	err := xproto.ChangeWindowAttributesChecked(x.x.Conn(), win, xproto.CwCursor,
		[]uint32{uint32(defaultCursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (x *x11WM) setupWindow(win xproto.Window) {
	if windowName(x.x, win) == "" {
		windowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_TASKBAR")
		windowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_PAGER")
	}
	c := x.clientForWin(win)
	if c != nil {
		return
	}
	c = newClient(win, x)
	x.AddWindow(c)
	c.RaiseToTop()
	c.Focus()
	windowClientListUpdate(x)
	windowClientListStackingUpdate(x)
}

func (x *x11WM) showWindow(win xproto.Window) {
	name := windowName(x.x, win)
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
	override := windowOverrideGet(x.x, win)
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
