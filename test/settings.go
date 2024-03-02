package test

import (
	"fyne.io/fyne/v2"
	"fyshos.com/fynedesk"
)

// Settings is a simple struct for managing settings within our tests
type Settings struct {
	background             string
	iconTheme              string
	launcherIcons          []string
	launcherIconSize       float32
	launcherZoomScale      float32
	launcherDisableZoom    bool
	launcherDisableTaskbar bool
	borderButtonPosition   string
	clockFormatting        string

	moduleNames []string

	narrowPanel, narrowLeftLauncher bool
}

// NewSettings returns an in-memory settings instance
func NewSettings() *Settings {
	return &Settings{}
}

// AddChangeListener is ignored for test instance
func (*Settings) AddChangeListener(listener chan fynedesk.DeskSettings) {
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
func (s *Settings) LauncherIconSize() float32 {
	if s.launcherIconSize == 0 {
		return 32
	}
	return s.launcherIconSize
}

// SetLauncherIconSize allows configuring the icon size in app launcher
func (s *Settings) SetLauncherIconSize(size float32) {
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
func (s *Settings) LauncherZoomScale() float32 {
	if s.launcherZoomScale == 0 {
		return 1.5
	}
	return s.launcherZoomScale
}

// SetLauncherZoomScale supports setting the scale value for hovered bar icons
func (s *Settings) SetLauncherZoomScale(scale float32) {
	s.launcherZoomScale = scale
}

// KeyboardModifier returns the preferred keyboard modifier for shortcuts.
func (s *Settings) KeyboardModifier() fyne.KeyModifier {
	return fyne.KeyModifierSuper
}

// ModuleNames returns the names of modules that should be enabled
func (s *Settings) ModuleNames() []string {
	return s.moduleNames
}

// SetModuleNames supports configuring the modules that should be loaded
func (s *Settings) SetModuleNames(mods []string) {
	s.moduleNames = mods
}

// NarrowLeftLauncher returns true when the user requested a narrow launcher bar on the left.
func (s *Settings) NarrowLeftLauncher() bool {
	return s.narrowLeftLauncher
}

// SetNarrowLeftLauncher allows tests to specify the value for a narrow left hand launcher.
func (s *Settings) SetNarrowLeftLauncher(narrow bool) {
	s.narrowLeftLauncher = narrow
}

// NarrowWidgetPanel returns true when the user requested a narrow widget panel.
func (s *Settings) NarrowWidgetPanel() bool {
	return s.narrowPanel
}

// SetNarrowWidgetPanel allows tests to specify the value for a narrow widget panel.
func (s *Settings) SetNarrowWidgetPanel(narrow bool) {
	s.narrowPanel = narrow
}

// BorderButtonPosition returns the position of the toolbar buttons.
func (s *Settings) BorderButtonPosition() string {
	return s.borderButtonPosition
}

// SetBorderButtonPosition sets the toolbar button position.
func (s *Settings) SetBorderButtonPosition(pos string) {
	s.borderButtonPosition = pos
}

// ClockFormatting returns the format that the clock uses for displaying the time. Either 12h or 24h.
func (s *Settings) ClockFormatting() string {
	return s.clockFormatting
}

// SetClockFormatting support setting the format that the clock should display
func (s *Settings) SetClockFormatting(format string) {
	if format == "24h" {
		s.clockFormatting = format
	} else {
		s.clockFormatting = "12h"
	}
}
