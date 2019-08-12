package wm

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/fyne-io/desktop"

	"fyne.io/fyne"

	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func newBorder(title string, class []string, command string, iconName string) fyne.CanvasObject {
	var res fyne.Resource
	filler := canvas.NewRectangle(theme.BackgroundColor()) // make a border on the X axis only
	filler.SetMinSize(fyne.NewSize(0, 2))                  // 0 wide forced
	fdoDesktop := desktop.FdoLookupApplicationWinInfo(title, class, command, iconName)
	if fdoDesktop != nil {
		bytes, err := ioutil.ReadFile(fdoDesktop.IconPath)
		if err != nil {
			log.Print(err)
		} else {
			res = fyne.NewStaticResource(desktop.FdoResourceFormat(filepath.Base(fdoDesktop.IconPath)), bytes)
		}
	}
	titleBar := widget.NewHBox(filler,
		widget.NewButtonWithIcon("", theme.CancelIcon(), func() {}),
		widget.NewLabel(title),
		layout.NewSpacer())

	if res != nil {
		icon := fyne.NewContainerWithLayout(layout.NewCenterLayout(), widget.NewIcon(res))
		titleBar.Append(icon)
	}

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(titleBar, nil, nil, nil),
		titleBar)
}
