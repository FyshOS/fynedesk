package fyles

import (
	"image/color"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"

	lib "github.com/fyshos/fyles/pkg/fyles"
	"golang.org/x/sys/execabs"

	"fyshos.com/fynedesk"
	wmtheme "fyshos.com/fynedesk/theme"
)

var fylesMeta = fynedesk.ModuleMetadata{
	Name:        "Desktop Files",
	NewInstance: newFyles,
}

type fyles struct{}

func (f *fyles) Destroy() {
}

func (f *fyles) ScreenAreaWidget() fyne.CanvasObject {
	icons := lib.NewFylesPanel(f.tapped, fynedesk.Instance().Root())
	icons.HideParent = true
	home, _ := os.UserHomeDir()
	icons.SetDir(storage.NewFileURI(filepath.Join(home, "Desktop")))

	desk := fynedesk.Instance()
	var barPad *canvas.Rectangle
	if desk.Settings().NarrowLeftLauncher() {
		barPad = canvas.NewRectangle(color.Transparent)
		barPad.SetMinSize(fyne.NewSize(wmtheme.NarrowBarWidth, 1))
	}

	rightIndent := wmtheme.WidgetPanelWidth
	if desk.Settings().NarrowWidgetPanel() {
		rightIndent = wmtheme.NarrowBarWidth
	}
	widgetPad := canvas.NewRectangle(color.Transparent)
	widgetPad.SetMinSize(fyne.NewSize(rightIndent, 1))

	return container.NewBorder(nil, nil, barPad, widgetPad, container.NewPadded(icons))
}

func (f *fyles) Metadata() fynedesk.ModuleMetadata {
	return fylesMeta
}

func (f *fyles) tapped(u fyne.URI) {
	p, err := execabs.LookPath("fyles")
	if p != "" && err == nil {
		if ok, _ := storage.CanList(u); ok {
			err := execabs.Command(p, u.Path()).Start()
			if err != nil {
				log.Println("Error opening Fyles", err)
			}
			return
		}
	} else {
		log.Println(">>> dir", u)
		return
	}
	log.Println(">>> open", u)
}

// newFyles creates a new module that will manage desktop file icons.
func newFyles() fynedesk.Module {
	return &fyles{}
}
