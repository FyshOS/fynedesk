package wm

import (
	"fyne.io/fyne"
	"github.com/godbus/dbus/v5"
)

var (
	server             *notifications
	lastNotificationID uint32
)

type Notification struct {
	ID          uint32
	Title, Body string
}

func NewNotification(title, body string) *Notification {
	lastNotificationID++

	item := &Notification{ID: lastNotificationID, Title: title, Body: body}

	return item
}

func SetNotificationListener(listen func(*Notification)) {
	s := startNotifications()

	s.listener = listen
	server = s
}

func SendNotification(n *Notification) {
	if server == nil || server.listener == nil {
		fyne.LogError("No notifications listener attached", nil)
		return
	}

	server.listener(n)
}

type notifications struct {
	notifs []Notification

	listener func(*Notification)
}

func (n *notifications) Notify(appName string, replacesID uint32, appIcon, summary, body string,
	actions []string, hints map[string]interface{}, timeout int32) (uint32, *dbus.Error) {
	item := NewNotification(summary, body)

	SendNotification(item)
	return item.ID, nil
}

func (n *notifications) GetServerInformation() (string, string, string, string, *dbus.Error) {
	return "FyneDesk", "Fyne.io", "0", "0", nil
}

func (n *notifications) GetCapabilities() ([]string, *dbus.Error) {
	return []string{"body", "icon-static", "persistence"}, nil
}

func startNotifications() *notifications {
	n := &notifications{}
	err := RegisterService(n, "/org/freedesktop/Notifications", "org.freedesktop.Notifications")
	if err != nil {
		fyne.LogError("Could not start DBus notifications server, using local only", err)
	}

	return n
}
