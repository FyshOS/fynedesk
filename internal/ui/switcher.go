package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
)

const (
	switcherIconSize = 64
	switcherTextSize = 24
)

type switchIcon struct {
	widget.BaseWidget
	current bool

	parent *Switcher
	win    fynedesk.Window
}

func (s *switchIcon) CreateRenderer() fyne.WidgetRenderer {
	var res fyne.Resource
	title := s.win.Properties().Title()
	app := s.parent.provider.FindAppFromWinInfo(s.win)
	if app != nil {
		res = app.Icon(fynedesk.Instance().Settings().IconTheme(), switcherIconSize*2)
		title = app.Name()
	} else {
		res = s.win.Properties().Icon()
	}

	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = theme.InputRadiusSize()
	img := canvas.NewImageFromResource(res)
	text := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{})
	text.Truncation = fyne.TextTruncateEllipsis
	return &switchIconRenderer{icon: s, bg: bg,
		img: img, text: text, objects: []fyne.CanvasObject{bg, img, text}}
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

func newSwitchIcon(p *Switcher, win fynedesk.Window) *switchIcon {
	ret := &switchIcon{
		parent: p,
		win:    win,
	}
	ret.ExtendBaseWidget(ret)
	return ret
}

type switchIconRenderer struct {
	icon *switchIcon

	bg      *canvas.Rectangle
	img     *canvas.Image
	text    *widget.Label
	objects []fyne.CanvasObject
}

func (s switchIconRenderer) Layout(size fyne.Size) {
	s.bg.Move(fyne.NewPos(-theme.Padding()/2, -theme.Padding()/2))
	s.bg.Resize(size.Add(fyne.NewSize(theme.Padding(), theme.Padding())))
	s.img.Resize(fyne.NewSize(switcherIconSize, switcherIconSize))
	s.text.Resize(fyne.NewSize(switcherIconSize+theme.Padding()*2, switcherTextSize))
	s.text.Move(fyne.NewPos(-theme.Padding(), switcherIconSize-theme.Padding()/2))
}

func (s switchIconRenderer) MinSize() fyne.Size {
	return fyne.NewSize(switcherIconSize, switcherIconSize+switcherTextSize+theme.Padding())
}

func (s switchIconRenderer) Refresh() {
	if s.icon.current {
		s.bg.FillColor = theme.PrimaryColor()
	} else {
		s.bg.FillColor = color.Transparent
	}
	canvas.Refresh(s.icon)
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
	provider fynedesk.ApplicationProvider
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

func (s *Switcher) loadUI(title string) {
	win := s.win
	if win == nil {
		if d, ok := fyne.CurrentApp().Driver().(deskDriver.Driver); ok {
			win = d.CreateSplashWindow()
			win.SetPadded(true)
		} else {
			win = fyne.CurrentApp().NewWindow(title)
		}
		s.win = win
	}

	win.SetContent(container.NewHBox(s.icons...))
	win.CenterOnScreen()
	win.SetTitle(title)
}

func (s *Switcher) loadIcons(list []fynedesk.Window) []fyne.CanvasObject {
	var ret []fyne.CanvasObject

	for _, item := range list {
		if item.Desktop() != fynedesk.Instance().Desktop() {
			continue
		}
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
	go func() {
		time.Sleep(time.Millisecond * 100)
		s.win.Hide()
	}()
}

// Show the app switcher, it would then be hidden with HideApply or HideCancel.
func (s *Switcher) Show() {
	s.win.Show()
}

func newAppSwitcherAt(off int, wins []fynedesk.Window, prov fynedesk.ApplicationProvider) *Switcher {
	s := &Switcher{provider: prov}
	s.icons = s.loadIcons(wins)
	if len(s.icons) <= 1 { // don't actually show if only 1 is visible
		return nil
	}

	s.loadUI("Window switcher " + SkipTaskbarHint)
	if off < 0 {
		off = len(s.icons) + off // plus a negative is minus
	}
	s.win.Canvas().Focus(s.icons[off].(*switchIcon))
	return s
}

// NewAppSwitcher creates the application Switcher to change windows.
// The most recently used not-top window will be selected by default.
// If the Switcher was already visible then it will select the next window in order.
func NewAppSwitcher(wins []fynedesk.Window, prov fynedesk.ApplicationProvider) *Switcher {
	return newAppSwitcherAt(1, wins, prov)
}

// NewAppSwitcherReverse creates the application Switcher to change windows.
// The least recently used window will be selected by default.
// If the Switcher was already visible then it will select the last window in order.
func NewAppSwitcherReverse(wins []fynedesk.Window, prov fynedesk.ApplicationProvider) *Switcher {
	return newAppSwitcherAt(-1, wins, prov)
}
