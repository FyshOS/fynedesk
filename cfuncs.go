// +build !ci

package desktop

/*
#cgo pkg-config: evas ecore-evas
#cgo CFLAGS: -DEFL_BETA_API_SUPPORT=1
#include <Evas.h>
#include <Ecore.h>
#include <Ecore_Evas.h>
#include <Ecore_Input.h>

void onMouseMove_cgo(void *data, int type, void *event_info)
{
	void onMouseMove(Ecore_Window, Ecore_Event_Mouse_Move*);
	Ecore_Event_Mouse_Move *move_ev = (Ecore_Event_Mouse_Move *) event_info;
	onMouseMove(move_ev->window, move_ev);
}

*/
import "C"
