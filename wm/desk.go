// +build linux

package wm // import "fyne.io/desktop/wm"

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"fyne.io/desktop"
	"fyne.io/desktop/internal/notify"
	"fyne.io/desktop/internal/ui"

	"fyne.io/fyne"
)

type x11WM struct {
	stack
	x                 *xgbutil.XUtil
	framedExisting    bool
	moveResizing      bool
	moveResizingLastX int16
	moveResizingLastY int16
	moveResizingType  moveResizeType
	altTabList        []desktop.Window
	altTabIndex       int

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

	keyCodeTab   = 23
	keyCodeAlt   = 64
	keyCodeSpace = 65
)

func (x *x11WM) Close() {
	log.Println("Disconnecting from X server")

	for _, child := range x.clients {
		child.(*client).frame.unFrame()
	}

	x.x.Conn().Close()
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

// NewX11WindowManager sets up a new X11 Window Manager to control a desktop in X11.
func NewX11WindowManager(a fyne.App) (desktop.WindowManager, error) {
	conn, err := xgbutil.NewConn()
	if err != nil {
		fyne.LogError("Failed to connect to the XServer", err)
		return nil, err
	}

	mgr := &x11WM{x: conn}
	root := conn.RootWin()
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
			mgr.layoutRoots()
		}
	}()

	return mgr, nil
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
		case xproto.MapRequestEvent:
			x.showWindow(ev.Window)
		case xproto.UnmapNotifyEvent:
			x.hideWindow(ev.Window)
		case xproto.ConfigureRequestEvent:
			x.configureWindow(ev.Window, ev)
		case xproto.ConfigureNotifyEvent:
			if ev.Window != x.x.RootWin() || desktop.Instance() == nil {
				break
			}
			x.layoutRoots()
		case xproto.CreateNotifyEvent:
			err := xproto.ChangeWindowAttributesChecked(x.x.Conn(), ev.Window, xproto.CwCursor,
				[]uint32{uint32(defaultCursor)}).Check()
			if err != nil {
				fyne.LogError("Set Cursor Error", err)
			}
		case xproto.DestroyNotifyEvent:
			x.destroyWindow(ev.Window)
		case xproto.PropertyNotifyEvent:
			x.handlePropertyChange(ev)
		case xproto.ClientMessageEvent:
			x.handleClientMessage(ev)
		case xproto.ExposeEvent:
			border := x.clientForWin(ev.Window)
			if border != nil && border.(*client).frame != nil {
				border.(*client).frame.applyTheme(false)
			}
		case xproto.ButtonPressEvent:
			for _, c := range x.clients {
				if c.(*client).id == ev.Event {
					c.(*client).frame.press(ev.RootX, ev.RootY)
				}
			}
			xevent.ReplayPointer(x.x)
		case xproto.ButtonReleaseEvent:
			for _, c := range x.clients {
				if c.(*client).id == ev.Event {
					if !x.moveResizing {
						c.(*client).frame.release(ev.RootX, ev.RootY)
					}
					x.moveResizeEnd(c.(*client))
				}
			}
		case xproto.MotionNotifyEvent:
			for _, c := range x.clients {
				if c.(*client).id == ev.Event {
					if x.moveResizing {
						x.moveResize(ev.RootX, ev.RootY, c.(*client))
						break
					}
					if ev.State&xproto.ButtonMask1 != 0 {
						c.(*client).frame.drag(ev.RootX, ev.RootY)
					} else {
						c.(*client).frame.motion(ev.RootX, ev.RootY)
					}
					break
				}
			}
		case xproto.EnterNotifyEvent:
			err := xproto.ChangeWindowAttributesChecked(x.x.Conn(), ev.Event, xproto.CwCursor,
				[]uint32{uint32(defaultCursor)}).Check()
			if err != nil {
				fyne.LogError("Set Cursor Error", err)
			}
			if mouseNotify, ok := desktop.Instance().(notify.MouseNotify); ok {
				mouseNotify.MouseOutNotify()
			}
		case xproto.LeaveNotifyEvent:
			if mouseNotify, ok := desktop.Instance().(notify.MouseNotify); ok {
				screen := desktop.Instance().Screens().ScreenForGeometry(int(ev.RootX), int(ev.RootY), 0, 0)
				mouseNotify.MouseInNotify(fyne.NewPos(int(float32(ev.RootX)/screen.CanvasScale()),
					int(float32(ev.RootY)/screen.CanvasScale())))
			}
		case xproto.KeyPressEvent:
			if ev.Detail == keyCodeSpace {
				go ui.ShowAppLauncher()
				break
			} else if ev.Detail != keyCodeTab {
				break
			}
			if x.altTabList == nil {
				x.altTabList = []desktop.Window{}
				for _, win := range x.Windows() {
					if win.Iconic() {
						continue
					}
					x.altTabList = append(x.altTabList, win)
				}
				x.altTabIndex = 0

				xproto.GrabKeyboard(x.x.Conn(), true, x.x.RootWin(), xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)
			}

			winCount := len(x.altTabList)
			if winCount <= 1 {
				break
			}
			if ev.State&xproto.ModMaskShift != 0 {
				x.altTabIndex--
				if x.altTabIndex < 0 {
					x.altTabIndex = winCount - 1
				}
			} else {
				x.altTabIndex++
				if x.altTabIndex == winCount {
					x.altTabIndex = 0
				}
			}

			x.RaiseToTop(x.altTabList[x.altTabIndex])
			windowClientListStackingUpdate(x)
		case xproto.KeyReleaseEvent:
			if ev.Detail == keyCodeAlt {
				x.altTabList = nil
				xproto.UngrabKeyboard(x.x.Conn(), xproto.TimeCurrentTime)
			}
		}
	}

	fyne.LogError("X11 connection terminated!", nil)
}

func (x *x11WM) getWindowFromName(screenName string) xproto.Window {
	for _, id := range x.rootIDs {
		name := windowName(x.x, id)
		pos := strings.LastIndex(name, ui.RootWindowName) + len(ui.RootWindowName) + 1
		outputName := name[pos:]
		if outputName == screenName {
			return id
		}
	}
	return 0
}

func (x *x11WM) layoutRoots() {
	for _, screen := range desktop.Instance().Screens().Screens() {
		win := x.getWindowFromName(screen.Name)
		if win != 0 {
			xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
				[]uint32{uint32(screen.X), uint32(screen.Y), uint32(screen.Width), uint32(screen.Height)}).Check()
			notifyEv := xproto.ConfigureNotifyEvent{Event: win, Window: win, AboveSibling: 0,
				X: int16(screen.X), Y: int16(screen.Y), Width: uint16(screen.Width), Height: uint16(screen.Height),
				BorderWidth: 0, OverrideRedirect: false}
			xproto.SendEvent(x.x.Conn(), false, win, xproto.EventMaskStructureNotify, string(notifyEv.Bytes()))
		}
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
		if len(name) <= len(ui.RootWindowName) {
			continue
		}
		pos := strings.Index(ui.RootWindowName, name) + len(ui.RootWindowName) + 1
		outputName := name[pos:]
		if outputName == screen.Name {
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
	}

	err := xproto.ConfigureWindowChecked(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(xcoord), uint32(ycoord), uint32(width), uint32(height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (x *x11WM) handlePropertyChange(ev xproto.PropertyNotifyEvent) {
	c := x.clientForWin(ev.Window)
	if c == nil {
		return
	}
	propAtom, err := xprop.AtomName(x.x, ev.Atom)
	if err != nil {
		fyne.LogError("Error getting event", err)
		return
	}
	switch propAtom {
	case "_NET_WM_NAME":
		c.(*client).updateTitle()
	case "WM_NAME":
		c.(*client).updateTitle()
	case "_MOTIF_WM_HINTS":
		c.(*client).setupBorder()
	}
}

func (x *x11WM) moveResizeEnd(c *client) {
	x.moveResizing = false
	xproto.UngrabPointer(x.x.Conn(), xproto.TimeCurrentTime)
	xproto.UngrabKeyboard(x.x.Conn(), xproto.TimeCurrentTime)

	// ensure menus etc update
	c.frame.notifyInnerGeometry()
}

func (x *x11WM) moveResize(moveX, moveY int16, c *client) {
	xcoord, ycoord, width, height := c.getWindowGeometry()
	w := int16(width)
	h := int16(height)
	deltaW := moveX - x.moveResizingLastX
	deltaH := moveY - x.moveResizingLastY

	switch x.moveResizingType {
	case moveResizeTopLeft:
		//Move both X,Y coords and resize both W,H
		xcoord += deltaW
		ycoord += deltaH
		w -= deltaW
		h -= deltaH
	case moveResizeTop:
		//Move Y coord and resize H
		ycoord += deltaH
		h -= deltaH
	case moveResizeTopRight:
		//Move Y coord and resize both W,H
		ycoord += deltaH
		w += deltaW
		h -= deltaH
	case moveResizeRight:
		//Keep X coord and resize W
		w += deltaW
	case moveResizeBottomRight, moveResizeKeyboard:
		//Keep both X,Y coords and resize both W,H
		w += deltaW
		h += deltaH
	case moveResizeBottom:
		//Keep Y coord and resize H
		h += deltaH
	case moveResizeBottomLeft:
		//Move X coord and resize both W,H
		xcoord += deltaW
		w -= deltaW
		h += deltaH
	case moveResizeLeft:
		//Move X coord and resize W
		xcoord += deltaW
		w -= deltaW
	case moveResizeMove, moveResizeMoveKeyboard:
		//Move both X,Y coords and no resize
		xcoord += deltaW
		ycoord += deltaH
	case moveResizeCancel:
		x.moveResizeEnd(c)
	}
	x.moveResizingLastX = moveX
	x.moveResizingLastY = moveY
	c.setWindowGeometry(xcoord, ycoord, uint16(w), uint16(h))
}

func (x *x11WM) handleMoveResize(ev xproto.ClientMessageEvent, c *client) {
	x.moveResizing = true
	x.moveResizingLastX = int16(ev.Data.Data32[0])
	x.moveResizingLastY = int16(ev.Data.Data32[1])
	x.moveResizingType = moveResizeType(ev.Data.Data32[2])
	xproto.GrabPointer(x.x.Conn(), true, c.id,
		xproto.EventMaskButtonPress|xproto.EventMaskButtonRelease|xproto.EventMaskPointerMotion,
		xproto.GrabModeAsync, xproto.GrabModeAsync, x.x.RootWin(), xproto.CursorNone, xproto.TimeCurrentTime)
	xproto.GrabKeyboard(x.x.Conn(), true, c.id, xproto.TimeCurrentTime, xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (x *x11WM) handleStateActionRequest(ev xproto.ClientMessageEvent, removeState func(), addState func(), toggleCheck bool) {
	switch clientMessageStateAction(ev.Data.Data32[0]) {
	case clientMessageStateActionRemove:
		removeState()
	case clientMessageStateActionAdd:
		addState()
	case clientMessageStateActionToggle:
		if toggleCheck {
			removeState()
		} else {
			addState()
		}
	}
}

func (x *x11WM) handleInitialHints(ev xproto.ClientMessageEvent, hint string) {
	switch clientMessageStateAction(ev.Data.Data32[0]) {
	case clientMessageStateActionRemove:
		windowExtendedHintsRemove(x.x, ev.Window, hint)
		x.showWindow(ev.Window)
	case clientMessageStateActionAdd:
		windowExtendedHintsAdd(x.x, ev.Window, hint)
	}
}

func (x *x11WM) handleClientMessage(ev xproto.ClientMessageEvent) {
	c := x.clientForWin(ev.Window)
	msgAtom, err := xprop.AtomName(x.x, ev.Type)
	if err != nil {
		fyne.LogError("Error getting event", err)
		return
	}
	switch msgAtom {
	case "WM_CHANGE_STATE":
		if c == nil {
			return
		}
		switch ev.Data.Bytes()[0] {
		case icccm.StateIconic:
			c.(*client).iconifyClient()
		case icccm.StateNormal:
			c.(*client).uniconifyClient()
		}
	case "_NET_ACTIVE_WINDOW":
		if c == nil {
			return
		}
		xproto.SetInputFocus(x.x.Conn(), 0, ev.Window, 0)
		windowActiveSet(x.x, ev.Window)
	case "_NET_WM_FULLSCREEN_MONITORS":
		// TODO WHEN WE SUPPORT MULTI-MONITORS - THIS TELLS WHICH/HOW MANY MONITORS
		// TO FULLSCREEN ACROSS
	case "_NET_WM_MOVERESIZE":
		if c == nil {
			return
		}
		if c.Maximized() || c.Fullscreened() {
			return
		}
		x.handleMoveResize(ev, c.(*client))
	case "_NET_WM_STATE":
		subMsgAtom, err := xprop.AtomName(x.x, xproto.Atom(ev.Data.Data32[1]))
		if err != nil {
			fyne.LogError("Error getting event", err)
			return
		}
		if c == nil {
			x.handleInitialHints(ev, subMsgAtom)
			return
		}
		switch subMsgAtom {
		case "_NET_WM_STATE_FULLSCREEN":
			x.handleStateActionRequest(ev, c.(*client).unfullscreenClient, c.(*client).fullscreenClient, c.Fullscreened())
		case "_NET_WM_STATE_HIDDEN":
			fyne.LogError("Extended Window Manager Hints says to ignore the HIDDEN state.", nil)
		//	x.handleStateActionRequest(ev, c.(*client).uniconifyClient, c.(*client).iconifyClient, c.Iconic())
		case "_NET_WM_STATE_MAXIMIZED_VERT", "_NET_WM_STATE_MAXIMIZED_HORZ":
			extraMsgAtom, err := xprop.AtomName(x.x, xproto.Atom(ev.Data.Data32[2]))
			if err != nil {
				fyne.LogError("Error getting event", err)
				return
			}
			if extraMsgAtom == subMsgAtom {
				return
			}
			if extraMsgAtom == "_NET_WM_STATE_MAXIMIZED_VERT" || extraMsgAtom == "_NET_WM_STATE_MAXIMIZED_HORZ" {
				x.handleStateActionRequest(ev, c.(*client).unmaximizeClient, c.(*client).maximizeClient, c.Maximized())
			}
		}
	}
}

func (x *x11WM) showWindow(win xproto.Window) {
	name := windowName(x.x, win)
	if strings.Index(name, ui.RootWindowName) == 0 {
		err := xproto.MapWindowChecked(x.x.Conn(), win).Check()
		if err != nil {
			fyne.LogError("Show Window Error", err)
		}
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

func (x *x11WM) hideWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		return
	}

	xproto.UnmapWindow(x.x.Conn(), c.(*client).id)
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

func (x *x11WM) destroyWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil {
		return
	}
	x.RemoveWindow(c)
	windowClientListUpdate(x)
	windowClientListStackingUpdate(x)
}

func (x *x11WM) bindKeys(win xproto.Window) {
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMask1, keyCodeSpace, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMask1, keyCodeTab, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, xproto.ModMaskShift|xproto.ModMask1, keyCodeTab, xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (x *x11WM) frameExisting() {
	tree, err := xproto.QueryTree(x.x.Conn(), x.x.RootWin()).Reply()
	if err != nil {
		fyne.LogError("Query Tree Error", err)
		return
	}

	for _, child := range tree.Children {
		name := windowName(x.x, child)
		if strings.Index(name, ui.RootWindowName) == 0 {
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

func (x *x11WM) scaleToPixels(i int, screen *desktop.Screen) uint16 {
	return uint16(float32(i) * screen.CanvasScale())
}
