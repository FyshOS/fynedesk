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

func makeFiller(width int) fyne.CanvasObject {
	filler := canvas.NewRectangle(theme.BackgroundColor()) // make a border on the X axis only
	filler.SetMinSize(fyne.NewSize(width, 2))              // width forced

	return filler
}

// NewBorder creates a new window border for the given window details
func NewBorder(win fynedesk.Window, icon fyne.Resource, canMaximize bool) fyne.CanvasObject {
	desk := fynedesk.Instance()

	if icon == nil {
		iconTheme := desk.Settings().IconTheme()
		app := desk.IconProvider().FindAppFromWinInfo(win)
		if app != nil {
			icon = app.Icon(iconTheme, wmTheme.TitleHeight*2)
		}
	}

	max := widget.NewButtonWithIcon("", wmTheme.MaximizeIcon, func() {
		if win.Maximized() {
			win.Unmaximize()
		} else {
			win.Maximize()
		}
	})
	if win.Maximized() {
		max.Icon = theme.ViewRestoreIcon()
	}
	if !canMaximize {
		max.Disable()
	}
	titleBar := newColoredHBox(win.Focused(), win, makeFiller(0),
		newCloseButton(win),
		max,
		widget.NewButtonWithIcon("", wmTheme.IconifyIcon, func() {
			win.Iconify()
		}),
		widget.NewLabel(win.Properties().Title()),
		layout.NewSpacer())

	if icon != nil {
		appIcon := canvas.NewImageFromResource(icon)
		appIcon.SetMinSize(fyne.NewSize(wmTheme.TitleHeight, wmTheme.TitleHeight))
		titleBar.Append(appIcon)
		titleBar.Append(makeFiller(1))
	}

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(titleBar, nil, nil, nil),
		titleBar)
}

type coloredHBox struct {
	*widget.Box
	focused bool
	win fynedesk.Window
}

type coloredBoxRenderer struct {
	fyne.WidgetRenderer
	focused bool
}

func (c *coloredHBox) DoubleTapped(*fyne.PointEvent) {
	if c.win.Maximized() {
		c.win.Unmaximize()
		return
	}
	c.win.Maximize()
}

func (c *coloredHBox) CreateRenderer() fyne.WidgetRenderer {
	render := &coloredBoxRenderer{focused: c.focused}
	render.WidgetRenderer = c.Box.CreateRenderer()
	return render
}

func (r *coloredBoxRenderer) BackgroundColor() color.Color {
	if r.focused {
		return theme.BackgroundColor()
	}
	return theme.ButtonColor()
}

func newColoredHBox(focused bool, win fynedesk.Window, objs ...fyne.CanvasObject) *coloredHBox {
	ret := &coloredHBox{focused: focused, win: win}
	ret.Box = widget.NewHBox(objs...)
	ret.ExtendBaseWidget(ret)

	return ret
}
