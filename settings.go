package fynedesk

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	Background() string
	IconTheme() string
	ClockFormatting() string

	LauncherIcons() []string
	LauncherIconSize() float32
	LauncherDisableTaskbar() bool
	LauncherDisableZoom() bool
	LauncherZoomScale() float32

	ModuleNames() []string

	AddChangeListener(listener chan DeskSettings)
}
