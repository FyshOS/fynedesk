package launcher

import (
	"time"

	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/internal/ui"
	wmTheme "fyshos.com/fynedesk/theme"
	"github.com/fyne-io/terminal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

const (
	delay     = time.Second / 25
	termTitle = "Quake Terminal " + ui.SkipTaskbarHint
	height    = 240
	step      = 40
)

var termMeta = fynedesk.ModuleMetadata{
	Name:        "\"Quake\" (hover) terminal",
	NewInstance: newTerm,
}

type term struct {
	shown bool
	win   fynedesk.Window
	ui    fyne.Window
}

func (t *term) Destroy() {
}

func (t *term) Metadata() fynedesk.ModuleMetadata {
	return termMeta
}

func (t *term) Shortcuts() map[*fynedesk.Shortcut]func() {
	return map[*fynedesk.Shortcut]func(){
		&fynedesk.Shortcut{Name: "Open Quake Terminal", KeyName: fyne.KeyBackTick, Modifier: fynedesk.UserModifier}: func() {
			t.toggle()
		}}
}

func (t *term) createTerm() {
	win := fyne.CurrentApp().Driver().(desktop.Driver).CreateSplashWindow()
	win.SetTitle(termTitle)

	bg := canvas.NewRectangle(theme.BackgroundColor())
	img := canvas.NewImageFromResource(theme.NewDisabledResource(theme.ComputerIcon()))
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(200, 200))
	over := canvas.NewRectangle(wmTheme.WidgetPanelBackground())
	matchTheme(bg, over)

	console := terminal.New()
	win.SetContent(container.NewStack(bg, img, over, console))
	win.Canvas().Focus(console)
	t.ui = win

	go func() {
		err := console.RunLocalShell()
		if err != nil {
			fyne.LogError("Failed to open terminal", err)
		}
		t.hide() // terminal exited

		t.createTerm() // reset for next usage
	}()
}

func (t *term) getHandle() fynedesk.Window {
	// TODO a better way to capture window frame without showing it and waiting...
	//t.ui.Resize(fyne.NewSize(0, 0))
	t.ui.Show()

	i := 0
	for {
		time.Sleep(time.Second / 50)

		for _, w := range fynedesk.Instance().WindowManager().Windows() {
			if w.Properties().Title() == termTitle {
				return w
			}
		}

		i++
		if i > 50 {
			return nil // something went wrong
		}
	}
}

func (t *term) hide() {
	screen := fynedesk.Instance().Screens().Primary()
	left := float32(screen.X) / screen.Scale
	y := float32(screen.Y) / screen.Scale
	end := float32(screen.Y)/screen.Scale - height
	for y > end {
		t.win.Move(fyne.NewPos(left, y))
		time.Sleep(delay)
		y -= step
	}
	t.win.Move(fyne.NewPos(left, end))

	t.ui.Hide()
	t.shown = false
}

func (t *term) show() {
	screen := fynedesk.Instance().Screens().Primary()
	t.win.Resize(fyne.NewSize(float32(screen.Width)/screen.Scale, height))
	//	t.ui.Show()
	t.win.RaiseToTop()

	left := float32(screen.X) / screen.Scale
	y := float32(screen.Y)/screen.Scale - height
	end := float32(screen.Y) / screen.Scale
	for y < end {
		t.win.Resize(fyne.NewSize(float32(screen.Width)/screen.Scale, height)) // force it ASAP
		t.win.Move(fyne.NewPos(left, y))
		time.Sleep(delay)
		y += step
	}
	t.win.Move(fyne.NewPos(left, end))
	t.shown = true
}

func (t *term) toggle() {
	if !t.shown {
		t.win = t.getHandle()

		t.show()
	} else {
		t.hide()
		t.win = nil
	}
}

func matchTheme(bg, over *canvas.Rectangle) {
	ch := make(chan fyne.Settings)
	go func() {
		for {
			<-ch

			bg.FillColor = theme.BackgroundColor()
			bg.Refresh()
			over.FillColor = wmTheme.WidgetPanelBackground()
			over.Refresh()
		}
	}()
	fyne.CurrentApp().Settings().AddChangeListener(ch)
}

func newTerm() fynedesk.Module {
	t := &term{}
	t.createTerm()
	return t
}
