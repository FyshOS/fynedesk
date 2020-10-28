// +build linux openbsd freebsd netbsd

package wm // import "fyne.io/fynedesk/internal/x11/wm"

import (
	"bytes"
	"errors"
	"fmt"
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
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/nfnt/resize"

	"fyne.io/fyne"
	deskDriver "fyne.io/fyne/driver/desktop"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/ui"
	"fyne.io/fynedesk/internal/x11"
	xwin "fyne.io/fynedesk/internal/x11/win"
	wmTheme "fyne.io/fynedesk/theme"
	"fyne.io/fynedesk/wm"
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

	allowedActions  []string
	supportedHints  []string
	currentBindings []*fynedesk.Shortcut

	died         bool
	rootID       xproto.Window
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
	keyCodeReturn = 36
	keyCodeAlt    = 64
	keyCodeSpace  = 65

	keyCodeEnter = 108
	keyCodeLeft  = 113
	keyCodeRight = 114

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
	mgr.setupX11DPIForScale(a.Settings().Scale())

	listener := make(chan fyne.Settings)
	a.Settings().AddChangeListener(listener)
	go func() {
		for {
			s := <-listener
			for _, c := range mgr.clients {
				c.(x11.XWin).SettingsChanged()
			}
			mgr.setupX11DPIForScale(s.Scale())
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
		exec.Command("xset", "-display", os.Getenv("DISPLAY"), "dpms", "force", "off").Start()
	}()
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
	go x.updateBackgrounds()
	go x.runLoop()
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
	switch n {
	case fyne.KeySpace:
		return keyCodeSpace
	case fyne.KeyTab:
		return keyCodeTab
	case fynedesk.KeyBrightnessDown:
		return keyCodeBrightLess
	case fynedesk.KeyBrightnessUp:
		return keyCodeBrightMore
	}

	return 0
}

func (x *x11WM) modifierToKeyMask(m deskDriver.Modifier) uint16 {
	mask := uint16(0)
	if m&deskDriver.AltModifier != 0 {
		mask |= xproto.ModMask1
	}
	if m&deskDriver.ControlModifier != 0 {
		mask |= xproto.ModMaskControl
	}
	if m&deskDriver.ShiftModifier != 0 {
		mask |= xproto.ModMaskShift
	}

	if mask == 0 {
		return xproto.ModMaskAny
	}
	return mask
}

func (x *x11WM) runLoop() {
	x.setupBindings()
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

	minX, minY, maxX, maxY := math.MaxInt16, math.MaxInt16, 0, 0
	for _, screen := range fynedesk.Instance().Screens().Screens() {
		minX = fyne.Min(minX, screen.X)
		minY = fyne.Min(minY, screen.Y)
		maxX = fyne.Max(maxX, screen.X+screen.Width)
		maxY = fyne.Max(maxY, screen.Y+screen.Height)

		if screen == fynedesk.Instance().Screens().Primary() {
			notifyEv := xproto.ConfigureNotifyEvent{Event: x.rootID, Window: x.rootID, AboveSibling: 0,
				X: int16(screen.X), Y: int16(screen.Y), Width: uint16(screen.Width), Height: uint16(screen.Height),
				BorderWidth: 0, OverrideRedirect: false}
			xproto.SendEvent(x.x.Conn(), false, x.rootID, xproto.EventMaskStructureNotify, string(notifyEv.Bytes()))

			err := xproto.ConfigureWindowChecked(x.x.Conn(), x.rootID, xproto.ConfigWindowX|xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
				[]uint32{uint32(screen.X), uint32(screen.Y), uint32(screen.Width), uint32(screen.Height)}).Check()
			if err != nil {
				fyne.LogError("Configure Window Error", err)
			}
		}
	}

	rootWidth := maxX - minX
	rootHeight := maxY - minY

	ewmh.DesktopGeometrySet(x.x, &ewmh.DesktopGeometry{Width: rootWidth, Height: rootHeight})              // The size will grow when virtual desktops are supported
	ewmh.WorkareaSet(x.x, []ewmh.Workarea{{X: 0, Y: 0, Width: uint(rootWidth), Height: uint(rootHeight)}}) // The array will grow when virtual desktops are supported
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
	err := xproto.ChangeWindowAttributesChecked(x.x.Conn(), win, xproto.CwCursor,
		[]uint32{uint32(x11.DefaultCursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (x *x11WM) setupBindings() {
	deskListener := make(chan fynedesk.DeskSettings)
	fynedesk.Instance().Settings().AddChangeListener(deskListener)
	go func() {
		for {
			<-deskListener
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
	if x11.WindowName(x.x, win) == "" {
		x11.WindowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_TASKBAR")
		x11.WindowExtendedHintsAdd(x.x, win, "_NET_WM_STATE_SKIP_PAGER")
	}
	c := x.clientForWin(win)
	if c != nil {
		return
	}
	c = xwin.NewClient(win, x)
	x.bindShortcuts(win)
	x.AddWindow(c)
	c.RaiseToTop()
	c.Focus()
	windowClientListUpdate(x)
	windowClientListStackingUpdate(x)
}

func (x *x11WM) setupX11DPIForScale(scale float32) {
	cmd := exec.Command("xrandr", "--dpi", strconv.Itoa(int(float32(baselineDPI)*scale)))
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
		x.bindShortcuts(win)
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

func (x *x11WM) updatedBackgroundImage() image.Image {
	path := fynedesk.Instance().Settings().Background()
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			fyne.LogError("Failed to open background image", err)
		}
		img, _, err := image.Decode(file)
		if err != nil {
			fyne.LogError("Failed to read background image", err)
		}
		_ = file.Close()
		return img
	}

	img, _, err := image.Decode(bytes.NewReader(wmTheme.Background.StaticContent))
	if err != nil {
		fyne.LogError("Failed to read background resource", err)
	}
	return img
}

func (x *x11WM) updateBackgrounds() {
	geom, err := xproto.GetGeometry(x.x.Conn(), xproto.Drawable(x.x.RootWin())).Reply()
	if err != nil {
		fyne.LogError("Unable to look up root geometry", err)
		return
	}
	root := xgraphics.New(x.x, image.Rect(0, 0, int(geom.Width), int(geom.Height)))

	data := x.updatedBackgroundImage()
	for _, screen := range fynedesk.Instance().Screens().Screens() {
		scaled := resize.Resize(uint(screen.Width), uint(screen.Height), data, resize.Lanczos3)
		for y := screen.Y; y < screen.Y+screen.Height; y++ {
			for x := screen.X; x < screen.X+screen.Width; x++ {
				root.Set(x, y, scaled.At(x-screen.X, y-screen.Y))
			}
		}
	}

	root.XSurfaceSet(x.x.RootWin())
	root.XDraw()
	root.XPaint(x.x.RootWin())

	root.Destroy()
}
