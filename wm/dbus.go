package wm

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

var conn *dbus.Conn

func connectBus() (*dbus.Conn, error) {
	if conn != nil {
		return conn, nil
	}

	c, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}
	// TODO cleanly shut down with conn.Close()

	conn = c
	return conn, nil
}

// RegisterService allows an object to be exported to the DBus messaging system.
// Methods on the object exposed can add an additional error parameter to the
// return types, in which case a non-nil error will send an error message
// instead of the object response.
func RegisterService(obj interface{}, path, iface string) error {
	conn, err := connectBus()
	if err != nil {
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
