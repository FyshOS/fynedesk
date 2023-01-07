package ui

import (
	"os"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
	wmtheme "fyshos.com/fynedesk/theme"
)

func (w *widgetPanel) appendAppCategories(acc *widget.Accordion, win fyne.Window) {
	accList := acc.Items
	cats := w.desk.IconProvider().CategorizedApps()
	var catNames []string
	hasOther := false
	for cat := range cats {
		if cat == "Other" {
			hasOther = true
			continue
		}
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)
	if hasOther {
		catNames = append(catNames, "Other")
	}

	for _, cat := range catNames {
		list := cats[cat]
		var items []fyne.CanvasObject
		for _, app := range list {
			if app.Hidden() {
				continue
			}
			btn := w.newAppButton(app, win)
			items = append(items, btn)
			defer w.loadIcon(app, btn)
		}
		accList = append(accList, widget.NewAccordionItem(cat,
			container.NewVBox(items...)))
	}
	acc.Items = accList
	acc.Refresh()
}

func (w *widgetPanel) showAccountMenu(_ fyne.CanvasObject) {
	w2 := fyne.CurrentApp().Driver().(deskDriver.Driver).CreateSplashWindow()
	w2.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeyEscape {
			w2.Close()
		}
	})
	items1 := []fyne.CanvasObject{container.NewMax(canvas.NewRectangle(theme.ErrorColor()),
		&widget.Button{Icon: theme.LogoutIcon(), Importance: widget.LowImportance, OnTapped: func() {
			w2.Close()
			w.desk.WindowManager().Close()
		}})}
	isEmbed := w.desk.(*desktop).root.Title() != RootWindowName
	if !isEmbed {
		items1 = append(items1, &widget.Button{Icon: wmtheme.LockIcon, Importance: widget.LowImportance, OnTapped: func() {
			w2.Close()
			w.desk.(*desktop).LockScreen()
		}})
		if os.Getenv("FYNE_DESK_RUNNER") != "" {
			items1 = append(items1, &widget.Button{Icon: theme.ViewRefreshIcon(), Importance: widget.LowImportance, OnTapped: func() {
				os.Exit(5)
			}})
		}
	}

	items2 := []fyne.CanvasObject{
		&widget.Button{Icon: theme.QuestionIcon(), Importance: widget.LowImportance, OnTapped: func() {
			w.showAbout()
			w2.Close()
		}},
		&widget.Button{Icon: theme.SettingsIcon(), Importance: widget.LowImportance, OnTapped: func() {
			w.showSettings()
			w2.Close()
		}}}
	items := container.NewBorder(nil, nil, container.NewHBox(items1...), container.NewHBox(items2...),
		&widget.Button{Icon: theme.SearchIcon(), Text: "Search", Importance: widget.LowImportance, OnTapped: func() {
			ShowAppLauncher()
			w2.Close()
		}})

	var recent []fyne.CanvasObject
	for _, app := range w.desk.RecentApps() {
		btn := w.newAppButton(app, w2)
		recent = append(recent, btn)
		defer w.loadIcon(app, btn)
	}

	acc := widget.NewAccordion(widget.NewAccordionItem("Recent",
		container.NewVBox(recent...)))
	acc.MultiOpen = true
	acc.Open(0)
	go w.appendAppCategories(acc, w2)

	w2.SetContent(container.NewBorder(
		items, nil, nil, nil,
		container.NewScroll(acc)))
	winSize := w.desk.(*desktop).root.Canvas().Size()
	pos := fyne.NewPos(winSize.Width-300, winSize.Height-360)
	w.desk.WindowManager().ShowOverlay(w2, fyne.NewSize(300, 360), pos)
}

func (w *widgetPanel) newAppButton(app fynedesk.AppData, w2 fyne.Window) *widget.Button {
	b := widget.NewButtonWithIcon(app.Name(), wmtheme.BrokenImageIcon, func() {
		w2.Close()
		_ = w.desk.RunApp(app)
	})
	b.Alignment = widget.ButtonAlignLeading
	return b
}

func (w *widgetPanel) loadIcon(app fynedesk.AppData, btn *widget.Button) {
	iconRes := app.Icon(w.desk.Settings().IconTheme(), int(64*w.desk.Screens().Primary().CanvasScale()))

	btn.SetIcon(iconRes)
}
