package ui

import (
	"image/color"
	"os"
	"path"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
	wmtheme "fyne.io/desktop/theme"
)

type widgetRenderer struct {
	panel *widgetPanel

	layout  fyne.Layout
	objects []fyne.CanvasObject
}

func (w *widgetRenderer) MinSize() fyne.Size {
	return w.layout.MinSize(w.objects)
}

func (w *widgetRenderer) Layout(size fyne.Size) {
	w.layout.Layout(w.objects, size)
}

func (w *widgetRenderer) Refresh() {
	w.panel.clock.Color = theme.TextColor()
	canvas.Refresh(w.panel.clock)
}

func (w *widgetRenderer) BackgroundColor() color.Color {
	r, _, _, _ := theme.BackgroundColor().RGBA()
	if uint8(r) > 0x99 {
		return wmtheme.WidgetPanelBackgroundLight
	}
	return wmtheme.WidgetPanelBackgroundDark
}

func (w *widgetRenderer) Objects() []fyne.CanvasObject {
	return w.objects
}

func (w *widgetRenderer) Destroy() {
}

type widgetPanel struct {
	widget.BaseWidget

	desk       desktop.Desktop
	root       fyne.Window
	appExecWin fyne.Window

	clock *canvas.Text
	date  *widget.Label
}

func (w *widgetPanel) clockTick() {
	tick := time.NewTicker(time.Second)
	go func() {
		for {
			<-tick.C
			w.clock.Text = formattedTime()
			canvas.Refresh(w.clock)

			w.date.SetText(formattedDate())
			canvas.Refresh(w.date)
		}
	}()
}

func formattedTime() string {
	return time.Now().Format("15:04pm")
}

func formattedDate() string {
	return time.Now().Format("2 January")
}

func (w *widgetPanel) createClock() {
	var style fyne.TextStyle
	style.Monospace = true

	w.clock = &canvas.Text{
		Color:     theme.TextColor(),
		Text:      formattedTime(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
		TextSize:  3 * theme.TextSize(),
	}
	w.date = &widget.Label{
		Text:      formattedDate(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
	}

	go w.clockTick()
}

func (w *widgetPanel) showAccountMenu(from fyne.CanvasObject) {
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("About", func() {
			showAbout()
		}),
		fyne.NewMenuItem("Settings", func() {
			showSettings(w.desk.Settings().(*deskSettings))
		}),
	}
	if w.desk.WindowManager() != nil {
		items = append(items, fyne.NewMenuItem("Blank Screen", w.desk.WindowManager().Blank))
	}
	if os.Getenv("FYNE_DESK_RUNNER") != "" && w.desk.(*deskLayout).wm != nil {
		items = append(items, fyne.NewMenuItem("Reload", func() {
			os.Exit(1)
		}))
	}

	closeLabel := "Log Out"
	if w.desk.(*deskLayout).wm == nil {
		closeLabel = "Quit"
	}
	items = append(items, fyne.NewMenuItem(closeLabel, func() {
		w.root.Close()
	}))

	popup := widget.NewPopUpMenu(fyne.NewMenu("Account", items...), w.root.Canvas())

	bottomLeft := fyne.CurrentApp().Driver().AbsolutePositionForObject(from)
	popup.Move(bottomLeft.Subtract(fyne.NewPos(0, popup.MinSize().Height)))
	popup.Resize(fyne.NewSize(from.Size().Width, popup.Content.MinSize().Height))
}

func (w *widgetPanel) CreateRenderer() fyne.WidgetRenderer {
	accountLabel := "Account"
	homedir, err := os.UserHomeDir()
	if err == nil {
		accountLabel = path.Base(homedir)
	} else {
		fyne.LogError("Unable to look up user", err)
	}
	var account *widget.Button
	account = widget.NewButtonWithIcon(accountLabel, wmtheme.UserIcon, func() {
		w.showAccountMenu(account)
	})
	appExecButton := widget.NewButtonWithIcon("Applications", theme.SearchIcon(), ShowAppLauncher)

	mods := w.desk.Modules()
	objects := []fyne.CanvasObject{
		w.clock,
		w.date}

	objects = append(objects, layout.NewSpacer())
	for _, m := range mods {
		if statusMod, ok := m.(desktop.StatusAreaModule); ok {
			wid := statusMod.StatusAreaWidget()
			if wid == nil {
				continue
			}

			objects = append(objects, wid)
		}
	}

	objects = append(objects, appExecButton, account)

	return &widgetRenderer{
		panel:   w,
		layout:  layout.NewVBoxLayout(),
		objects: objects,
	}
}

func newWidgetPanel(rootDesk desktop.Desktop) *widgetPanel {
	w := &widgetPanel{
		desk:       rootDesk,
		root:       rootDesk.(*deskLayout).primaryWin,
		appExecWin: nil,
	}
	w.ExtendBaseWidget(w)
	w.createClock()

	return w
}
