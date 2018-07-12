package desktop

import "os/exec"
import "time"

import "github.com/fyne-io/fyne"
import "github.com/fyne-io/fyne/layout"
import "github.com/fyne-io/fyne/theme"
import "github.com/fyne-io/fyne/widget"

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

func createClock() *widget.List {
	clock := widget.NewLabel(formattedTime())
	clock.Alignment = fyne.TextAlignCenter
	clock.TextStyle.Monospace = true
	date := widget.NewLabel(formattedDate())
	date.Alignment = fyne.TextAlignCenter
	date.TextStyle.Monospace = true

	go clockTick(clock, date)

	return widget.NewList(clock, date)
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

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, quit, clock),
		quit,
		clock,
		buttons,
	)
}
