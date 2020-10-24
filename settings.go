package fynedesk

// DeskSettings describes the configuration options available for Fyne desktop
type DeskSettings interface {
	Background() string
	IconTheme() string
	ClockFormatting() string

	LauncherIcons() []string
	LauncherIconSize() int
	LauncherDisableTaskbar() bool
	LauncherDisableZoom() bool
	LauncherZoomScale() float64

	ModuleNames() []string

	AddChangeListener(listener chan DeskSettings)
}
