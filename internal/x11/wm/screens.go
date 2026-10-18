//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package wm

import (
	"fyne.io/fyne/v2"

	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"

	"fyshos.com/tyde"
	"fyshos.com/tyde/internal/x11"
)

const baselineDPI = 120.0

type x11ScreensProvider struct {
	screens []*tyde.Screen
	active  *tyde.Screen
	primary *tyde.Screen
	single  bool
	x       *x11WM

	onChange []func()
}

// NewX11ScreensProvider returns a screen provider for use in x11 desktop mode
func NewX11ScreensProvider(mgr tyde.WindowManager) tyde.ScreenList {
	screensProvider := &x11ScreensProvider{}
	screensProvider.x = mgr.(*x11WM)
	err := randr.Init(screensProvider.x.x.Conn())
	if err != nil {
		fyne.LogError("Could not initialize randr", err)
		return screensProvider
	}
	randr.SelectInput(screensProvider.x.x.Conn(), screensProvider.x.x.RootWin(), randr.NotifyMaskScreenChange)
	screensProvider.setupScreens()

	return screensProvider
}

func (xsp *x11ScreensProvider) SetActive(s *tyde.Screen) {
	xsp.active = s
}

func (xsp *x11ScreensProvider) Active() *tyde.Screen {
	return xsp.active
}

func (xsp *x11ScreensProvider) AddChangeListener(f func()) {
	xsp.onChange = append(xsp.onChange, f)
}

func (xsp *x11ScreensProvider) Primary() *tyde.Screen {
	return xsp.primary
}

func (xsp *x11ScreensProvider) RefreshScreens() {
	if xsp.single {
		xsp.setupSingleScreen()
	} else {
		xsp.setupScreens()
	}

	for _, listener := range xsp.onChange {
		listener()
	}
}

func (xsp *x11ScreensProvider) Screens() []*tyde.Screen {
	return xsp.screens
}

// ScreenForGeometry returns the screen whose area contains the centre of the
// rectangle. Virtual desktops stack vertically in root coordinates, so the centre
// is first wrapped into the height of the root to resolve the screen a window on
// another desktop would show on.
func (xsp *x11ScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *tyde.Screen {
	if len(xsp.screens) <= 1 {
		return xsp.screens[0]
	}
	middleW := x + width/2
	middleH := wrapIntoRoot(y+height/2, xsp.screens)
	for i := 0; i < len(xsp.screens); i++ {
		xx, yy, ww, hh := xsp.screens[i].X, xsp.screens[i].Y,
			xsp.screens[i].Width, xsp.screens[i].Height
		if middleW >= xx && middleH >= yy &&
			middleW <= xx+ww && middleH <= yy+hh {
			return xsp.screens[i]
		}
	}
	return xsp.active
}

func (xsp *x11ScreensProvider) ScreenForWindow(win tyde.Window) *tyde.Screen {
	if len(xsp.screens) <= 1 {
		return xsp.screens[0]
	}

	x, y, w, h := win.(x11.XWin).Geometry()
	if w == 0 && h == 0 {
		return xsp.Primary()
	}
	return xsp.ScreenForGeometry(x, y, int(w), int(h))
}

// wrapIntoRoot folds a root-relative y onto the visible root, whose height is the
// bottom edge of the lowest screen; one virtual desktop is one root height.
func wrapIntoRoot(y int, screens []*tyde.Screen) int {
	rootHeight := 0
	for _, s := range screens {
		if bottom := s.Y + s.Height; bottom > rootHeight {
			rootHeight = bottom
		}
	}
	if rootHeight == 0 {
		return y
	}
	y %= rootHeight
	if y < 0 {
		y += rootHeight
	}
	return y
}

func getScale(widthPx, widthMm uint16) float32 {
	dpi := float32(widthPx) / (float32(widthMm) / 25.4)
	if dpi > 1000 || dpi < 10 {
		dpi = baselineDPI
	}

	scale := float32(float64(dpi) / baselineDPI)
	if scale < 1.0 {
		return 1.0
	}
	return scale
}

// screenOutput is the information a connected output contributes to the screen
// layout (an abstraction over the randr types.
type screenOutput struct {
	name                string
	x, y, width, height int
	scale               float32
	primary             bool
}

func insertInOrder(screens []*tyde.Screen, newScreen *tyde.Screen) []*tyde.Screen {
	insertIndex := -1
	for i, screen := range screens {
		if screen.X >= newScreen.X && screen.Y >= newScreen.Y {
			insertIndex = i
			break
		}
	}

	if insertIndex == -1 {
		return append(screens, newScreen)
	}

	screens = append(screens, nil)
	copy(screens[insertIndex+1:], screens[insertIndex:])
	screens[insertIndex] = newScreen
	return screens
}

// screenForOutput returns the screen an output joins rather than adds to, which
// is one starting at the same place - two outputs showing the same corner of the
// desktop are mirroring each other.
func screenForOutput(screens []*tyde.Screen, out screenOutput) *tyde.Screen {
	for _, screen := range screens {
		if screen.X == out.x && screen.Y == out.y {
			return screen
		}
	}

	return nil
}

// screensFromOutputs works out the logical screens of the desktop from the
// connected outputs, and which of them is primary.
//
// Mirrored outputs make up a single screen. The screen also takes the identity
// of the primary output of its group.
func screensFromOutputs(outputs []screenOutput) ([]*tyde.Screen, *tyde.Screen) {
	var screens []*tyde.Screen
	var primary *tyde.Screen
	for _, out := range outputs {
		if mirrored := screenForOutput(screens, out); mirrored != nil {
			// The desktop covers the largest of the mirrored modes, as that is
			// what the X screen was sized to hold.
			mirrored.Width = max(mirrored.Width, out.width)
			mirrored.Height = max(mirrored.Height, out.height)
			if out.primary {
				mirrored.Name = out.name
				mirrored.Scale = out.scale
				primary = mirrored
			}
			continue
		}

		screen := &tyde.Screen{
			Name: out.name,
			X:    out.x, Y: out.y, Width: out.width, Height: out.height,
			Scale: out.scale,
		}
		screens = insertInOrder(screens, screen)
		if out.primary {
			primary = screen
		}
	}

	if primary == nil && len(screens) > 0 {
		primary = screens[0]
	}
	return screens, primary
}

func (xsp *x11ScreensProvider) setupScreens() {
	root := xproto.Setup(xsp.x.x.Conn()).DefaultScreen(xsp.x.x.Conn()).Root
	resources, err := randr.GetScreenResources(xsp.x.x.Conn(), root).Reply()
	if err != nil || len(resources.Outputs) == 0 {
		fyne.LogError("Could not get randr screen resources", err)
		xsp.setupSingleScreen()
		return
	}

	var primaryInfo *randr.GetOutputInfoReply
	primary, err := randr.GetOutputPrimary(xsp.x.x.Conn(), root).Reply()
	if err == nil {
		primaryInfo, _ = randr.GetOutputInfo(xsp.x.x.Conn(), primary.Output, 0).Reply()
	}

	var outputs []screenOutput
	for _, output := range resources.Outputs {
		outputInfo, err := randr.GetOutputInfo(xsp.x.x.Conn(), output, 0).Reply()
		if err != nil {
			fyne.LogError("Could not get randr output", err)
			continue
		}
		if outputInfo.Crtc == 0 || outputInfo.Connection == randr.ConnectionDisconnected {
			continue
		}
		crtcInfo, err := randr.GetCrtcInfo(xsp.x.x.Conn(), outputInfo.Crtc, 0).Reply()
		if err != nil {
			fyne.LogError("Could not get randr crtcs", err)
			continue
		}

		outputs = append(outputs, screenOutput{
			name: string(outputInfo.Name),
			x:    int(crtcInfo.X), y: int(crtcInfo.Y),
			width: int(crtcInfo.Width), height: int(crtcInfo.Height),
			scale:   getScale(crtcInfo.Width, uint16(outputInfo.MmWidth)),
			primary: primaryInfo != nil && string(primaryInfo.Name) == string(outputInfo.Name),
		})
	}

	screens, primaryScreen := screensFromOutputs(outputs)
	if len(screens) == 0 { // nothing is switched on, fall back to the X screen size
		xsp.setupSingleScreen()
		return
	}

	xsp.screens = screens
	xsp.primary = primaryScreen
	xsp.active = primaryScreen
}

func (xsp *x11ScreensProvider) setupSingleScreen() {
	xsp.single = true
	xsp.screens = []*tyde.Screen{{
		Name: "Screen0",
		X:    xwindow.RootGeometry(xsp.x.x).X(), Y: xwindow.RootGeometry(xsp.x.x).Y(),
		Width: xwindow.RootGeometry(xsp.x.x).Width(), Height: xwindow.RootGeometry(xsp.x.x).Height(),
		Scale: 1.0,
	}}
	xsp.primary = xsp.screens[0]
	xsp.active = xsp.screens[0]
}
