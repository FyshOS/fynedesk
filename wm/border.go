package wm

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyshos.com/fynedesk"
	wmTheme "fyshos.com/fynedesk/theme"
)

// makeFiller creates a 0-width object to force padding on the edge of a container
func makeFiller() fyne.CanvasObject {
	filler := canvas.NewRectangle(color.Transparent)
	filler.SetMinSize(fyne.NewSize(0, 2))

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
		titleBar = newColoredHBox(win.Focused(), win,
			title,
			layout.NewSpacer(),
			min,
			max,
			newCloseButton(win),
			makeFiller(),
		)
	} else {
		titleBar = newColoredHBox(win.Focused(), win, makeFiller(),
			newCloseButton(win),
			max,
			min,
			title,
			layout.NewSpacer(),
		)
	}

	titleBar.title = title
	titleBar.max = max

	appIcon := &widget.Button{Icon: icon, Importance: widget.LowImportance}
	appIcon.OnTapped = func() {
		titleBar.showMenu(appIcon)
	}

	if buttonPos == "Right" {
		titleBar.prepend(appIcon)
		titleBar.prepend(makeFiller())
	} else {
		titleBar.append(appIcon)
		titleBar.append(makeFiller())
	}
	titleBar.appIcon = appIcon
	return titleBar
}

// Border represents a window border. It draws the title bar and provides functions to manipulate it.
type Border struct {
	widget.BaseWidget
	content *fyne.Container
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

	pos := c.win.Position()
	menuPos := pos.Add(from.Position())
	menu := fyne.NewMenu("",
		title,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Minimize", func() {
			c.win.Iconify()
		}),
		max,
		fyne.NewMenuItemSeparator(),
		c.makeDesktopMenu(menuPos),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Close", func() {
			c.win.Close()
		}))

	fynedesk.Instance().ShowMenuAt(menu, menuPos)
}

func (c *Border) makeDesktopMenu(pos fyne.Position) *fyne.MenuItem {
	desks := make([]*fyne.MenuItem, 4)
	for i := 0; i < 4; i++ {
		deskID := i
		desks[i] = fyne.NewMenuItem(fmt.Sprintf("Desktop %d", i+1), func() {
			if c.win.Pinned() {
				c.win.Unpin()
			}

			c.win.SetDesktop(deskID)
		})
	}
	pin := fyne.NewMenuItem("All Desktops", func() {
		if c.win.Pinned() {
			return
		}

		c.win.Pin()
	})
	if c.win.Pinned() {
		pin.Checked = true
	}
	desks = append(desks, pin)

	ret := fyne.NewMenuItem("Move to Desktop", func() {
		fynedesk.Instance().ShowMenuAt(fyne.NewMenu("", desks...),
			pos.Add(fyne.NewSize(40, 120)))

	})
	ret.ChildMenu = fyne.NewMenu("") // No-op to add the arrow...
	return ret
}

// CreateRenderer creates a new renderer for this border
//
// Implements: fyne.Widget
func (c *Border) CreateRenderer() fyne.WidgetRenderer {
	render := &coloredBoxRenderer{b: c, bg: canvas.NewRectangle(theme.BackgroundColor())}
	return render
}

// SetIcon tells the border to change the icon that should be used
func (c *Border) SetIcon(icon fyne.Resource) {
	if icon == nil {
		c.appIcon.Icon = nil
		return
	}

	c.appIcon.Icon = icon
}

// SetMaximized updates the state of the border maximize indicators and refreshes
func (c *Border) SetMaximized(isMax bool) {
	if isMax {
		c.max.Icon = theme.ViewRestoreIcon()
	} else {
		c.max.Icon = wmTheme.MaximizeIcon
	}
}

// SetTitle updates the title portion of this border and refreshes.
func (c *Border) SetTitle(title string) {
	c.title.Text = title
}

func newColoredHBox(focused bool, win fynedesk.Window, objs ...fyne.CanvasObject) *Border {
	ret := &Border{win: win}
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
	r.bg.Resize(r.b.Size()) // Not sure why this resize is needed, but it is...
	if r.b.win.Focused() {
		r.bg.FillColor = theme.BackgroundColor()
	} else {
		r.bg.FillColor = theme.DisabledButtonColor()
	}
	r.bg.Refresh()

	r.b.title.Color = theme.ForegroundColor()
	r.b.content.Refresh()
}
