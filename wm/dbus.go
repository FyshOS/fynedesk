package wm

import (
	"errors"

	"fyne.io/fyne"

	"github.com/godbus/dbus/v5"
)

// CallMethod simplifies calling a function on the DBus message system.
// The name of the method and its interface and path are passed as strings.
// The in parameters must be in the correct number and type to match the
// requested method and the returned parameters will be returned from this
// method. If an error occurred the last return parameter will be set.
func CallMethod(in []interface{}, path, iface, meth string) ([]interface{}, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		fyne.LogError("Error opening DBus connection", err)
		return nil, err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fyne.LogError("Error closing DBus connection", err)
		}
	}()

	obj := conn.Object(iface, dbus.ObjectPath(path))
	call := obj.Call(meth, 0, in...)
	if call.Err != nil {
		return nil, call.Err
	}

	return call.Body, nil
}

// RegisterService allows an object to be exported to the DBus messaging system.
// Methods on the object exposed can add an additional error parameter to the
// return types, in which case a non-nil error will send an error message
// instead of the object response.
func RegisterService(obj interface{}, path, iface string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		fyne.LogError("Error accessing to shared DBus connection", err)
		return err
	}

	err = conn.ExportAll(obj, dbus.ObjectPath(path), iface)
	if err != nil {
		return err
	}

	reply, err := conn.RequestName(iface, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return errors.New("name already taken")
	}

	return nil
}
