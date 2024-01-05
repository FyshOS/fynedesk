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
	"fyne.io/fyne/v2/theme"

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
	icons.Filter = filterHidden()
	f.setDesktopDir(icons)

	desk := fynedesk.Instance()
	var barPad fyne.CanvasObject
	if desk.Settings().NarrowLeftLauncher() {
		r := canvas.NewRectangle(color.Transparent)
		r.SetMinSize(fyne.NewSize(wmtheme.NarrowBarWidth, 1))
		barPad = r
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

func (f *fyles) setDesktopDir(p *lib.Panel) {
	home, _ := os.UserHomeDir()
	u := storage.NewFileURI(filepath.Join(home, "Desktop"))
	settings := newCustomURI("settings://", "Settings", theme.SettingsIcon())
	trash := newCustomURI("file://"+filepath.Join(home, ".local", "share", "Trash", "files"), "Trash", theme.DeleteIcon())

	list, err := storage.List(u)
	list = append([]fyne.URI{settings, trash}, list...)
	if err != nil {
		fyne.LogError("Could not read Desktop dir", err)
	} else {
		p.SetListing(list)
	}
}

func (f *fyles) tapped(u fyne.URI) {
	if u.Scheme() == "settings" {
		fynedesk.Instance().ShowSettings()
		return
	}
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

type filter struct{}

func (f *filter) Matches(u fyne.URI) bool {
	return u.Name()[0] != '.'
}

func filterHidden() storage.FileFilter {
	return &filter{}
}

type trashURI struct {
	fyne.URI

	name string
	icon fyne.Resource
}

func newCustomURI(str, name string, icon fyne.Resource) fyne.URI {
	u, _ := storage.ParseURI(str)
	return &trashURI{URI: u, name: name, icon: icon}
}

func (t *trashURI) Name() string {
	return t.name
}

func (t *trashURI) Icon() fyne.Resource {
	return t.icon
}
