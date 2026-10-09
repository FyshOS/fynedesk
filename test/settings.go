package test

import (
	"fyshos.com/tyde"

	"fyne.io/fyne/v2"
)

// Settings is a simple struct for managing settings within our tests
type Settings struct {
	background             string
	backgroundFill         string
	backgroundColor        string
	iconTheme              string
	launcherIcons          []string
	launcherDisableTaskbar bool
	borderButtonPosition   string
	clockFormatting        string
	computerType           string

	moduleNames []string

	narrowPanel  bool
	desktopCount int
}

// NewSettings returns an in-memory settings instance
func NewSettings() *Settings {
	return &Settings{}
}

// AddChangeListener is ignored for test instance
func (*Settings) AddChangeListener(listener func(tyde.DeskSettings)) {
}

// Background returns the path to background image (or "" if not set)
func (s *Settings) Background() string {
	return s.background
}

// SetBackground configures a background image path, passing "" removes the configuration
func (s *Settings) SetBackground(bg string) {
	s.background = bg
}

// BackgroundFill returns how the background image should fill the screen.
func (s *Settings) BackgroundFill() string {
	return s.backgroundFill
}

// SetBackgroundFill configures how the background image should fill the screen.
func (s *Settings) SetBackgroundFill(fill string) {
	s.backgroundFill = fill
}

// BackgroundColor returns the color drawn behind the background image, as a hex string.
func (s *Settings) BackgroundColor() string {
	return s.backgroundColor
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

// LauncherDisableTaskbar returns true if the taskbar should be disabled
func (s *Settings) LauncherDisableTaskbar() bool {
	return s.launcherDisableTaskbar
}

// SetLauncherDisableTaskbar allows configuring whether the taskbar should be disabled
func (s *Settings) SetLauncherDisableTaskbar(bar bool) {
	s.launcherDisableTaskbar = bar
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

// ComputerType returns the type of computer the desktop is configured for.
func (s *Settings) ComputerType() string {
	if s.computerType == "" {
		return tyde.ComputerLaptop
	}

	return s.computerType
}

// SetComputerType sets the type of computer the desktop is configured for.
func (s *Settings) SetComputerType(kind string) {
	s.computerType = kind
}

// DesktopCount returns how many virtual desktops are offered, 4 unless set.
func (s *Settings) DesktopCount() int {
	if s.desktopCount == 0 {
		return 4
	}

	return s.desktopCount
}

// SetDesktopCount sets how many virtual desktops are offered.
func (s *Settings) SetDesktopCount(count int) {
	s.desktopCount = count
}

// ScreenSaverClock returns if the text on the screensaver should be a clock.
func (s *Settings) ScreenSaverClock() bool {
	return true
}

// ScreenSaverLabel returns the string to use in the screensaver (if not a clock).
func (s *Settings) ScreenSaverLabel() string {
	return "FyshOS"
}

// ScreenSaverType returns whether this user should use FyshOS or XScreensaver savers.
func (s *Settings) ScreenSaverType() string {
	return "FyshOS"
}

// SetClockFormatting support setting the format that the clock should display
func (s *Settings) SetClockFormatting(format string) {
	if format == "24h" {
		s.clockFormatting = format
	} else {
		s.clockFormatting = "12h"
	}
}
