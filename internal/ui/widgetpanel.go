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

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
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

	desk       fynedesk.Desktop
	appExecWin fyne.Window

	clock         *canvas.Text
	date          *widget.Label
	modules       *fyne.Container
	notifications fyne.CanvasObject
}

func (w *widgetPanel) clockTick() {
	tick := time.NewTicker(time.Second)
	go func() {
		for {
			<-tick.C
			w.clock.Text = w.formattedTime()
			canvas.Refresh(w.clock)

			w.date.SetText(formattedDate())
			canvas.Refresh(w.date)
		}
	}()
}

func (w *widgetPanel) formattedTime() string {
	if w.desk.Settings().ClockFormatting() == "12h" {
		return time.Now().Format("03:04pm")
	}

	return time.Now().Format("15:04")
}

func formattedDate() string {
	return time.Now().Format("2 January")
}

func (w *widgetPanel) createClock() {
	var style fyne.TextStyle
	style.Monospace = true

	w.clock = &canvas.Text{
		Color:     theme.TextColor(),
		Text:      w.formattedTime(),
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
	isEmbed := w.desk.(*desktop).root.Title() != RootWindowName
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("About", func() {
			showAbout()
		}),
		fyne.NewMenuItem("Settings", func() {
			showSettings(w.desk.Settings().(*deskSettings))
		}),
	}
	if !isEmbed {
		items = append(items, fyne.NewMenuItem("Blank Screen", w.desk.WindowManager().Blank))
		if os.Getenv("FYNE_DESK_RUNNER") != "" {
			items = append(items, fyne.NewMenuItem("Reload", func() {
				os.Exit(5)
			}))
		}
	}

	closeLabel := "Log Out"
	if isEmbed {
		closeLabel = "Quit"
	}

	root := w.desk.(*desktop).root
	items = append(items, fyne.NewMenuItem(closeLabel, func() {
		w.desk.WindowManager().Close()
	}))

	popup := widget.NewPopUpMenuAtPosition(fyne.NewMenu("Account", items...), root.Canvas(), fyne.NewPos(0, 0)) //lint:ignore SA1019 Sort this out when deprecations gets sorted out upstream

	bottomLeft := fyne.CurrentApp().Driver().AbsolutePositionForObject(from)
	popup.Move(bottomLeft.Subtract(fyne.NewPos(0, popup.MinSize().Height)))
	popup.Resize(fyne.NewSize(from.Size().Width, popup.MinSize().Height))
	popup.Show()
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

	objects := []fyne.CanvasObject{
		w.clock,
		w.date,
		w.notifications}

	w.modules = fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	objects = append(objects, layout.NewSpacer(), w.modules, appExecButton, account)
	w.loadModules(w.desk.Modules())

	return &widgetRenderer{
		panel:   w,
		layout:  layout.NewVBoxLayout(),
		objects: objects,
	}
}

func (w *widgetPanel) reloadModules(mods []fynedesk.Module) {
	w.modules.Objects = nil
	w.loadModules(mods)
	w.modules.Refresh()
}

func (w *widgetPanel) loadModules(mods []fynedesk.Module) {
	for _, m := range mods {
		if statusMod, ok := m.(fynedesk.StatusAreaModule); ok {
			wid := statusMod.StatusAreaWidget()
			if wid == nil {
				continue
			}

			w.modules.Objects = append(w.modules.Objects, wid)
		}
	}
}

func newWidgetPanel(rootDesk fynedesk.Desktop) *widgetPanel {
	w := &widgetPanel{
		desk:       rootDesk,
		appExecWin: nil,
	}
	w.ExtendBaseWidget(w)
	w.notifications = startNotifications()
	w.createClock()

	return w
}
