package wm

import (
	"fmt"
	"image/color"

	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/internal/icon"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewBorder creates a new window border for the given window details
func NewBorder(win fynedesk.Window, ico fyne.Resource, canMaximize bool) *Border {
	desk := fynedesk.Instance()
	border := &Border{win: win}
	border.ExtendBaseWidget(border)
	border.SetTitle(win.Properties().Title())
	border.SetContent(canvas.NewRectangle(color.Transparent))
	height := border.Theme().Size(theme.SizeNameWindowTitleBarHeight)

	if ico == nil {
		iconTheme := desk.Settings().IconTheme()
		app := icon.FindAppFromWinInfo(win, desk.IconProvider())
		if app != nil {
			ico = app.Icon(iconTheme, int(height*2))
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
	if fynedesk.Instance().Settings().BorderButtonPosition() == "Right" {
		buttonAlign = widget.ButtonAlignTrailing
	}
	border.Alignment = buttonAlign

	border.OnTappedIcon = func() {
		border.showMenu(border) // TODO appIcon position (estimate)
	}

	border.Icon = ico
	border.Refresh()
	return border
}

// Border represents a window border. It draws the title bar and provides functions to manipulate it.
type Border struct {
	container.InnerWindow
	win fynedesk.Window
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

func (c *Border) showMenu(_ fyne.CanvasObject) {
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
	menuPos := pos.AddXY(c.win.Size().Width-32, 0)
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

	return fyne.NewMenuItem("Move to Desktop...", func() {
		fynedesk.Instance().ShowMenuAt(fyne.NewMenu("", desks...),
			pos.Add(fyne.NewSize(40, 120)))

	})
}
