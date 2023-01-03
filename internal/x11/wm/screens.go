//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package wm

import (
	"fyne.io/fyne/v2"

	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"

	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/internal/x11"
)

const baselineDPI = 120.0

type x11ScreensProvider struct {
	screens []*fynedesk.Screen
	active  *fynedesk.Screen
	primary *fynedesk.Screen
	single  bool
	x       *x11WM

	onChange []func()
}

// NewX11ScreensProvider returns a screen provider for use in x11 desktop mode
func NewX11ScreensProvider(mgr fynedesk.WindowManager) fynedesk.ScreenList {
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

func (xsp *x11ScreensProvider) SetActive(s *fynedesk.Screen) {
	xsp.active = s
}

func (xsp *x11ScreensProvider) Active() *fynedesk.Screen {
	return xsp.active
}

func (xsp *x11ScreensProvider) AddChangeListener(f func()) {
	xsp.onChange = append(xsp.onChange, f)
}

func (xsp *x11ScreensProvider) Primary() *fynedesk.Screen {
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

func (xsp *x11ScreensProvider) Screens() []*fynedesk.Screen {
	return xsp.screens
}

func (xsp *x11ScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *fynedesk.Screen {
	if len(xsp.screens) <= 1 {
		return xsp.screens[0]
	}
	for i := 0; i < len(xsp.screens); i++ {
		xx, yy, ww, hh := xsp.screens[i].X, xsp.screens[i].Y,
			xsp.screens[i].Width, xsp.screens[i].Height
		middleW := width / 2
		middleH := height / 2
		middleW += x
		middleH += y
		if middleW >= xx && middleH >= yy &&
			middleW <= xx+ww && middleH <= yy+hh {
			return xsp.screens[i]
		}
	}
	return xsp.active
}

func (xsp *x11ScreensProvider) ScreenForWindow(win fynedesk.Window) *fynedesk.Screen {
	if len(xsp.screens) <= 1 {
		return xsp.screens[0]
	}

	x, y, w, h := win.(x11.XWin).Geometry()
	if w == 0 && h == 0 {
		return xsp.Primary()
	}
	return xsp.ScreenForGeometry(x, y, int(w), int(h))
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

func (xsp *x11ScreensProvider) insertInOrder(tmpScreens []*fynedesk.Screen, outputInfo *randr.GetOutputInfoReply, crtcInfo *randr.GetCrtcInfoReply) ([]*fynedesk.Screen, int) {
	insertIndex := -1
	for i, screen := range tmpScreens {
		if screen.X >= int(crtcInfo.X) && screen.Y >= int(crtcInfo.Y) {
			insertIndex = i
			break
		}

	}

	newScreen := &fynedesk.Screen{Name: string(outputInfo.Name),
		X: int(crtcInfo.X), Y: int(crtcInfo.Y), Width: int(crtcInfo.Width), Height: int(crtcInfo.Height),
		Scale: getScale(crtcInfo.Width, uint16(outputInfo.MmWidth))}
	if insertIndex == -1 {
		tmpScreens = append(tmpScreens, newScreen)
		insertIndex = len(tmpScreens) - 1
	} else {
		tmpScreens = append(tmpScreens, nil)
		copy(tmpScreens[insertIndex+1:], tmpScreens[insertIndex:])
		tmpScreens[insertIndex] = newScreen
	}
	return tmpScreens, insertIndex
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
	primaryFound := false
	var tmpScreens []*fynedesk.Screen
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
		insertIndex := 0
		tmpScreens, insertIndex = xsp.insertInOrder(tmpScreens, outputInfo, crtcInfo)
		if primaryInfo != nil {
			if string(primaryInfo.Name) == string(outputInfo.Name) {
				primaryFound = true
				xsp.primary = tmpScreens[insertIndex]
				xsp.active = tmpScreens[insertIndex]
			}
		}
	}
	if !primaryFound {
		xsp.primary = tmpScreens[0]
		xsp.active = tmpScreens[0]
	}
	xsp.screens = tmpScreens
}

func (xsp *x11ScreensProvider) setupSingleScreen() {
	xsp.single = true
	xsp.screens = []*fynedesk.Screen{{Name: "Screen0",
		X: xwindow.RootGeometry(xsp.x.x).X(), Y: xwindow.RootGeometry(xsp.x.x).Y(),
		Width: xwindow.RootGeometry(xsp.x.x).Width(), Height: xwindow.RootGeometry(xsp.x.x).Height(),
		Scale: 1.0}}
	xsp.primary = xsp.screens[0]
	xsp.active = xsp.screens[0]
}
