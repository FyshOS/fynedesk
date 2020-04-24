package test

import "fyne.io/fynedesk"

// Settings is a simple struct for managing settings within our tests
type Settings struct {
	background             string
	iconTheme              string
	launcherIcons          []string
	launcherIconSize       int
	launcherZoomScale      float64
	launcherDisableZoom    bool
	launcherDisableTaskbar bool

	moduleNames []string
}

// NewSettings returns an in-memory settings instance
func NewSettings() *Settings {
	return &Settings{}
}

// AddChangeListener is ignored for test instance
func (*Settings) AddChangeListener(listener chan fynedesk.DeskSettings) {
	return
}

// Background returns the path to background image (or "" if not set)
func (s *Settings) Background() string {
	return s.background
}

// SetBackground configures a background image path, passing "" removes the configuration
func (s *Settings) SetBackground(bg string) {
	s.background = bg
}

// IconTheme returns the configured icon theme
func (s *Settings) IconTheme() string {
	return s.iconTheme
}

// SetIconTheme supports setting the chosen icon theme
func (s *Settings) SetIconTheme(theme string) {
	s.iconTheme = theme
}

// LauncherIcons returns the names of the apps to appear in the launcher
func (s *Settings) LauncherIcons() []string {
	return s.launcherIcons
}

// SetLauncherIcons configures the app to be included in the launcher
func (s *Settings) SetLauncherIcons(icons []string) {
	s.launcherIcons = icons
}

// LauncherIconSize returns the standard (non-zoomed) icon size for app launcher
func (s *Settings) LauncherIconSize() int {
	if s.launcherIconSize == 0 {
		return 32
	}
	return s.launcherIconSize
}

// SetLauncherIconSize allows configuring the icon size in app launcher
func (s *Settings) SetLauncherIconSize(size int) {
	s.launcherIconSize = size
}

// LauncherDisableTaskbar returns true if the taskbar should be disabled
func (s *Settings) LauncherDisableTaskbar() bool {
	return s.launcherDisableTaskbar
}

// SetLauncherDisableTaskbar allows configuring whether the taskbar should be disabled
func (s *Settings) SetLauncherDisableTaskbar(bar bool) {
	s.launcherDisableTaskbar = bar
}

// LauncherDisableZoom returns true if zoom is disabled on the launcher
func (s *Settings) LauncherDisableZoom() bool {
	return s.launcherDisableZoom
}

// SetLauncherDisableZoom allows configuring whether the taskbar should disable zooming
func (s *Settings) SetLauncherDisableZoom(zoom bool) {
	s.launcherDisableZoom = zoom
}

// LauncherZoomScale returns how much the icons should zoom when hovered
func (s *Settings) LauncherZoomScale() float64 {
	if s.launcherZoomScale == 0 {
		return 1.5
	}
	return s.launcherZoomScale
}

// SetLauncherZoomScale supports setting the scale value for hovered bar icons
func (s *Settings) SetLauncherZoomScale(scale float64) {
	s.launcherZoomScale = scale
}

// ModuleNames returns the names of modules that should be enabled
func (s *Settings) ModuleNames() []string {
	return s.moduleNames
}

// SetModuleNames supports configuring the modules that should be loaded
func (s *Settings) SetModuleNames(mods []string) {
	s.moduleNames = mods
}
