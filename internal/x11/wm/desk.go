//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package wm // import "fyshos.com/fynedesk/internal/x11/wm"

import (
	"errors"
	"image"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/FyshOS/backgrounds/builtin"
	"github.com/nfnt/resize"

	"fyne.io/fyne/v2"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/internal/ui"
	"fyshos.com/fynedesk/internal/x11"
	xwin "fyshos.com/fynedesk/internal/x11/win"
	"fyshos.com/fynedesk/wm"
)

type x11WM struct {
	stack
	x                       *xgbutil.XUtil
	framedExisting          bool
	moveResizing            bool
	moveResizingX           int
	moveResizingY           int
	moveResizingStartX      int16
	moveResizingStartY      int16
	moveResizingLastX       int16
	moveResizingLastY       int16
	moveResizingStartWidth  uint
	moveResizingStartHeight uint
	moveResizingType        moveResizeType
	screenChangeTimestamp   xproto.Timestamp

	currentBindings []*fynedesk.Shortcut

	died         bool
	rootID       xproto.Window
	menuSize     fyne.Size
	menuPos      fyne.Position
	transientMap map[xproto.Window][]xproto.Window
	oldRoot      *xgraphics.Image
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

	keyCodeEscape      = 9
	keyCodeTab         = 23
	keyCodeReturn      = 36
	keyCodeBacktick    = 49
	keyCodeAlt         = 64
	keyCodeSpace       = 65
	keyCodePrintScreen = 107
	keyCodeSuper       = 133
	keyCodeCalculator  = 148

	keyCodeEnter = 108
	keyCodeLeft  = 113
	keyCodeRight = 114
	keyCodeUp    = 111
	keyCodeDown  = 116

	keyCodeBrightLess = 232
	keyCodeBrightMore = 233

	keyCodeVolumeMute = 121
	keyCodeVolumeLess = 122
	keyCodeVolumeMore = 123

	windowNameMenu = "FyneDesk Menu"
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

	err = ewmh.SupportedSet(mgr.x, x11.SupportedHints)
	if err != nil {
		fyne.LogError("", err)
	}
	err = ewmh.SupportingWmCheckSet(mgr.x, mgr.x.RootWin(), mgr.x.Dummy())
	if err != nil {
		fyne.LogError("", err)
	}
	err = ewmh.SupportingWmCheckSet(mgr.x, mgr.x.Dummy(), mgr.x.Dummy())
	if err != nil {
		fyne.LogError("", err)
	}
	err = ewmh.WmNameSet(mgr.x, mgr.x.Dummy(), ui.RootWindowName)
	if err != nil {
		fyne.LogError("", err)
	}
	err = ewmh.DesktopViewportSet(mgr.x, []ewmh.DesktopViewport{{X: 0, Y: 0}}) // Will always be 0, 0 until virtual desktops are supported
	if err != nil {
		fyne.LogError("", err)
	}
	err = ewmh.NumberOfDesktopsSet(mgr.x, 1) // Will always be 1 until virtual desktops are supported
	if err != nil {
		fyne.LogError("", err)
	}
	err = ewmh.CurrentDesktopSet(mgr.x, 0) // Will always be 0 until virtual desktops are supported
	if err != nil {
		fyne.LogError("", err)
	}

	x11.LoadCursors(conn)

	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	a.Preferences().AddChangeListener(mgr.refreshBorders)
	go func() {
		for range listener {
			mgr.updateBackgrounds()
			mgr.refreshBorders()
			mgr.configureRoots()
		}
	}()

	return mgr, nil
}

func (x *x11WM) AddStackListener(l fynedesk.StackListener) {
	x.stack.listeners = append(x.stack.listeners, l)
}

func (x *x11WM) Blank() {
	go func() {
		time.Sleep(time.Second / 3)
		err := exec.Command("xset", "-display", os.Getenv("DISPLAY"), "dpms", "force", "off").Start()
		if err != nil {
			fyne.LogError("", err)
		}
	}()
}

func (x *x11WM) Capture() image.Image {
	root := x.x.RootWin()
	return x11.CaptureWindow(x.x.Conn(), root)
}

func (x *x11WM) Close() {
	for _, child := range x.clients {
		child.Close()
	}
	if x.died {
		// x server died, no point attempting to shut it cleanly
		return
	}

	cancel := false
	exit := make(chan interface{})
	go func() {
		for !cancel && len(x.clients) > 0 {
			time.Sleep(time.Millisecond * 100)
		}

		close(exit)
	}()

	go func() {
		select {
		case <-exit:
			x.x.Conn().Close()
			os.Exit(0)
		case <-time.NewTimer(time.Second * 10).C:
			notify := wm.NewNotification("Log Out", "Log Out was cancelled by an open application")
			wm.SendNotification(notify)
			cancel = true
		}
	}()
}

func (x *x11WM) Conn() *xgb.Conn {
	return x.x.Conn()
}

func (x *x11WM) Run() {
	x.setupBindings()
	go x.runLoop()
}

func (x *x11WM) ShowOverlay(w fyne.Window, s fyne.Size, p fyne.Position) {
	w.SetTitle(windowNameMenu)
	w.SetFixedSize(true)
	w.Resize(s)

	w.Show()
	x.menuSize = s
	x.menuPos = p
}

func (x *x11WM) ShowMenuOverlay(m *fyne.Menu, s fyne.Size, p fyne.Position) {
	win := fyne.CurrentApp().Driver().(deskDriver.Driver).CreateSplashWindow()
	for _, item := range m.Items {
		action := item.Action
		item.Action = func() {
			action()
			win.Close()
		}
	}

	pop := widget.NewPopUpMenu(m, win.Canvas())
	pop.OnDismiss = win.Close
	pop.Show()
	pop.Resize(s)
	go func() {
		// TODO figure why sometimes this doesn't draw (size and minsize are correct)
		// and then remove this workaround goroutine
		time.Sleep(time.Second / 10)
		pop.Resize(s)
		time.Sleep(time.Second / 4)
		pop.Resize(s)
	}()
	x.ShowOverlay(win, s, p)
}

func (x *x11WM) ShowModal(w fyne.Window, s fyne.Size) {
	w.SetTitle(windowNameMenu)
	w.SetFixedSize(true)
	w.Resize(s)
	w.CenterOnScreen()

	w.Show()
	x.menuSize = s

	root := fynedesk.Instance().Screens().Primary()
	scale := root.CanvasScale()
	p := fyne.NewPos((float32(root.Width)/scale-s.Width)/2, (float32(root.Height)/scale-s.Height)/2)

	x.menuPos = p
}

func (x *x11WM) X() *xgbutil.XUtil {
	return x.x
}

func (x *x11WM) bindShortcut(short *fynedesk.Shortcut, win xproto.Window) {
	mask := x.modifierToKeyMask(short.Modifier)
	code := x.keyNameToCode(short.KeyName)
	if code == 0 {
		return
	}

	xproto.GrabKey(x.x.Conn(), true, win, mask, code, xproto.GrabModeAsync, xproto.GrabModeAsync)
	if mask == xproto.ModMaskAny {
		return // no need for the extra binds
	}
	xproto.GrabKey(x.x.Conn(), true, win, mask|xproto.ModMaskLock, code, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, mask|xproto.ModMask2, code, xproto.GrabModeAsync, xproto.GrabModeAsync)
	xproto.GrabKey(x.x.Conn(), true, win, mask|xproto.ModMask3, code, xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (x *x11WM) bindShortcuts(win xproto.Window) {
	if _, ok := fynedesk.Instance().(wm.ShortcutManager); !ok {
		return
	}

	shortcutList := fynedesk.Instance().(wm.ShortcutManager).Shortcuts()
	for _, shortcut := range shortcutList {
		x.bindShortcut(shortcut, win)
	}

	if x.currentBindings == nil {
		x.currentBindings = shortcutList
	}
}

func (x *x11WM) keyNameToCode(n fyne.KeyName) xproto.Keycode {
	keybind.Initialize(x.x)
	switch n {
	case fyne.KeySpace:
		return keyCodeSpace
	case fyne.KeyLeft:
		return keyCodeLeft
	case fyne.KeyRight:
		return keyCodeRight
	case fyne.KeyUp:
		return keyCodeUp
	case fyne.KeyDown:
		return keyCodeDown
	case fyne.KeyTab:
		return keyCodeTab
	case fyne.KeyBackTick:
		return keyCodeBacktick
	case deskDriver.KeyPrintScreen:
		return keyCodePrintScreen
	case fynedesk.KeyBrightnessDown:
		return keyCodeBrightLess
	case fynedesk.KeyBrightnessUp:
		return keyCodeBrightMore
	case fynedesk.KeyCalculator:
		return keyCodeCalculator
	case fynedesk.KeyVolumeMute:
		return keyCodeVolumeMute
	case fynedesk.KeyVolumeDown:
		return keyCodeVolumeLess
	case fynedesk.KeyVolumeUp:
		return keyCodeVolumeMore
	case fyne.KeyL:
		codes := keybind.StrToKeycodes(x.x, "L")
		return codes[0]
	}

	for i := 0; i <= 9; i++ {
		id := strconv.Itoa(i)
		if n == fyne.KeyName(id) {
			codes := keybind.StrToKeycodes(x.x, id)
			return codes[0]
		}
	}

	return 0
}

func (x *x11WM) modifierToKeyMask(m fyne.KeyModifier) uint16 {
	mask := uint16(0)
	if m&fynedesk.UserModifier != 0 {
		if fynedesk.Instance().Settings().KeyboardModifier() == fyne.KeyModifierAlt {
			m |= fyne.KeyModifierAlt
		} else {
			m |= fyne.KeyModifierSuper
		}
	}

	if m&fyne.KeyModifierAlt != 0 {
		mask |= xproto.ModMask1
	}
	if m&fyne.KeyModifierControl != 0 {
		mask |= xproto.ModMaskControl
	}
	if m&fyne.KeyModifierShift != 0 {
		mask |= xproto.ModMaskShift
	}
	if m&fyne.KeyModifierSuper != 0 {
		mask |= xproto.ModMask4
	}

	if mask == 0 {
		return xproto.ModMaskAny
	}
	return mask
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
			x.died = true
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
			if ev.Window == x.x.RootWin() {
				x.configureRoots()
			}
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

func (x *x11WM) configureRoots() {
	if fynedesk.Instance() == nil {
		return
	}

	x.setupX11DPIHints()
	minX, minY, maxX, maxY := math.MaxInt16, math.MaxInt16, 0, 0
	for _, screen := range fynedesk.Instance().Screens().Screens() {
		minX = min(minX, screen.X)
		minY = min(minY, screen.Y)
		maxX = max(maxX, screen.X+screen.Width)
		maxY = max(maxY, screen.Y+screen.Height)

		if screen == fynedesk.Instance().Screens().Primary() {
			priX, priY, priW, priH := 0, 0, 0, 0
			geom, err := xproto.GetGeometry(x.x.Conn(), xproto.Drawable(x.rootID)).Reply()
			if err == nil {
				priX, priY = int(geom.X), int(geom.Y)
				priW, priH = int(geom.Width), int(geom.Height)
			}
			if screen.X == priX && screen.Y == priY && screen.Width == priW && screen.Height == priH {
				continue
			}

			notifyEv := xproto.ConfigureNotifyEvent{Event: x.rootID, Window: x.rootID, AboveSibling: 0,
				X: int16(screen.X), Y: int16(screen.Y), Width: uint16(screen.Width), Height: uint16(screen.Height),
				BorderWidth: 0, OverrideRedirect: false}
			xproto.SendEvent(x.x.Conn(), false, x.rootID, xproto.EventMaskStructureNotify, string(notifyEv.Bytes()))

			// we need to trigger a move so that the correct scale is picked up
			xproto.ConfigureWindow(x.x.Conn(), x.rootID, xproto.ConfigWindowX|xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
				[]uint32{uint32(screen.X + 1), uint32(screen.Y + 1), uint32(screen.Width - 2), uint32(screen.Height - 2)})

			// and then set the correct location
			xproto.ConfigureWindow(x.x.Conn(), x.rootID, xproto.ConfigWindowX|xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
				[]uint32{uint32(screen.X), uint32(screen.Y), uint32(screen.Width), uint32(screen.Height)})
		}
	}

	rootWidth := maxX - minX
	rootHeight := maxY - minY

	err := ewmh.DesktopGeometrySet(x.x, &ewmh.DesktopGeometry{Width: rootWidth, Height: rootHeight}) // The size will grow when virtual desktops are supported
	if err != nil {
		fyne.LogError("", err)
	}

	err = ewmh.WorkareaSet(x.x, []ewmh.Workarea{{X: 0, Y: 0, Width: uint(rootWidth), Height: uint(rootHeight)}}) // The array will grow when virtual desktops are supported
	if err != nil {
		fyne.LogError("", err)
	}
	go x.updateBackgrounds()
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
	if x.isRootTitle(name) {
		x.rootID = win

		x.configureRoots() // we added a root window, so reconfigure
		return
	}
	xproto.ConfigureWindow(x.x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(xcoord), uint32(ycoord), uint32(width), uint32(height)})
}

func (x *x11WM) destroyWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil || win == c.FrameID() {
		return
	}
	transient := x11.WindowTransientForGet(x.x, win)
	if transient > 0 && transient != win {
		x.transientChildRemove(transient, win)
	} else if transient > 0 && transient == win {
		x.transientLeaderRemove(transient)
	}
	windowClientListUpdate(x)
	windowClientListStackingUpdate(x)

	xproto.DestroyWindow(x.x.Conn(), c.FrameID())
	xproto.DestroyWindow(x.x.Conn(), c.ChildID())
}

func (x *x11WM) exposeWindow(win xproto.Window) {
	attrs, err := xproto.GetWindowAttributes(x.x.Conn(), win).Reply()
	if err == nil && attrs.MapState == xproto.MapStateUnmapped { // ignore expose for windows closing
		return
	}

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

func (x *x11WM) RootID() xproto.Window {
	return x.rootID
}

func (x *x11WM) NotifyWindowMoved(win fynedesk.Window) {
	for _, l := range x.listeners {
		go l.WindowMoved(win)
	}
}

func (x *x11WM) hideWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c == nil || win == c.FrameID() {
		return
	}
	xproto.UnmapWindow(x.x.Conn(), c.FrameID())
	if !c.Iconic() {
		x.RemoveWindow(c)
	}
}

func (x *x11WM) isRootTitle(title string) bool {
	return strings.Index(title, ui.RootWindowName) == 0
}

func (x *x11WM) refreshBorders() {
	for _, c := range x.clients {
		c.(x11.XWin).SettingsChanged()
	}
}

func screenNameFromRootTitle(title string) string {
	if len(title) <= len(ui.RootWindowName) {
		return ""
	}
	return title[len(ui.RootWindowName):]
}

func (x *x11WM) setActiveScreenFromWindow(win fynedesk.Window) {
	if win == nil || fynedesk.Instance() == nil {
		return
	}

	windowScreen := fynedesk.Instance().Screens().ScreenForWindow(win)
	if windowScreen != nil {
		fynedesk.Instance().Screens().SetActive(windowScreen)
	}
}

func (x *x11WM) setInitialWindowAttributes(win xproto.Window) {
	xproto.ChangeWindowAttributes(x.x.Conn(), win, xproto.CwCursor,
		[]uint32{uint32(x11.DefaultCursor)})
}

func (x *x11WM) setupBindings() {
	deskListener := make(chan fynedesk.DeskSettings)
	fynedesk.Instance().Settings().AddChangeListener(deskListener)
	go func() {
		for range deskListener {
			// this uses the state from the previous bind call
			x.unbindShortcuts(x.rootID)
			for _, c := range x.clients {
				x.unbindShortcuts(c.(x11.XWin).ChildID())
			}
			x.currentBindings = nil

			// this call sets up the new cache of shortcuts
			x.bindShortcuts(x.rootID)
			for _, c := range x.clients {
				x.bindShortcuts(c.(x11.XWin).ChildID())
			}

			go x.updateBackgrounds()
		}
	}()
}

func (x *x11WM) setupWindow(win xproto.Window) {
	c := x.clientForWin(win)
	if c != nil {
		return
	}
	c = xwin.NewClient(win, x)
	if c == nil {
		return // a previous reported problem occurred framing the window
	}

	x.bindShortcuts(win)
	if x11.WindowName(x.x, win) == "" {
		x11.WindowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_TASKBAR")
		x11.WindowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_PAGER")
	}
	x.AddWindow(c)
	c.RaiseToTop()
	c.Focus()
	windowClientListUpdate(x)
	windowClientListStackingUpdate(x)
}

func (x *x11WM) setupX11DPIHints() {
	// TODO move from global once xrandr --dpi <dpi>/<output> is better supported
	canvasScale := fynedesk.Instance().Screens().Primary().CanvasScale()
	dpi := int(float32(baselineDPI) * canvasScale)
	cmd := exec.Command("xrandr", "--dpi", strconv.Itoa(dpi))
	_ = cmd.Start() // if it fails that's a shame but it's just info
}

func (x *x11WM) showWindow(win xproto.Window, parent xproto.Window) {
	name := x11.WindowName(x.x, win)
	if x.isRootTitle(name) {
		err := xproto.MapWindowChecked(x.x.Conn(), win).Check()
		if err != nil {
			fyne.LogError("Show Window Error", err)
		}
		xproto.ConfigureWindow(x.x.Conn(), win, xproto.ConfigWindowStackMode, []uint32{xproto.StackModeBelow})
		_ = ewmh.WmWindowTypeSet(x.x, win, []string{windowTypeDesktop})
		x.bindShortcuts(win)
		if !x.framedExisting {
			x.framedExisting = true
			go x.frameExisting()
		}
		return
	}
	if name == windowNameMenu {
		x11.WindowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_TASKBAR")
		x11.WindowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_PAGER")
		xproto.ChangeWindowAttributes(x.Conn(), win, xproto.CwEventMask, []uint32{xproto.EventMaskLeaveWindow})

		screen := fynedesk.Instance().Screens().Primary()
		w, h := x.menuSize.Width*screen.CanvasScale(), x.menuSize.Height*screen.CanvasScale()
		mx, my := screen.X+int(x.menuPos.X*screen.CanvasScale()), screen.Y+int(x.menuPos.Y*screen.CanvasScale())
		xproto.ConfigureWindow(x.Conn(), win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight, []uint32{uint32(mx), uint32(my),
			uint32(w), uint32(h)})

		x.bindShortcuts(win)
		xproto.MapWindow(x.Conn(), win)
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
	switch winType[len(winType)-1] { // KDE etc put their window types first
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
	name := "WM_S" + strconv.Itoa(x.x.Conn().DefaultScreen)
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

func (x *x11WM) unbindShortcut(short *fynedesk.Shortcut, win xproto.Window) {
	mask := x.modifierToKeyMask(short.Modifier)
	code := x.keyNameToCode(short.KeyName)
	if code == 0 {
		return
	}

	xproto.UngrabKey(x.x.Conn(), code, win, mask)
	xproto.UngrabKey(x.x.Conn(), code, win, mask|xproto.ModMaskLock)
	xproto.UngrabKey(x.x.Conn(), code, win, mask|xproto.ModMask2)
	xproto.UngrabKey(x.x.Conn(), code, win, mask|xproto.ModMask3)
}

func (x *x11WM) unbindShortcuts(win xproto.Window) {
	if _, ok := fynedesk.Instance().(wm.ShortcutManager); !ok {
		return
	}

	for _, shortcut := range x.currentBindings {
		x.unbindShortcut(shortcut, win)
	}
}

func (x *x11WM) updatedBackgroundImage(w, h int) image.Image {
	path := fynedesk.Instance().Settings().Background()
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			fyne.LogError("Failed to open background image", err)
		} else {
			img, _, err := image.Decode(file)
			if err != nil {
				fyne.LogError("Failed to read background image", err)
			} else {
				_ = file.Close()
				return resize.Resize(uint(w), uint(h), img, resize.Lanczos3)
			}
		}
	}

	set := fyne.CurrentApp().Settings()
	b := &builtin.Builtin{}
	c := software.NewCanvas()
	c.SetContent(b.Load(set.Theme(), set.ThemeVariant()))
	c.SetScale(1.0)
	c.Resize(fyne.NewSize(float32(w), float32(h)))
	return c.Capture()
}

func (x *x11WM) updateBackgrounds() {
	geom, err := xproto.GetGeometry(x.x.Conn(), xproto.Drawable(x.x.RootWin())).Reply()
	if err != nil {
		fyne.LogError("Unable to look up root geometry", err)
		return
	}
	root := xgraphics.New(x.x, image.Rect(0, 0, int(geom.Width), int(geom.Height)))

	for _, screen := range fynedesk.Instance().Screens().Screens() {
		scaled := x.updatedBackgroundImage(screen.Width, screen.Height)
		for y := screen.Y; y < screen.Y+screen.Height; y++ {
			for x := screen.X; x < screen.X+screen.Width; x++ {
				root.Set(x, y, scaled.At(x-screen.X, y-screen.Y))
			}
		}
	}

	err = root.XSurfaceSet(x.x.RootWin())
	if err != nil {
		fyne.LogError("", err)
	}
	root.XDraw()
	root.XPaint(x.x.RootWin())

	if x.oldRoot != nil {
		x.oldRoot.Destroy()
		x.oldRoot = nil
	}

	err = xprop.ChangeProp32(x.x, x.x.RootWin(), "_XROOTPMAP_ID", "PIXMAP", uint(root.Pixmap))
	if err != nil {
		fyne.LogError("rootprop", err)
	}
	err = xprop.ChangeProp32(x.x, x.x.RootWin(), "ESETROOT_PMAP_ID", "PIXMAP", uint(root.Pixmap))
	if err != nil {
		fyne.LogError("esetrootprop", err)
	}

	// save root so we can free it later if not needed
	x.oldRoot = root
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
