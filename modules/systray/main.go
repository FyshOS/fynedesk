//Note that you need to have github.com/knightpp/dbus-codegen-go installed from "custom" branch
//go:generate dbus-codegen-go -prefix org.kde -package notifier -output generated/notifier/status_notifier_item.go StatusNotifierItem.xml
//go:generate dbus-codegen-go -prefix org.kde -package watcher -output generated/watcher/status_notifier_watcher.go StatusNotifierWatcher.xml
//go:generate dbus-codegen-go -prefix com.canonical -package menu -output generated/menu/dbus_menu.go DbusMenu.xml

package status

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/icon"
	"fyne.io/fynedesk/modules/systray/generated/notifier"
	"fyne.io/fynedesk/modules/systray/generated/watcher"
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
	conn  *dbus.Conn
	box   *fyne.Container
	nodes map[dbus.Sender]*widget.Button
}

// NewTray creates a new module that will show a system tray in the status area
func NewTray() fynedesk.Module {
	t := &tray{box: container.NewHBox(layout.NewSpacer()),
		nodes: make(map[dbus.Sender]*widget.Button)}

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
			ni.Activate(t.conn.Context(), 5, 5)
		})
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
		} else {
			fullPath = icon.FdoLookupIconPath("", 64, name)
		}
		// TODO handle errors
		img, _ := ioutil.ReadFile(fullPath)
		ico.SetIcon(fyne.NewStaticResource(name, img))
	}

	ico.Refresh()
	t.box.Refresh()

	return nil
}

func (t *tray) RegisterStatusNotifierHost(service string, sender dbus.Sender) (err *dbus.Error) {
	log.Println("Register Host", service, sender)

	watcher.Emit(t.conn, &watcher.StatusNotifierWatcher_StatusNotifierHostRegisteredSignal{
		Path: dbus.ObjectPath(service),
		Body: &watcher.StatusNotifierWatcher_StatusNotifierHostRegisteredSignalBody{},
	})
	return nil
}

func (t *tray) Metadata() fynedesk.ModuleMetadata {
	return trayMeta
}

func (t *tray) StatusAreaWidget() fyne.CanvasObject {
	return t.box
}

func createPropSpec() map[string]map[string]*prop.Prop {
	return map[string]map[string]*prop.Prop{
		"org.kde.StatusNotifierWatcher": {
			"RegisteredStatusNotifierItems": {
				[]string{},
				false,
				prop.EmitTrue,
				nil,
			},
			"IsStatusNotifierHostRegistered": {
				true,
				false,
				prop.EmitTrue,
				nil,
			},
			"ProtocolVersion": {
				int32(25),
				false,
				prop.EmitTrue,
				nil,
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
