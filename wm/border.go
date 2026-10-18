package wm

import (
	"image/color"
	"strconv"

	"fyshos.com/tyde"
	"fyshos.com/tyde/internal/icon"
	wmTheme "fyshos.com/tyde/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewBorder creates a new window border for the given window details
func NewBorder(win tyde.Window, ico fyne.Resource, canMaximize bool) *Border {
	desk := tyde.Instance()
	border := &Border{win: win}
	border.ExtendBaseWidget(border)
	border.SetTitle(win.Properties().Title())
	border.SetContent(canvas.NewRectangle(color.Transparent))

	if ico == nil {
		iconTheme := desk.Settings().IconTheme()
		app := icon.FindAppFromWinInfo(win, desk.IconProvider())
		if app != nil {
			// load at twice the title bar height, so it is still sharp on touch version.
			ico = app.Icon(iconTheme, int(wmTheme.TitleHeight*2))
		}
	}

	if canMaximize {
		border.OnMaximized = func() {
			if win.Maximized() {
				win.Unmaximize()
			} else {
				win.Maximize()
			}
		}
	}
	if win.Maximized() {
		border.SetMaximized(true)
	}

	border.OnMinimized = func() {
		win.Iconify()
	}

	buttonAlign := widget.ButtonAlignLeading
	if tyde.Instance().Settings().BorderButtonPosition() == "Right" {
		buttonAlign = widget.ButtonAlignTrailing
	}
	border.Alignment = buttonAlign

	border.OnTappedIcon = func() {
		border.showMenu()
	}

	border.Icon = ico
	border.Refresh()
	return border
}

// Border represents a window border. It draws the title bar and provides functions to manipulate it.
type Border struct {
	container.InnerWindow
	win tyde.Window
}

// DoubleTapped is called when the user double taps a frame, it toggles the maximised state.
func (c *Border) DoubleTapped(*fyne.PointEvent) {
	if c.win.Maximized() {
		c.win.Unmaximize()
		return
	}
	c.win.Maximize()
}

// SetIcon updates the icon used in the window border.
func (c *Border) SetIcon(icon fyne.Resource) {
	c.Icon = icon
	c.Refresh()
}

func (c *Border) showMenu() {
	name := c.win.Properties().Title()
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
	x := c.win.Size().Width - 32
	if tyde.Instance().Settings().BorderButtonPosition() == "Right" {
		x = 0
	}
	menuPos := pos.AddXY(x, 0)
	menu := fyne.NewMenu("",
		title,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Minimize", func() {
			c.win.Iconify()
		}),
		max,
		fyne.NewMenuItemSeparator(),
		c.makeDesktopMenu(),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Close", func() {
			c.win.Close()
		}))

	tyde.Instance().ShowMenuAt(menu, menuPos)
}

func (c *Border) makeDesktopMenu() *fyne.MenuItem {
	settings := tyde.Instance().Settings()
	count := settings.DesktopCount()
	desks := make([]*fyne.MenuItem, count)
	for i := 0; i < count; i++ {
		deskID := i
		label := tyde.DesktopName(settings, i)
		if label == strconv.Itoa(i+1) { // unnamed: the bare number reads better as "Desktop N"
			label = "Desktop " + label
		}
		desks[i] = fyne.NewMenuItem(label, func() {
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

	item := fyne.NewMenuItem("Move to Desktop...", nil)
	item.ChildMenu = fyne.NewMenu("", append(desks, pin)...)
	return item
}
