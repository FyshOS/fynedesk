package desktop

// #cgo pkg-config: ecore ecore-input
// #include <Ecore.h>
// #include <Ecore_Input.h>
//
// void onMouseMove_cgo(void *data, int type, void *event_info);
import "C"

import "unsafe"

import "github.com/fyne-io/fyne"
import "github.com/fyne-io/fyne/canvas"
import "github.com/fyne-io/fyne/theme"

import wmtheme "github.com/fyne-io/desktop/theme"

var mouse *canvas.Image

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

func newMouse() fyne.CanvasObject {
	mouse = canvas.NewImageFromFile(wmtheme.PointerDefault.CachePath())
	mouse.Options.RepeatEvents = true
	mouse.Resize(fyne.NewSize(24, 24))

	C.ecore_event_handler_add(C.ECORE_EVENT_MOUSE_MOVE, (C.Ecore_Event_Handler_Cb)(unsafe.Pointer(C.onMouseMove_cgo)), nil)

	return mouse
}
