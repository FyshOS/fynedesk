package wm

import "fyne.io/fynedesk"

// ScaleToPixels calculates the pixels required to show a specified Fyne dimension on the given screen
func ScaleToPixels(i float32, screen *fynedesk.Screen) int {
	return int(i * screen.CanvasScale())
}
