package desktop

import (
	"os/exec"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

func newBar() fyne.CanvasObject {
	return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		widget.NewButton("Browser", func() {
			exec.Command("chromium").Start()
		}),
		widget.NewButton("Terminal", func() {
			exec.Command("xterm").Start()
		}),
		layout.NewSpacer(),
	)
}
