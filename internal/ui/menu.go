package ui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

func (w *widgetPanel) showAccountMenu(_ fyne.CanvasObject) {
	w2 := fyne.CurrentApp().Driver().(deskDriver.Driver).CreateSplashWindow()
	w2.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeyEscape {
			w2.Close()
		}
	})
	items1 := []fyne.CanvasObject{container.NewMax(canvas.NewRectangle(theme.ErrorColor()),
		&widget.Button{Icon: theme.LogoutIcon(), Importance: widget.LowImportance, OnTapped: func() {
			w.desk.WindowManager().Close()
		}})}
	isEmbed := w.desk.(*desktop).root.Title() != RootWindowName
	if !isEmbed {
		items1 = append(items1, &widget.Button{Icon: wmtheme.LockIcon, Importance: widget.LowImportance, OnTapped: func() {
			w2.Close()
			w.desk.WindowManager().Blank()
		}})
		if os.Getenv("FYNE_DESK_RUNNER") != "" {
			items1 = append(items1, &widget.Button{Icon: theme.ViewRefreshIcon(), Importance: widget.LowImportance, OnTapped: func() {
				os.Exit(5)
			}})
		}
	}

	items2 := []fyne.CanvasObject{
		&widget.Button{Icon: theme.QuestionIcon(), Importance: widget.LowImportance, OnTapped: func() {
			showAbout()
			w2.Close()
		}},
		&widget.Button{Icon: theme.SettingsIcon(), Importance: widget.LowImportance, OnTapped: func() {
			showSettings(w.desk.Settings().(*deskSettings))
			w2.Close()
		}}}
	items := container.NewBorder(nil, nil, container.NewHBox(items1...), container.NewHBox(items2...),
		&widget.Button{Icon: theme.SearchIcon(), Text: "Search", Importance: widget.LowImportance, OnTapped: func() {
			ShowAppLauncher()
			w2.Close()
		}})

	var recent []fyne.CanvasObject
	for _, app := range w.desk.RecentApps() {
		recent = append(recent, w.newAppButton(app, w2))
	}

	acc := widget.NewAccordion(widget.NewAccordionItem("Recent",
		container.NewVBox(recent...)))

	for cat, list := range w.desk.IconProvider().CategorizedApps() {
		var items []fyne.CanvasObject
		for _, app := range list {
			if app.Hidden() {
				continue
			}
			items = append(items, w.newAppButton(app, w2))
		}
		acc.Append(widget.NewAccordionItem(cat,
			container.NewVBox(items...)))
	}
	acc.MultiOpen = true
	acc.Open(0)
	w2.SetContent(container.NewBorder(
		items, nil, nil, nil,
		container.NewScroll(acc)))

	winSize := w.desk.(*desktop).root.Canvas().Size()
	pos := fyne.NewPos(winSize.Width-300, winSize.Height-360)
	w.desk.WindowManager().ShowOverlay(w2, fyne.NewSize(300, 360), pos)
}

func (w *widgetPanel) newAppButton(app fynedesk.AppData, w2 fyne.Window) fyne.CanvasObject {
	iconRes := app.Icon(w.desk.Settings().IconTheme(), int(64*w.desk.Screens().Primary().CanvasScale()))

	return widget.NewButtonWithIcon(app.Name(), iconRes, func() {
		w2.Close()
		_ = w.desk.RunApp(app)
	})
}
