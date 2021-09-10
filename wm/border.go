package wm

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

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

	title := canvas.NewText(win.Properties().Title(), theme.ForegroundColor())
	buttonPos := fynedesk.Instance().Settings().BorderButtonPosition()

	var titleBar *Border
	if buttonPos == "Right" {
		titleBar = newColoredHBox(win.Focused(), win, makeFiller(1),
			title,
			layout.NewSpacer(),
			min,
			max,
			newCloseButton(win),
			makeFiller(0),
		)
	} else {
		titleBar = newColoredHBox(win.Focused(), win, makeFiller(0),
			newCloseButton(win),
			max,
			min,
			title,
			layout.NewSpacer(),
		)
	}

	titleBar.title = title
	titleBar.max = max

	if icon != nil {
		appIcon := &widget.Button{Icon: icon, Importance: widget.LowImportance}
		appIcon.OnTapped = func() {
			titleBar.showMenu(appIcon)
		}

		if buttonPos == "Right" {
			titleBar.prepend(appIcon)
			titleBar.prepend(makeFiller(1))
		} else {
			titleBar.append(appIcon)
			titleBar.append(makeFiller(1))
		}
		titleBar.appIcon = appIcon
	}

	return titleBar
}

// Border represents a window border. It draws the title bar and provides functions to manipulate it.
type Border struct {
	widget.BaseWidget
	content *fyne.Container
	focused bool
	appIcon *widget.Button
	title   *canvas.Text
	max     *widget.Button
	win     fynedesk.Window
}

// DoubleTapped is called when the user double taps a frame, it toggles the maximised state.
func (c *Border) DoubleTapped(*fyne.PointEvent) {
	if c.win.Maximized() {
		c.win.Unmaximize()
		return
	}
	c.win.Maximize()
}

func (c *Border) prepend(obj fyne.CanvasObject) {
	c.content.Objects = append([]fyne.CanvasObject{obj}, c.content.Objects...)
	c.Refresh()
}

func (c *Border) append(obj fyne.CanvasObject) {
	c.content.Add(obj)
	c.Refresh()
}

func (c *Border) showMenu(from fyne.CanvasObject) {
	name := c.title.Text
	if len(name) > 25 {
		name = name[:25] + "..."
	}
	title := fyne.NewMenuItem(name, func() {})
	title.Disabled = true
	max := fyne.NewMenuItem("Maximize", func() {
		if c.win.Maximized() {
			c.win.Unmaximize()
		} else {
			c.win.Maximize()
		}
	})
	if c.win.Maximized() {
		max.Checked = true
	}
	menu := fyne.NewMenu("",
		title,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Minimize", func() {
			c.win.Iconify()
		}),
		max,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Close", func() {
			c.win.Close()
		}))

	pos := c.win.Position()
	fynedesk.Instance().ShowMenuAt(menu, pos.Add(from.Position()))
}

// CreateRenderer creates a new renderer for this border
//
// Implements: fyne.Widget
func (c *Border) CreateRenderer() fyne.WidgetRenderer {
	render := &coloredBoxRenderer{b: c, bg: canvas.NewRectangle(theme.BackgroundColor())}
	return render
}

// SetFocused specifies whether this window is focused, and updates visuals accordingly.
func (c *Border) SetFocused(focus bool) {
	c.focused = focus
	c.Refresh()
}

// SetIcon tells the border to change the icon that should be used
func (c *Border) SetIcon(icon fyne.Resource) {
	if icon == nil {
		return
	}

	c.appIcon.SetIcon(icon)
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
	c.title.Text = title
	c.title.Refresh()
}

func newColoredHBox(focused bool, win fynedesk.Window, objs ...fyne.CanvasObject) *Border {
	ret := &Border{focused: focused, win: win}
	ret.content = container.NewHBox(objs...)
	ret.ExtendBaseWidget(ret)

	return ret
}

type coloredBoxRenderer struct {
	b  *Border
	bg *canvas.Rectangle
}

func (r *coloredBoxRenderer) Destroy() {
}

func (r *coloredBoxRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.b.content.Resize(size)
}

func (r *coloredBoxRenderer) MinSize() fyne.Size {
	return r.b.content.MinSize()
}

func (r *coloredBoxRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.b.content}
}

func (r *coloredBoxRenderer) Refresh() {
	r.b.title.Color = theme.ForegroundColor()
	r.b.title.Refresh()

	if r.b.focused {
		r.bg.FillColor = theme.BackgroundColor()
	} else {
		r.bg.FillColor = theme.DisabledButtonColor()
	}
	r.bg.Refresh()
}
