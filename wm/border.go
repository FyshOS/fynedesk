package wm

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/fynedesk"
	wmTheme "fyne.io/fynedesk/theme"
)

func makeFiller(width float32) fyne.CanvasObject {
	filler := canvas.NewRectangle(color.Transparent) // make a border on the X axis only
	filler.SetMinSize(fyne.NewSize(width, 2))        // width forced

	return filler
}

// NewBorder creates a new window border for the given window details
func NewBorder(win fynedesk.Window, icon fyne.Resource, canMaximize bool) *Border {
	desk := fynedesk.Instance()

	if icon == nil {
		iconTheme := desk.Settings().IconTheme()
		app := desk.IconProvider().FindAppFromWinInfo(win)
		if app != nil {
			icon = app.Icon(iconTheme, int(wmTheme.TitleHeight*2))
		}
	}

	max := &widget.Button{Icon: wmTheme.MaximizeIcon, Importance: widget.LowImportance, OnTapped: func() {
		if win.Maximized() {
			win.Unmaximize()
		} else {
			win.Maximize()
		}
	}}

	if win.Maximized() {
		max.Icon = theme.ViewRestoreIcon()
	}
	if !canMaximize {
		max.Disable()
	}

	min := &widget.Button{Icon: wmTheme.IconifyIcon, Importance: widget.LowImportance, OnTapped: func() {
		win.Iconify()
	}}

	title := widget.NewLabel(win.Properties().Title())
	titleBar := newColoredHBox(win.Focused(), win, makeFiller(0),
		newCloseButton(win),
		max,
		min,
		title,
		layout.NewSpacer())
	titleBar.title = title
	titleBar.max = max

	if icon != nil {
		appIcon := canvas.NewImageFromResource(icon)
		appIcon.SetMinSize(fyne.NewSize(wmTheme.TitleHeight, wmTheme.TitleHeight))
		titleBar.Append(appIcon)
		titleBar.Append(makeFiller(1))
	}

	return titleBar
}

// Border represents a window border. It draws the title bar and provides functions to manipulate it.
type Border struct {
	*widget.Box
	focused bool
	title   *widget.Label
	max     *widget.Button
	win     fynedesk.Window
}

type coloredBoxRenderer struct {
	fyne.WidgetRenderer
	b *Border
}

// DoubleTapped is called when the user double taps a frame, it toggles the maximised state.
func (c *Border) DoubleTapped(*fyne.PointEvent) {
	if c.win.Maximized() {
		c.win.Unmaximize()
		return
	}
	c.win.Maximize()
}

// CreateRenderer creates a new renderer for this border
//
// Implements: fyne.Widget
func (c *Border) CreateRenderer() fyne.WidgetRenderer {
	render := &coloredBoxRenderer{b: c}
	render.WidgetRenderer = c.Box.CreateRenderer()
	return render
}

// SetFocused specifies whether this window is focused, and updates visuals accordingly.
func (c *Border) SetFocused(focus bool) {
	c.focused = focus
	c.Refresh()
}

// SetMaximized updates the state of the border maximize indicators and refreshes
func (c *Border) SetMaximized(isMax bool) {
	if isMax {
		c.max.Icon = theme.ViewRestoreIcon()
	} else {
		c.max.Icon = wmTheme.MaximizeIcon
	}
	c.max.Refresh()
}

// SetTitle updates the title portion of this border and refreshes.
func (c *Border) SetTitle(title string) {
	c.title.SetText(title)
}

func (r *coloredBoxRenderer) BackgroundColor() color.Color {
	if r.b.focused {
		return theme.BackgroundColor()
	}
	return theme.DisabledButtonColor()
}

func newColoredHBox(focused bool, win fynedesk.Window, objs ...fyne.CanvasObject) *Border {
	ret := &Border{focused: focused, win: win}
	ret.Box = widget.NewHBox(objs...)
	ret.ExtendBaseWidget(ret)

	return ret
}
