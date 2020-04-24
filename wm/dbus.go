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
// Methods on that object to be exposed need to return *dbus.Error as their last parameter.
func RegisterService(obj interface{}, path, iface string) error {
	conn, err := connectBus()
	if err != nil {
		return err
	}

	err = conn.Export(obj, dbus.ObjectPath(path), iface)
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
