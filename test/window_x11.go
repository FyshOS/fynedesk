package test

import "github.com/BurntSushi/xgb/xproto"

// just stubs so we can use the generic test Window in X11 tests

func (w *Window) ChildID() xproto.Window {
	return 0
}

func (w *Window) Expose() {
	// no-op
}

func (w *Window) FrameID() xproto.Window {
	return 0
}

func (w *Window) Geometry() (int, int, uint, uint) {
	return 0, 0, 10, 10
}

func (w *Window) NotifyBorderChange() {
	// no-op
}

func (w *Window) NotifyGeometry(int, int, uint, uint) {
	// no-op
}

func (w *Window) NotifyMoveResizeEnded() {
	// no-op
}

func (w *Window) NotifyFullscreen() {
	// no-op
}

func (w *Window) NotifyIconify() {
	// no-op
}

func (w *Window) NotifyMaximize() {
	// no-op
}

func (w *Window) NotifyUnFullscreen() {
	// no-op
}

func (w *Window) NotifyUnIconify() {
	// no-op
}

func (w *Window) NotifyUnMaximize() {
	// no-op
}

func (w *Window) NotifyMouseDrag(int16, int16) {
	// no-op
}

func (w *Window) NotifyMouseMotion(int16, int16) {
	// no-op
}

func (w *Window) NotifyMousePress(int16, int16) {
	// no-op
}

func (w *Window) NotifyMouseRelease(int16, int16) {
	// no-op
}

func (w *Window) Refresh() {
	// no-op
}

func (w *Window) SettingsChanged() {
	// no-op
}

func (w *Window) SizeMin() (uint, uint) {
	return 0, 0
}

func (w *Window) SizeMax() (int, int) {
	return -1, -1
}
