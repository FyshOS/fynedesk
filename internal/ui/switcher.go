package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/desktop"
)

const (
	switcherIconSize = 64
	switcherTextSize = 24
)

type switchIcon struct {
	widget.BaseWidget
	current bool

	parent *Switcher
	win    desktop.Window
}

func (s *switchIcon) CreateRenderer() fyne.WidgetRenderer {
	var res fyne.Resource
	title := s.win.Title()
	app := s.parent.provider.FindAppFromWinInfo(s.win)
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

// FocusGained is called when this icon gets focus - it becomes the candidate for window raising
func (s *switchIcon) FocusGained() {
	s.current = true
	s.Refresh()
}

// FocusLost is called when a different item is focused
func (s *switchIcon) FocusLost() {
	s.current = false
	s.Refresh()
}

// Focused returns whether or not this icon has focus
func (s *switchIcon) Focused() bool {
	return s.current
}

// TypedRune notifies when a rune is typed, we don't care
func (s *switchIcon) TypedRune(rune) {
}

// TypedKey is called when a key is typed. Currently this is handled by the window manager
func (s *switchIcon) TypedKey(*fyne.KeyEvent) {
}

func newSwitchIcon(p *Switcher, win desktop.Window) *switchIcon {
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

// Switcher is an instance of a visible app switcher that can request a change in window stacking order
type Switcher struct {
	win      fyne.Window
	icons    []fyne.CanvasObject
	provider desktop.ApplicationProvider
}

func (s *Switcher) currentIndex() int {
	for i, item := range s.icons {
		if item.(*switchIcon).current {
			return i
		}
	}

	return 0
}

func (s *Switcher) setCurrent(i int) {
	s.win.Canvas().Focus(s.icons[i].(*switchIcon))
}

// Next selects the next logical lower window in the stack.
// If the bottom most window is selected then it will wrap to the topmost.
// This will be raised if the change is applied by calling HideApply.
func (s *Switcher) Next() {
	if len(s.icons) == 0 {
		return
	}

	i := s.currentIndex()
	i++
	if i >= len(s.icons) {
		i = 0
	}
	s.setCurrent(i)
}

// Previous selects the next logical higher window in the stack.
// If the top most window was selected it wraps to the lowest.
// This will be raised if the change is applied by calling HideApply.
func (s *Switcher) Previous() {
	if len(s.icons) == 0 {
		return
	}

	i := s.currentIndex()
	i--
	if i < 0 {
		i = len(s.icons) - 1
	}
	s.setCurrent(i)
}

func (s *Switcher) raise(icon *switchIcon) {
	if icon.win.Iconic() {
		icon.win.Uniconify()
	}
	icon.win.RaiseToTop()
}

func (s *Switcher) loadUI() fyne.Window {
	win := fyne.CurrentApp().NewWindow("Application instance")

	win.SetContent(widget.NewHBox(s.icons...))
	win.CenterOnScreen()

	return win
}

func (s *Switcher) loadIcons(list []desktop.Window) []fyne.CanvasObject {
	var ret []fyne.CanvasObject

	for _, item := range list {
		ret = append(ret, newSwitchIcon(s, item))
	}

	return ret
}

// HideApply dismisses the application Switcher and raises
// whichever window was selected.
func (s *Switcher) HideApply() {
	s.HideCancel()
	s.raise(s.win.Canvas().Focused().(*switchIcon))
}

// HideCancel dismisses the application Switcher without changing window order.
func (s *Switcher) HideCancel() {
	s.win.Close()
}

func showAppSwitcherAt(off int, wins []desktop.Window, prov desktop.ApplicationProvider) *Switcher {
	if len(wins) <= 1 { // don't actually show
		return nil
	}

	s := &Switcher{provider: prov}
	s.icons = s.loadIcons(wins)
	s.win = s.loadUI()
	if off < 0 {
		off = len(s.icons) + off // plus a negative is minus
	}
	s.win.Canvas().Focus(s.icons[off].(*switchIcon))
	s.win.Show()
	return s
}

// ShowAppSwitcher shows the application Switcher to change windows.
// The most recently used not-top window will be selected by default.
// If the Switcher was already visible then it will select the next window in order.
func ShowAppSwitcher(wins []desktop.Window, prov desktop.ApplicationProvider) *Switcher {
	return showAppSwitcherAt(1, wins, prov)
}

// ShowAppSwitcherReverse shows the application Switcher to change windows.
// The least recently used window will be selected by default.
// If the Switcher was already visible then it will select the last window in order.
func ShowAppSwitcherReverse(wins []desktop.Window, prov desktop.ApplicationProvider) *Switcher {
	return showAppSwitcherAt(-1, wins, prov)
}
