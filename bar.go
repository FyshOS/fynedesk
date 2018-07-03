package desktop

import "os/exec"
import "time"

import "github.com/fyne-io/fyne"
import "github.com/fyne-io/fyne/layout"
import "github.com/fyne-io/fyne/widget"

func clockTick(clock *widget.Label) {
	tick := time.NewTicker(time.Second)
	go func() {
		for {
			<-tick.C
			clock.SetText(formattedTime())
		}
	}()
}

func formattedTime() string {
	return time.Now().Format("15:04:05")
}

func createClock() *widget.Label {
	clock := widget.NewLabel(formattedTime())
	clock.Alignment = fyne.TextAlignTrailing

	go clockTick(clock)

	return clock
}

func newBar(app fyne.App) fyne.CanvasObject {
	return fyne.NewContainerWithLayout(layout.NewGridLayout(3),
		widget.NewButton("Quit", func() {
			app.Quit()
		}),
		widget.NewButton("Terminal", func() {
			exec.Command("terminology").Start()
		}),
		createClock(),
	)
}
