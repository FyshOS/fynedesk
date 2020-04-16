package ui

import (
	"net/url"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

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

func showAbout() {
	w := fyne.CurrentApp().NewWindow("About FyneDesk")

	title := widget.NewLabelWithStyle("Fyne Desk 0.1.3", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	authors := widget.NewLabel("\nAuthors:\n\n    Andy Williams\n    Stephen Houston\n")
	buttons := fyne.NewContainerWithLayout(layout.NewGridLayout(3),
		newURLButton("Home Page", "https://fyne.io"),
		newURLButton("Report Issue", "https://github.com/fyne-io/desktop/issues/new"),
		newURLButton("Sponsor", "https://github.com/sponsors/fyne-io"),
	)

	bg := canvas.NewImageFromResource(wmTheme.FyneAboutBackground)
	bg.FillMode = canvas.ImageFillContain
	bg.Translucency = 0.67
	w.SetContent(fyne.NewContainerWithLayout(layout.NewMaxLayout(), bg,
		fyne.NewContainerWithLayout(layout.NewBorderLayout(title, buttons, nil, nil),
			title, authors, buttons)))

	w.CenterOnScreen()
	w.Show()
}
