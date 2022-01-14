//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package wm

import (
	"github.com/BurntSushi/xgb/xproto"
)

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

func (x *x11WM) transientLeaderRemove(leader xproto.Window) {
	delete(x.transientMap, leader)
}
