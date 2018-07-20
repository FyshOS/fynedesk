// +build !ci

package desktop

import "unsafe"

import "github.com/fyne-io/fyne/desktop"
import "github.com/fyne-io/fyne/theme"

// #cgo pkg-config: ecore ecore-input
// #include <stdlib.h>
// #include <Ecore.h>
// #include <Ecore_Input.h>
//
// void onMouseMove_cgo(void *data, int type, void *event_info);
import "C"

import "github.com/fyne-io/fyne"

func isEmbedded() bool {
	env := C.getenv(C.CString("DISPLAY"))
	if env != nil {
		return true
	}

	env = C.getenv(C.CString("WAYLAND_DISPLAY"))
	return env != nil
}

// NewDesktopWindow creates a new desktop window (fullscreen for main usage
// or a smaller window when used "embedded" for testing).
// This will automatically detect which mode to run in based on the presence
// of a DISPLAY or WAYLAND_DISPLAY environment variable.
func NewDesktopWindow(app fyne.App) fyne.Window {
	if isEmbedded() {
		return app.NewWindow("Fyne Desktop")
	}

	desk := desktop.CreateWindowWithEngine("drm")
	desk.SetFullscreen(true)

	return desk
}

//export onMouseMove
func onMouseMove(ew C.Ecore_Window, info *C.Ecore_Event_Mouse_Move) {
	canvas := fyne.GetCanvas(mouse)
	x := int(float32(info.x) / canvas.Scale())
	y := int(float32(info.y) / canvas.Scale())

	if !fyne.GetWindow(canvas).Fullscreen() {
		x -= theme.Padding()
		y -= theme.Padding()
	}
	mouse.Move(fyne.NewPos(x, y))
	canvas.Refresh(mouse)
}

func initInput() {
	C.ecore_event_handler_add(C.ECORE_EVENT_MOUSE_MOVE, (C.Ecore_Event_Handler_Cb)(unsafe.Pointer(C.onMouseMove_cgo)), nil)
}
