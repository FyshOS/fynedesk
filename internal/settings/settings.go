package settings

import (
	"fyne.io/desktop/theme"
	"fyne.io/fyne"
	"fyne.io/fyne/cmd/fyne_settings/settings"
	"fyne.io/fyne/widget"
)

func loadBarScreen() fyne.CanvasObject {
	return widget.NewLabel("TODO")
}

// Show loads and shows the fynedesk settings window
func Show() {
	w := fyne.CurrentApp().NewWindow("Fyne Settings")
	fyneSettings := settings.NewSettings()

	tabs := widget.NewTabContainer(
		&widget.TabItem{Text: "Appearance", Icon: fyneSettings.AppearanceIcon(),
			Content: fyneSettings.LoadAppearanceScreen()},
		&widget.TabItem{Text: "App Bar", Icon: theme.IconifyIcon, Content: loadBarScreen()})
	tabs.SetTabLocation(widget.TabLocationLeading)
	w.SetContent(tabs)

	w.Resize(fyne.NewSize(480, 320))
	w.Show()
}
