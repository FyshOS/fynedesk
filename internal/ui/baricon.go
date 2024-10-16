package ui

import (
	"image/color"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
)

// appWindow describes a type of icon that refers to an open window rather than an app.
// The findApp function can be used to attempt looking up the application from it's window.
type appWindow struct {
	win fynedesk.Window
	bar *bar
}

// findApp will try to return an application data associated with a window.
// This may fail for many reasons, usually related too bad window metadata, and will then return nil.
func (a *appWindow) findApp() fynedesk.AppData {
	if a.win == nil {
		return nil
	}

	return a.bar.desk.IconProvider().FindAppFromWinInfo(a.win)
}

type barIconRenderer struct {
	objects []fyne.CanvasObject

	image *barIcon
}

func (bi *barIconRenderer) MinSize() fyne.Size {
	size := theme.IconInlineSize()
	return fyne.NewSize(size, size)
}

func (bi *barIconRenderer) Layout(size fyne.Size) {
	if len(bi.objects) == 0 {
		return
	}

	bi.objects[0].Resize(size)
}

func (bi *barIconRenderer) Objects() []fyne.CanvasObject {
	return bi.objects
}

func (bi *barIconRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (bi *barIconRenderer) Refresh() {
	bi.objects = nil

	if bi.image.resource != nil {
		raster := canvas.NewImageFromResource(bi.image.resource)
		raster.FillMode = canvas.ImageFillContain

		bi.objects = []fyne.CanvasObject{raster}
	}
	bi.Layout(bi.image.Size())

	if bi.image.windowData != nil && bi.image.windowData.win.Iconic() {
		if img, ok := bi.objects[0].(*canvas.Image); ok {
			img.Translucency = 0.67
		}
	}

	canvas.Refresh(bi.image)
}

func (bi *barIconRenderer) Destroy() {
}

// barIcon widget is a basic image component that load's its resource to match the theme.
type barIcon struct {
	widget.BaseWidget

	onTapped   func()           // The function that will be called when the icon is clicked
	resource   fyne.Resource    // The image data of the image that the icon uses
	appData    fynedesk.AppData // The application data corresponding to this icon.(if it is a launcher)
	windowData *appWindow       // The window data associated with this icon (if it is a task window)
}

// Tapped means barIcon has been clicked
func (bi *barIcon) Tapped(*fyne.PointEvent) {
	bi.onTapped()
}

func addToBar(icon fynedesk.AppData) {
	settings := fynedesk.Instance().Settings()
	icons := settings.LauncherIcons()
	icons = append(icons, icon.Name())

	settings.(*deskSettings).setLauncherIcons(icons)
}

func removeFromBar(icon fynedesk.AppData) {
	settings := fynedesk.Instance().Settings()
	icons := settings.LauncherIcons()

	index := -1
	for i, defaultApp := range icons {
		if defaultApp == icon.Name() {
			index = i
			break
		}
	}
	if index >= 0 {
		icons = append(icons[:index], icons[index+1:]...)
	}
	settings.(*deskSettings).setLauncherIcons(icons)
}

// TappedSecondary means barIcon has been clicked by a secondary binding
func (bi *barIcon) TappedSecondary(ev *fyne.PointEvent) {
	app := bi.appData
	if app == nil && bi.windowData != nil {
		app = bi.windowData.findApp()
	}
	if app == nil || app.Name() == "" {
		return
	}

	var menu *widget.PopUpMenu
	addRemove := fyne.NewMenuItem("Remove "+app.Name(), func() {
		if bi.windowData != nil {
			addToBar(app)
		} else {
			removeFromBar(app)
		}
		menu.Hide()
	})

	if bi.windowData != nil {
		addRemove.Label = "Pin " + app.Name()
	}

	items := []*fyne.MenuItem{addRemove}
	editor := editorPath()
	if app.Source() != nil && editor != "" {
		srcDir := filepath.Join(sourceRoot(), app.Name())
		items = append(items, fyne.NewMenuItem("Edit", func() {
			cmd := exec.Command(editor, srcDir)
			err := cmd.Start()

			if err != nil {
				fyne.LogError("Failed to start app editor: "+editor, err)
			}
		}))
	}

	fynedesk.Instance().ShowMenuAt(fyne.NewMenu("", items...), ev.AbsolutePosition)
}

// CreateRenderer is a private method to fyne which links this widget to its renderer
func (bi *barIcon) CreateRenderer() fyne.WidgetRenderer {
	render := &barIconRenderer{image: bi}
	render.Refresh()

	return render
}

func newBarIcon(res fyne.Resource, appData fynedesk.AppData, winData *appWindow) *barIcon {
	barIcon := &barIcon{resource: res, appData: appData, windowData: winData}
	barIcon.ExtendBaseWidget(barIcon)

	return barIcon
}

func editorPath() string {
	fysion, err := exec.LookPath("fysion")
	if err == nil && fysion != "" {
		return fysion
	}

	return ""
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func sourceRoot() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}

	return filepath.Join(u.HomeDir, "FysionApps")
}
