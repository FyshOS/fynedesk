// +build linux

package wm

import "github.com/BurntSushi/xgb/xproto"

func (x *x11WM) transientChildAdd(leader xproto.Window, child xproto.Window) {
	for _, win := range x.transientMap[leader] {
		if win == child {
			return
		}
	}
	x.transientMap[leader] = append(x.transientMap[leader], child)
}

func (x *x11WM) transientChildRemove(leader xproto.Window, child xproto.Window) {
	for i, win := range x.transientMap[leader] {
		if win == child {
			x.transientMap[leader] = append(x.transientMap[leader][:i], x.transientMap[leader][i+1:]...)
		}
	}
}

// Window C could be Transient for window B which is transient for WindowA - We sometimes need the very top level
func (x *x11WM) transientTopLeaderGet(child xproto.Window) xproto.Window {
	var topLeader xproto.Window
	for child != 0 {
		topLeader = child
		child = windowTransientForGet(x.x, child)
	}
	return topLeader
}

func (x *x11WM) transientLeaderRemove(leader xproto.Window) {
	delete(x.transientMap, leader)
}
