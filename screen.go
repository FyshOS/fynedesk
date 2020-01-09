package desktop

import (
	"os"
	"strconv"

	"fyne.io/fyne"
)

// ScreenList provides information about available physical screens for Fyne desktop
type ScreenList interface {
	Screens() []*Screen                                            // Screens returns a Screen type slice of each available physical screen
	Active() *Screen                                               // Active returns the screen index of the currently active screen
	Primary() *Screen                                              // Primary returns the screen index of the primary screen
	ScreenForWindow(Window) *Screen                                // Return the screen that a window is located on
	ScreenForGeometry(x int, y int, width int, height int) *Screen // Return the screen that a geometry is located on
}

// Screen provides relative information about a single physical screen
type Screen struct {
	Name                string  // Name is the randr provided name of the screen
	X, Y, Width, Height int     // Geometry of the screen
	Scale               float32 // Scale of this screen based on size and resolution
}

// CanvasScale calculates the scale for the contents of a desktop canvas on this screen
func (s *Screen) CanvasScale() float32 {
	user := userScale()
	if user == fyne.SettingsScaleAuto {
		user = 1.0
	}

	return s.Scale * user
}

func userScale() float32 {
	env := os.Getenv("FYNE_SCALE")

	if env != "" && env != "auto" {
		scale, err := strconv.ParseFloat(env, 32)
		if err == nil && scale != 0 {
			return float32(scale)
		}
		fyne.LogError("Error reading scale", err)
	}

	if env != "auto" {
		setting := fyne.CurrentApp().Settings().Scale()
		if setting != fyne.SettingsScaleAuto && setting != 0.0 {
			return setting
		}
	}

	return 1.0
}
