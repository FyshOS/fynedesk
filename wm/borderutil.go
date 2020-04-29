package wm

import "fyne.io/fynedesk"

// ScaleToPixels calculates the pixels required to show a specified Fyne dimension on the given screen
func ScaleToPixels(i int, screen *fynedesk.Screen) int {
	return int(float32(i) * screen.CanvasScale())
}
