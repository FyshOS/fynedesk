package wm

import (
	"math"
	"os"
	"strconv"

	"fyne.io/desktop"
	"fyne.io/fyne"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func getScale(widthPx uint16, widthMm uint32) float32 {
	env := os.Getenv("FYNE_SCALE")

	if env != "" && env != "auto" {
		scale, err := strconv.ParseFloat(env, 32)
		if err == nil && scale != 0 {
			return float32(scale)
		}
		fyne.LogError("Error reading scale", err)
	}

	if env != "auto" {
		setting := fyne.CurrentApp().Settings().Scale()
		switch setting {
		case fyne.SettingsScaleAuto:
			// fall through
		case 0.0:
			if env == "" {
				return 1.0
			}
			// fall through
		default:
			return setting
		}
	}

	dpi := float32(widthPx) / (float32(widthMm) / 25.4)
	if dpi > 1000 || dpi < 10 {
		dpi = 96
	}
	return float32(math.Round(float64(dpi)/144.0*10.0)) / 10.0
}

func (x *x11WM) setupScreens() {
	err := randr.Init(x.x.Conn())
	if err != nil {
		fyne.LogError("Could not initialize randr", err)
	} else {
		root := xproto.Setup(x.x.Conn()).DefaultScreen(x.x.Conn()).Root
		resources, err := randr.GetScreenResources(x.x.Conn(), root).Reply()
		if err != nil {
			fyne.LogError("Could not get randr screen resources", err)
		} else {
			primary, err := randr.GetOutputPrimary(x.x.Conn(),
				xproto.Setup(x.x.Conn()).DefaultScreen(x.x.Conn()).Root).Reply()
			if err != nil {
				fyne.LogError("Could not determine randr primary output", err)
			}
			primaryInfo, err := randr.GetOutputInfo(x.x.Conn(), primary.Output, 0).Reply()
			if err != nil {
				fyne.LogError("Could not determine randr primary output information", err)
			}
			primaryFound := false
			var scale float32 = 1.0
			var firstFoundMmWidth uint32 = 0
			var firstFoundWidth uint16 = 0
			i := 0
			for _, output := range resources.Outputs {
				outputInfo, err := randr.GetOutputInfo(x.x.Conn(), output, 0).Reply()
				if err != nil {
					fyne.LogError("Could not get randr output", err)
					continue
				}
				if outputInfo.Crtc == 0 || outputInfo.Connection == randr.ConnectionDisconnected {
					continue
				}
				crtcInfo, err := randr.GetCrtcInfo(x.x.Conn(), outputInfo.Crtc, 0).Reply()
				if err != nil {
					fyne.LogError("Could not get randr crtcs", err)
					continue
				}
				if i == 0 {
					firstFoundMmWidth = outputInfo.MmWidth
					firstFoundWidth = crtcInfo.Width
				}
				x.screens = append(x.screens, desktop.Screen{Name: string(outputInfo.Name), Index: i,
					X: int(crtcInfo.X), Y: int(crtcInfo.Y), Width: int(crtcInfo.Width), Height: int(crtcInfo.Height)})
				if primaryInfo != nil {
					if string(primaryInfo.Name) == string(outputInfo.Name) {
						primaryFound = true
						x.primary = i
						x.active = i
						scale = getScale(crtcInfo.Width, outputInfo.MmWidth)
					}
				}
				i++
			}
			if !primaryFound {
				scale = getScale(firstFoundWidth, firstFoundMmWidth)
			}
			for j, screen := range x.screens {
				screen.ScaledX = int(float32(screen.X) / scale)
				screen.ScaledY = int(float32(screen.Y) / scale)
				screen.ScaledWidth = int(float32(screen.Width) / scale)
				screen.ScaledHeight = int(float32(screen.Height) / scale)
				x.screens[j] = screen
			}
		}
	}
	if len(x.screens) == 0 {
		x.screens = append(x.screens, desktop.Screen{Name: "Screen0", Index: 0,
			X: xwindow.RootGeometry(x.x).X(), Y: xwindow.RootGeometry(x.x).Y(),
			Width: xwindow.RootGeometry(x.x).Width(), Height: xwindow.RootGeometry(x.x).Height(),
			ScaledX: xwindow.RootGeometry(x.x).X(), ScaledY: xwindow.RootGeometry(x.x).Y(),
			ScaledWidth: xwindow.RootGeometry(x.x).Width(), ScaledHeight: xwindow.RootGeometry(x.x).Height()})
		x.primary = 0
		x.active = 0
	}
}
