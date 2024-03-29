package desktops

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"fyshos.com/fynedesk"
)

const deskCount = 4

var desksMeta = fynedesk.ModuleMetadata{
	Name:        "Virtual Desktops",
	NewInstance: newDesktops,
}

type desktops struct {
	current int
	gui     *pager
}

func (d *desktops) DesktopChangeNotify(id int) {
	oldID := d.current
	d.current = id
	d.gui.refreshFrom(oldID)
}

func (d *desktops) Destroy() {
}

func (d *desktops) Metadata() fynedesk.ModuleMetadata {
	return desksMeta
}

func (d *desktops) Shortcuts() map[*fynedesk.Shortcut]func() {
	mapping := make(map[*fynedesk.Shortcut]func(), deskCount+2)
	for i := 0; i < deskCount; i++ {
		id := strconv.Itoa(i + 1)
		deskID := i
		mapping[&fynedesk.Shortcut{Name: "Switch to Desktop " + id, KeyName: fyne.KeyName(id), Modifier: fynedesk.UserModifier}] = func() {
			d.setDesktop(deskID)
		}
	}

	mapping[&fynedesk.Shortcut{Name: "Switch to Previous Desktop", KeyName: fyne.KeyUp, Modifier: fynedesk.UserModifier}] = func() {
		if d.current == 0 {
			return
		}
		d.setDesktop(d.current - 1)
	}
	mapping[&fynedesk.Shortcut{Name: "Switch to Next Desktop", KeyName: fyne.KeyDown, Modifier: fynedesk.UserModifier}] = func() {
		if d.current == deskCount-1 {
			return
		}
		d.setDesktop(d.current + 1)
	}
	return mapping
}

func (d *desktops) StatusAreaWidget() fyne.CanvasObject {
	return container.NewStack(d.gui.buttons, d.gui.wins, d.gui.labels)
}

func (d *desktops) setDesktop(id int) {
	oldID := d.current
	d.current = id
	fynedesk.Instance().SetDesktop(id)
	d.gui.refreshFrom(oldID)
}

// newDesktops creates a new module that will manage virtual desktops and display a pager widget.
func newDesktops() fynedesk.Module {
	d := &desktops{}
	d.gui = newPager(d)
	return d
}
