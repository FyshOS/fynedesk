package ui

import (
	"net/url"

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

func showAbout() {
	w := fyne.CurrentApp().NewWindow("About FyneDesk")

	title := widget.NewLabelWithStyle("Fyne Desk (develop)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	authors := widget.NewLabel("\nAuthors:\n\n    Andy Williams\n    Stephen Houston\n    Jacob Alzén\n")
	buttons := container.NewGridWithColumns(3,
		newURLButton("Home Page", "https://fyne.io"),
		newURLButton("Report Issue", "https://github.com/fyne-io/desktop/issues/new"),
		newURLButton("Sponsor", "https://github.com/sponsors/fyne-io"),
	)

	bg := canvas.NewImageFromResource(wmTheme.FyneAboutBackground)
	bg.FillMode = canvas.ImageFillContain
	bg.Translucency = 0.67
	w.SetContent(container.NewMax(bg, container.NewBorder(title, buttons, nil, nil, authors)))

	w.CenterOnScreen()
	w.Show()
}
