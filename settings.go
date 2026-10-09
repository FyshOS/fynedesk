package tyde

import "fyne.io/fyne/v2"

// The types of computer that Tyde can be configured for.
const (
	// ComputerDesktop is a machine with no battery and no touch screen.
	ComputerDesktop = "Desktop"
	// ComputerLaptop is a portable computer - it has a battery, but is driven
	// with a keyboard and pointer.
	ComputerLaptop = "Laptop"
	// ComputerTablet is a mobile computer - a battery and a touch screen.
	ComputerTablet = "Tablet"
)

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	Background() string
	BackgroundFill() string
	BackgroundColor() string
	IconTheme() string
	BorderButtonPosition() string
	ClockFormatting() string
	ComputerType() string
	DesktopCount() int // how many virtual desktops are offered
	NarrowWidgetPanel() bool

	LauncherIcons() []string
	LauncherDisableTaskbar() bool

	KeyboardModifier() fyne.KeyModifier
	ModuleNames() []string
	ScreenSaverType() string
	ScreenSaverClock() bool
	ScreenSaverLabel() string

	AddChangeListener(listener func(DeskSettings))
}
