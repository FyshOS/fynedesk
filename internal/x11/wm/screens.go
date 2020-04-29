// +build linux

package wm // import "fyne.io/fynedesk/wm"

import (
	"fyne.io/fyne"

	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/x11"
)

type x11ScreensProvider struct {
	screens []*fynedesk.Screen
	active  *fynedesk.Screen
	primary *fynedesk.Screen
	single  bool
	x       *x11WM
	root    xproto.Window

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

func (xsp *x11ScreensProvider) ScreenForGeometry(g fynedesk.Geometry) *fynedesk.Screen {
	if len(xsp.screens) <= 1 {
		return xsp.screens[0]
	}

	middleX, middleY := g.Center()
	for i := 0; i < len(xsp.screens); i++ {
		if xsp.screens[i].Contains(middleX, middleY) {
			return xsp.screens[i]
		}
	}
	return xsp.active
}

func (xsp *x11ScreensProvider) ScreenForWindow(win fynedesk.Window) *fynedesk.Screen {
	if len(xsp.screens) <= 1 {
		return xsp.screens[0]
	}

	return xsp.ScreenForGeometry(win.Geometry())
}

func getScale(widthPx, widthMm uint16) float32 {
	dpi := float32(widthPx) / (float32(widthMm) / 25.4)
	if dpi > 1000 || dpi < 10 {
		dpi = 96
	}

	return float32(float64(dpi) / 96.0)
}

func (xsp *x11ScreensProvider) insertInOrder(tmpScreens []*fynedesk.Screen, outputInfo *randr.GetOutputInfoReply,
	crtcInfo *randr.GetCrtcInfoReply) ([]*fynedesk.Screen, int) {
	insertIndex := -1
	for i, screen := range tmpScreens {
		if screen.X >= int(crtcInfo.X) && screen.Y >= int(crtcInfo.Y) {
			insertIndex = i
			break
		}
	}

	crtcGeom := fynedesk.Geometry{X: int(crtcInfo.X), Y: int(crtcInfo.Y), Width: uint(crtcInfo.Width), Height: uint(crtcInfo.Height)}
	if insertIndex == -1 {
		tmpScreens = append(tmpScreens, &fynedesk.Screen{Name: string(outputInfo.Name),
			Geometry: crtcGeom, Scale: getScale(crtcInfo.Width, uint16(outputInfo.MmWidth))})
		insertIndex = len(tmpScreens) - 1
	} else {
		tmpScreens = append(tmpScreens, nil)
		copy(tmpScreens[insertIndex+1:], tmpScreens[insertIndex:])
		tmpScreens[insertIndex] = &fynedesk.Screen{Name: string(outputInfo.Name),
			Geometry: crtcGeom, Scale: getScale(crtcInfo.Width, uint16(outputInfo.MmWidth))}
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
	xsp.screens = []*fynedesk.Screen{{Name: "Screen0", Geometry: x11.GeometryFromRect(xwindow.RootGeometry(xsp.x.x)),
		Scale: 1.0}}
	xsp.primary = xsp.screens[0]
	xsp.active = xsp.screens[0]
}
