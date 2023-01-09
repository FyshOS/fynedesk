//Note that you need to have github.com/knightpp/dbus-codegen-go installed from "custom" branch
//go:generate dbus-codegen-go -prefix org.kde -package notifier -output generated/notifier/status_notifier_item.go StatusNotifierItem.xml
//go:generate dbus-codegen-go -prefix org.kde -package watcher -output generated/watcher/status_notifier_watcher.go StatusNotifierWatcher.xml
//go:generate dbus-codegen-go -prefix com.canonical -package menu -output generated/menu/dbus_menu.go DbusMenu.xml

package systray

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/internal/icon"
	"fyshos.com/fynedesk/modules/systray/generated/menu"
	"fyshos.com/fynedesk/modules/systray/generated/notifier"
	"fyshos.com/fynedesk/modules/systray/generated/watcher"
	wmtheme "fyshos.com/fynedesk/theme"
)

const (
	path     = "/StatusNotifierWatcher"
	hostPath = "/StatusNotifierHost"
)

var resourceID = 0

func init() {
	fynedesk.RegisterModule(trayMeta)
}

var trayMeta = fynedesk.ModuleMetadata{
	Name:        "SystemTray",
	NewInstance: NewTray,
}

type tray struct {
	conn *dbus.Conn
	menu *menu.Dbusmenu

	box   *fyne.Container
	nodes map[dbus.Sender]*widget.Button
}

// NewTray creates a new module that will show a system tray in the status area
func NewTray() fynedesk.Module {
	iconSize := wmtheme.NarrowBarWidth
	grid := container.New(collapsingGridWrap(fyne.NewSize(iconSize, iconSize)))
	t := &tray{box: grid, nodes: make(map[dbus.Sender]*widget.Button)}

	conn, _ := dbus.ConnectSessionBus()
	t.conn = conn

	err := conn.ExportAll(struct{}{}, hostPath, "org.kde.StatusNotifierHost")
	if err != nil {
		log.Println("Err", err)
		return t
	}

	// TODO this is create watcher (optional)
	err = conn.ExportAll(t, path, "org.kde.StatusNotifierWatcher")
	if err != nil {
		log.Println("Err2", err)
		return t
	}

	_, err = conn.RequestName("org.kde.StatusNotifierWatcher", dbus.NameFlagDoNotQueue)
	if err != nil {
		log.Println("Failed to claim notifier watcher name", err)
		return t
	}

	_, err = prop.Export(conn, path, createPropSpec())
	if err != nil {
		log.Printf("Failed to export notifier item properties to bus")
		return t
	}

	node := introspect.Node{
		Name: path,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			watcher.IntrospectDataStatusNotifierWatcher,
		},
	}
	err = conn.Export(introspect.NewIntrospectable(&node), path,
		"org.freedesktop.DBus.Introspectable")
	if err != nil {
		log.Printf("Failed to export introspection %v", err)
		return t
	}
	// End TODO

	hostErr := t.RegisterStatusNotifierHost(conn.Names()[0], "")
	if hostErr != nil {
		fyne.LogError("Failed to register our host with the notifier watcher, maybe no watcher running? %v", hostErr)
		return t
	}

	watchErr := t.conn.AddMatchSignal(dbus.WithMatchInterface("org.freedesktop.DBus"), dbus.WithMatchObjectPath("/org/freedesktop/DBus"))
	if watchErr != nil {
		fyne.LogError("Failed to monitor systray name loss", watchErr)
		return t
	}

	c := make(chan *dbus.Signal, 10)
	t.conn.Signal(c)
	go func() {
		for v := range c {
			if v.Name != "org.freedesktop.DBus.NameOwnerChanged" {
				log.Println("Also", v.Name)
				continue
			}

			name := v.Body[0]
			newOwner := v.Body[2]
			if newOwner == "" {
				if item, ok := t.nodes[dbus.Sender(name.(string))]; ok {
					t.box.Remove(item)
					t.box.Refresh()
				}
			}
		}
	}()

	return t
}

func (t *tray) Destroy() {
}

func (t *tray) RegisterStatusNotifierItem(service string, sender dbus.Sender) (err *dbus.Error) {
	ni := notifier.NewStatusNotifierItem(t.conn.Object(string(sender), dbus.ObjectPath(service)))

	ico, ok := t.nodes[sender]
	if !ok {
		ico = widget.NewButton("", func() {
			if m, err := ni.GetMenu(t.conn.Context()); err == nil {
				t.showMenu(string(sender), m, ico)
				return
			}

			err := ni.Activate(t.conn.Context(), 5, 5)
			if err != nil { // try secondary if primary not known
				_ = ni.SecondaryActivate(t.conn.Context(), 5, 5)
			}
		})
		ico.Importance = widget.LowImportance
		t.nodes[sender] = ico
		t.box.Add(ico)
	}

	ic, _ := ni.GetIconPixmap(t.conn.Context())
	if len(ic) > 0 {
		img := pixelsToImage(ic[0])
		unique := strconv.Itoa(resourceID) + ".png"
		resourceID++
		w := &bytes.Buffer{}
		_ = png.Encode(w, img)
		ico.SetIcon(fyne.NewStaticResource(unique, w.Bytes()))
	} else {
		name, _ := ni.GetIconName(t.conn.Context())
		path, _ := ni.GetIconThemePath(t.conn.Context())
		fullPath := ""
		if path != "" {
			fullPath = filepath.Join(path, name+".png")
			if _, err := os.Stat(fullPath); err != nil { // not found, search instead
				fullPath = icon.FdoLookupIconPathInTheme("64", filepath.Join(path, "hicolor"), "", name)
			}
		} else {
			fullPath = icon.FdoLookupIconPath("", 64, name)
		}
		img, err := ioutil.ReadFile(fullPath)
		if err != nil {
			fyne.LogError("Failed to load status icon", err)
			ico.SetIcon(wmtheme.BrokenImageIcon)
		} else {
			ico.SetIcon(fyne.NewStaticResource(name, img))
		}
	}

	ico.Refresh()
	t.box.Refresh()

	return nil
}

func (t *tray) RegisterStatusNotifierHost(service string, sender dbus.Sender) (err *dbus.Error) {
	log.Println("Register Host", service, sender)

	e := watcher.Emit(t.conn, &watcher.StatusNotifierWatcher_StatusNotifierHostRegisteredSignal{
		Path: dbus.ObjectPath(service),
		Body: &watcher.StatusNotifierWatcher_StatusNotifierHostRegisteredSignalBody{},
	})
	// TODO: See the need of returning this error
	if e != nil {
		fyne.LogError("it was not emit the notification ", err)
	}
	return nil
}

func (t *tray) Metadata() fynedesk.ModuleMetadata {
	return trayMeta
}

func (t *tray) StatusAreaWidget() fyne.CanvasObject {
	return t.box
}

func (t *tray) parseMenu(parent int32, pos *fyne.Position, closer func()) fyne.CanvasObject {
	Y := pos.Y
	var items []*fyne.MenuItem
	_, l, _ := t.menu.GetLayout(t.conn.Context(), parent, 1, nil)
	for i, item := range l.V2 {
		data := item.Value().([]interface{})
		items = append(items, t.parseMenuItem(data[0].(int32), t.menu, data[1], pos, i, closer))

		Y += theme.TextSize() + theme.Padding()*2
	}
	m := fyne.NewMenu("", items...)
	return widget.NewMenu(m)
}

func (t *tray) parseMenuItem(id int32, menu *menu.Dbusmenu, in interface{}, pos *fyne.Position, off int, closer func()) *fyne.MenuItem {
	data := in.(map[string]dbus.Variant)
	ret := &fyne.MenuItem{}
	if ty, ok := data["type"]; ok {
		if ty.String() == "\"separator\"" {
			ret.IsSeparator = true
		}
	} else {
		ret.Label = fmt.Sprintf("%s", data["label"].Value())
		ret.Action = func() {
			err := menu.Event(t.conn.Context(), int32(id), "clicked", dbus.MakeVariant(id), uint32(time.Now().Unix()))
			if err != nil {
				fyne.LogError("Failed to message menu tap", err)
			}
			closer()
		}
	}

	if i, ok := data["icon-data"]; ok {
		ret.Icon = fyne.NewStaticResource(fmt.Sprintf("systray-icon-%d", id), i.Value().([]byte))
	}
	if e, ok := data["enabled"]; ok && e.Value() == false {
		ret.Disabled = true
	}

	if t, ok := data["toggle-type"]; ok && t.String() == "\"checkmark\"" {
		if s, ok := data["toggle-state"]; ok && s.Value() == true {
			ret.Checked = true
		}
	}

	if s, ok := data["children-display"]; ok && s.String() == "\"submenu\"" {
		ret.Action = func() {
			w := fyne.CurrentApp().Driver().(deskDriver.Driver).CreateSplashWindow()
			w.SetOnClosed(closer)
			childPos := &fyne.Position{}

			w.SetContent(t.parseMenu(id, childPos, func() {
				w.Close()
				closer()
			}))

			size := w.Content().MinSize()
			w.Resize(size)
			sub := (*pos).AddXY(-size.Width, float32(off)*(18+theme.Padding()*4))
			screen := fynedesk.Instance().Screens().Primary()
			if sub.Y+size.Height > float32(screen.Height)/screen.CanvasScale() {
				sub.Y = float32(screen.Height)/screen.CanvasScale() - size.Height
			}
			childPos.X, childPos.Y = sub.X, sub.Y

			fynedesk.Instance().WindowManager().ShowOverlay(w, size, *childPos)
		}

		ret.ChildMenu = fyne.NewMenu("")
	}
	return ret
}

func (t *tray) showMenu(sender string, name dbus.ObjectPath, from fyne.CanvasObject) {
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(from)
	w := fyne.CurrentApp().Driver().(deskDriver.Driver).CreateSplashWindow()
	t.menu = menu.NewDbusmenu(t.conn.Object(sender, name))
	w.SetContent(t.parseMenu(0, &pos, func() {
		w.Close()
	}))

	size := w.Content().MinSize()
	w.Resize(size)

	pos.X -= size.Width
	screen := fynedesk.Instance().Screens().Primary()
	if pos.Y+size.Height > float32(screen.Height)/screen.CanvasScale() {
		pos.Y = float32(screen.Height)/screen.CanvasScale() - size.Height
	}
	fynedesk.Instance().WindowManager().ShowOverlay(w, size, pos)
}

func createPropSpec() map[string]map[string]*prop.Prop {
	return map[string]map[string]*prop.Prop{
		"org.kde.StatusNotifierWatcher": {
			"RegisteredStatusNotifierItems": {
				Value:    []string{},
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
			"IsStatusNotifierHostRegistered": {
				Value:    true,
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
			"ProtocolVersion": {
				Value:    int32(25),
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
		},
	}
}

type img struct {
	w, h int
	data []byte
}

func (i *img) ColorModel() color.Model {
	return color.NRGBAModel
}

func (i *img) Bounds() image.Rectangle {
	return image.Rect(0, 0, i.w, i.h)
}

func (i *img) At(x, y int) color.Color {
	off := (y*i.w + x) * 4

	a, r, g, b := i.data[off], i.data[off+1], i.data[off+2], i.data[off+3]

	return color.NRGBA{r, g, b, a}
}

func pixelsToImage(in struct {
	V0 int32
	V1 int32
	V2 []byte
}) image.Image {
	return &img{int(in.V0), int(in.V1), in.V2}
}
