package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
)

var switcherInstance *switcher

const (
	switcherIconSize = 64
	switcherTextSize = 24
)

type switchIcon struct {
	widget.BaseWidget
	current bool

	parent *switcher
	win    desktop.Window
}

func (s *switchIcon) CreateRenderer() fyne.WidgetRenderer {
	var res fyne.Resource
	title := s.win.Title()
	app := desktop.Instance().IconProvider().FindAppFromWinInfo(s.win)
	if app != nil {
		res = app.Icon(desktop.Instance().Settings().IconTheme(), switcherIconSize*2)
		title = app.Name()
	} else {
		res = s.win.Icon()
	}

	img := canvas.NewImageFromResource(res)
	text := widget.NewLabel(title)
	return &switchIconRenderer{icon: s,
		img: img, text: text, objects: []fyne.CanvasObject{img, text}}
}

func (s *switchIcon) FocusGained() {
	s.current = true
	s.Refresh()
}

func (s *switchIcon) FocusLost() {
	s.current = false
	s.Refresh()
}

func (s *switchIcon) Focused() bool {
	return s.current
}

func (s *switchIcon) TypedRune(rune) {
}

func (s *switchIcon) TypedKey(ev *fyne.KeyEvent) {
	switch ev.Name {
	case fyne.KeyReturn, fyne.KeyEnter:
		s.parent.win.Close()

		s.parent.raise(s)
	case fyne.KeyRight:
		s.parent.next()
	case fyne.KeyLeft:
		s.parent.previous()
	case fyne.KeyEscape:
		s.parent.win.Close()
	}
}
func newSwitchIcon(p *switcher, win desktop.Window) *switchIcon {
	ret := &switchIcon{
		parent: p,
		win:    win,
	}
	ret.ExtendBaseWidget(ret)
	return ret
}

type switchIconRenderer struct {
	icon *switchIcon

	img     *canvas.Image
	text    *widget.Label
	objects []fyne.CanvasObject
}

func (s switchIconRenderer) Layout(fyne.Size) {
	s.img.Resize(fyne.NewSize(switcherIconSize, switcherIconSize))
	s.text.Resize(fyne.NewSize(switcherIconSize, switcherTextSize))
	s.text.Move(fyne.NewPos(0, switcherIconSize+theme.Padding()))
}

func (s switchIconRenderer) MinSize() fyne.Size {
	return fyne.NewSize(switcherIconSize, switcherIconSize+switcherTextSize+theme.Padding())
}

func (s switchIconRenderer) Refresh() {
	canvas.Refresh(s.icon)
}

func (s switchIconRenderer) BackgroundColor() color.Color {
	if s.icon.current {
		return theme.PrimaryColor()
	}
	return theme.BackgroundColor()
}

func (s switchIconRenderer) Objects() []fyne.CanvasObject {
	return s.objects
}

func (s switchIconRenderer) Destroy() {
}

type switcher struct {
	win   fyne.Window
	icons []fyne.CanvasObject
}

func (s *switcher) currentIndex() int {
	for i, item := range s.icons {
		if item.(*switchIcon).current {
			return i
		}
	}

	return 0
}

func (s *switcher) next() {
	if len(s.icons) == 0 {
		return
	}

	i := s.currentIndex()
	i++
	if i >= len(s.icons) {
		i = 0
	}
	s.win.Canvas().Focus(s.icons[i].(*switchIcon))
}

func (s *switcher) previous() {
	if len(s.icons) == 0 {
		return
	}

	i := s.currentIndex()
	i--
	if i < 0 {
		i = len(s.icons) - 1
	}
	s.win.Canvas().Focus(s.icons[i].(*switchIcon))
}

func (s *switcher) raise(icon *switchIcon) {
	if icon.win.Iconic() {
		icon.win.Uniconify()
	}
	icon.win.RaiseToTop()
}

func (s *switcher) loadUI() fyne.Window {
	win := fyne.CurrentApp().NewWindow("Application instance")

	win.SetContent(widget.NewHBox(s.icons...))
	win.CenterOnScreen()

	return win
}

func (s *switcher) loadIcons(list []desktop.Window) []fyne.CanvasObject {
	var ret []fyne.CanvasObject

	for _, item := range list {
		ret = append(ret, newSwitchIcon(s, item))
	}

	return ret
}

func showAppSwitcherAt(off int, wm desktop.WindowManager) {
	if wm == nil || len(wm.Windows()) <= 1 {
		return
	}
	s := &switcher{}
	s.icons = s.loadIcons(wm.Windows())
	s.win = s.loadUI()
	if off < 0 {
		off = len(s.icons) + off // plus a negative is minus
	}
	s.win.Canvas().Focus(s.icons[off].(*switchIcon))
	s.win.SetOnClosed(func() {
		switcherInstance = nil
	})
	s.win.Show()
	switcherInstance = s
}

// ShowAppSwitcher shows the application switcher to change windows.
// The most recently used not-top window will be selected by default.
// If the switcher was already visible then it will select the next window in order.
func ShowAppSwitcher() {
	wm := desktop.Instance().WindowManager()
	if switcherInstance != nil {
		switcherInstance.next()
		return
	}

	showAppSwitcherAt(1, wm)
}

// ShowAppSwitcherReverse shows the application switcher to change windows.
// The least recently used window will be selected by default.
// If the switcher was already visible then it will select the last window in order.
func ShowAppSwitcherReverse() {
	wm := desktop.Instance().WindowManager()
	if switcherInstance != nil {
		switcherInstance.previous()
		return
	}

	showAppSwitcherAt(-1, wm)
}

// HideAppSwitcher dismisses the application switcher and raises
// whichever window was selected.
func HideAppSwitcher() {
	if switcherInstance == nil {
		return
	}

	switcherInstance.win.Close()
	switcherInstance.raise(switcherInstance.win.Canvas().Focused().(*switchIcon))
}
