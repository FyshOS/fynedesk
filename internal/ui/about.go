package ui

import (
	"net/url"
	"runtime/debug"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	wmTheme "fyne.io/fynedesk/theme"
)

func newURLButton(label, link string) *widget.Button {
	return widget.NewButton(label, func() {
		u, err := url.Parse(link)
		if err != nil {
			fyne.LogError("Could not parse URL", err)
		}

		fyne.CurrentApp().OpenURL(u)
	})
}

func (w *widgetPanel) showAbout() {
	if w.about != nil {
		w.about.CenterOnScreen()
		w.about.Show()

		for _, win := range w.desk.WindowManager().Windows() {
			if win.Properties().Title() == w.about.Title() {
				w.desk.WindowManager().RaiseToTop(win)
				break
			}
		}
		return
	}
	win := fyne.CurrentApp().NewWindow("About FyneDesk")

	title := widget.NewRichTextFromMarkdown("**Version:** " + version())
	title.Segments[0].(*widget.TextSegment).Style.Alignment = fyne.TextAlignCenter
	authors := widget.NewRichTextFromMarkdown("\n**Authors:**\n\n * Andy Williams\n * Stephen Houston\n * Jacob Alzén\n")
	buttons := container.NewGridWithColumns(3,
		newURLButton("Home Page", "https://fyne.io"),
		newURLButton("Report Issue", "https://github.com/fyne-io/desktop/issues/new"),
		newURLButton("Sponsor", "https://github.com/sponsors/fyne-io"),
	)

	bg := canvas.NewImageFromResource(wmTheme.FyneAboutBackground)
	bg.FillMode = canvas.ImageFillContain
	bg.Translucency = 0.67
	win.SetContent(container.NewMax(bg, container.NewBorder(title, buttons, nil, nil, authors)))
	win.SetCloseIntercept(func() {
		win.Hide()
	})

	w.about = win
	win.CenterOnScreen()
	win.Show()
}

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}

	return "(devel)"
}
