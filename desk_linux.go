// +build !ci

package desktop

import "unsafe"

import "fyne.io/fyne"
import "fyne.io/fyne/driver/efl"
import "fyne.io/fyne/theme"

// #cgo pkg-config: ecore ecore-input
// #include <stdlib.h>
// #include <Ecore.h>
// #include <Ecore_Input.h>
//
// void onMouseMove_cgo(void *data, int type, void *event_info);
import "C"

var desk fyne.Window

func isEmbedded() bool {
	env := C.getenv(C.CString("DISPLAY"))
	if env != nil {
		return true
	}

	env = C.getenv(C.CString("WAYLAND_DISPLAY"))
	return env != nil
}

// newDesktopWindow will return a new window based the current environment.
// When running in an existing desktop then load a window.
// Otherwise let's return a desktop root!
func newDesktopWindow(app fyne.App) fyne.Window {
	if isEmbedded() {
		desk = app.NewWindow("Fyne Desktop")
		desk.SetPadded(false)
		return desk
	}

	desk = efl.CreateWindowWithEngine("drm")
	desk.SetFullScreen(true)
	desk.SetPadded(false)

	return desk
}

//export onMouseMove
func onMouseMove(ew C.Ecore_Window, info *C.Ecore_Event_Mouse_Move) {
	canvas := desk.Canvas()
	x := int(float32(info.x) / canvas.Scale())
	y := int(float32(info.y) / canvas.Scale())

	if !desk.FullScreen() {
		x -= theme.Padding()
		y -= theme.Padding()
	}
	mouse.Move(fyne.NewPos(x, y))
	canvas.Refresh(mouse)
}

func initInput() {
	C.ecore_event_handler_add(C.ECORE_EVENT_MOUSE_MOVE, (C.Ecore_Event_Handler_Cb)(unsafe.Pointer(C.onMouseMove_cgo)), nil)
}
