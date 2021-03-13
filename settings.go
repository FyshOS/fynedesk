package fynedesk

import deskDriver "fyne.io/fyne/v2/driver/desktop"

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	Background() string
	IconTheme() string
	ToolbarButtonPosition() string
	ClockFormatting() string

	LauncherIcons() []string
	LauncherIconSize() float32
	LauncherDisableTaskbar() bool
	LauncherDisableZoom() bool
	LauncherZoomScale() float32

	KeyboardModifier() deskDriver.Modifier
	ModuleNames() []string

	AddChangeListener(listener chan DeskSettings)
}
