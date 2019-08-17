package wm

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"github.com/fyne-io/desktop"
)

var (
	iconTheme = "hicolor"
	iconSize  = 32
)

func newBorder(win desktop.Window) fyne.CanvasObject {
	var res fyne.Resource

	filler := canvas.NewRectangle(theme.BackgroundColor()) // make a border on the X axis only
	filler.SetMinSize(fyne.NewSize(0, 2))                  // 0 wide forced

	icon := desktop.Environment().IconProvider().FindIconFromWinInfo(iconTheme, iconSize, win)
	if icon != nil {
		if icon.IconPath() != "" {
			bytes, err := ioutil.ReadFile(icon.IconPath())
			if err != nil {
				fyne.LogError("Cannot read file", err)
			} else {
				str := strings.Replace(icon.IconPath(), "-", "", -1)
				iconResource := strings.Replace(str, "_", "", -1)
				res = fyne.NewStaticResource(strings.ToLower(filepath.Base(iconResource)), bytes)
			}
		}
	}
	titleBar := widget.NewHBox(filler,
		widget.NewButtonWithIcon("", theme.CancelIcon(), func() {}),
		widget.NewLabel(win.Title()),
		layout.NewSpacer())

	if res != nil {
		icon := fyne.NewContainerWithLayout(layout.NewCenterLayout(), widget.NewIcon(res))
		titleBar.Append(icon)
	}

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(titleBar, nil, nil, nil),
		titleBar)
}
