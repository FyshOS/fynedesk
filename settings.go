package tyde

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
)

// DesktopName returns the name the user gave virtual desktop id, or its number
// counted from one when it has none.
func DesktopName(s DeskSettings, id int) string {
	names := s.DesktopNames()
	if id >= 0 && id < len(names) && strings.TrimSpace(names[id]) != "" {
		return names[id]
	}
	return strconv.Itoa(id + 1)
}

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
	DesktopCount() int      // how many virtual desktops are offered
	DesktopNames() []string // user-given desktop names, index by desktop; may be shorter than the count
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
