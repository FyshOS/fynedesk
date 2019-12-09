package desktop

// Screens provides information about available physical screens for Fyne desktop
type Screens interface {
	Screens() []*Head                               // Screens returns a Screen type slice of each available physical screen
	Active() *Head                                  // Active returns the screen index of the currently active screen
	Primary() *Head                                 // Primary returns the screen index of the primary screen
	Scale() float32                                 // Return the scale calculated for use across screens
	ScreenForWindow(windowX int, windowY int) *Head // Return the screen that a window is located on
}

// Head provides relative information about a single physical screen
type Head struct {
	Name                      string // Name is the randr provided name of the screen
	X, Y, Width, Height       int    // Geometry of the screen
	ScaledX, ScaledY          int    // Scaled position of the screen
	ScaledWidth, ScaledHeight int    // Scaled size of the screen
}
