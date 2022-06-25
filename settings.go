package fynedesk

import "fyne.io/fyne/v2"

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	Background() string
	IconTheme() string
	BorderButtonPosition() string
	ClockFormatting() string

	LauncherIcons() []string
	LauncherIconSize() float32
	LauncherDisableTaskbar() bool
	LauncherDisableZoom() bool
	LauncherZoomScale() float32

	KeyboardModifier() fyne.KeyModifier
	ModuleNames() []string

	AddChangeListener(listener chan DeskSettings)
}
