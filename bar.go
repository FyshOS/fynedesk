package desktop

import (
	"os/exec"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func clockTick(clock, date *widget.Label) {
	tick := time.NewTicker(time.Second)
	go func() {
		for {
			<-tick.C
			clock.SetText(formattedTime())
			date.SetText(formattedDate())
		}
	}()
}

func formattedTime() string {
	return time.Now().Format("15:04:05")
}

func formattedDate() string {
	return time.Now().Format("2 Jan")
}

func createClock() *widget.Box {
	clock := widget.NewLabel(formattedTime())
	clock.Alignment = fyne.TextAlignCenter
	clock.TextStyle.Monospace = true
	date := widget.NewLabel(formattedDate())
	date.Alignment = fyne.TextAlignCenter
	date.TextStyle.Monospace = true

	go clockTick(clock, date)

	return widget.NewHBox(clock, date)
}

func newBar(app fyne.App) fyne.CanvasObject {
	quit := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		app.Quit()
	})
	clock := createClock()
	buttons := fyne.NewContainerWithLayout(layout.NewGridLayout(5),
		widget.NewButton("Browser", func() {
			exec.Command("chromium").Start()
		}),
		widget.NewButton("Terminal", func() {
			exec.Command("terminology").Start()
		}),
	)

	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, quit, clock),
		quit,
		clock,
		buttons,
	)
	return fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		canvas.NewRectangle(theme.BackgroundColor()),
		content,
	)
}
