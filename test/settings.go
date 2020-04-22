package test

import "fyne.io/fynedesk"

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

func NewSettings() *Settings {
	return &Settings{}
}

func (*Settings) AddChangeListener(listener chan fynedesk.DeskSettings) {
	return
}

func (s *Settings) Background() string {
	return s.background
}

func (s *Settings) SetBackground(bg string) {
	s.background = bg
}

func (s *Settings) IconTheme() string {
	return s.iconTheme
}

func (s *Settings) SetIconTheme(theme string) {
	s.iconTheme = theme
}

func (s *Settings) LauncherIcons() []string {
	return s.launcherIcons
}

func (s *Settings) SetLauncherIcons(icons []string) {
	s.launcherIcons = icons
}

func (s *Settings) LauncherIconSize() int {
	if s.launcherIconSize == 0 {
		return 32
	}
	return s.launcherIconSize
}

func (s *Settings) SetLauncherIconSize(size int) {
	s.launcherIconSize = size
}

func (s *Settings) LauncherDisableTaskbar() bool {
	return s.launcherDisableTaskbar
}

func (s *Settings) SetLauncherDisableTaskbar(bar bool) {
	s.launcherDisableTaskbar = bar
}

func (s *Settings) LauncherDisableZoom() bool {
	return s.launcherDisableZoom
}

func (s *Settings) SetLauncherDisableZoom(zoom bool) {
	s.launcherDisableZoom = zoom
}

func (s *Settings) LauncherZoomScale() float64 {
	if s.launcherZoomScale == 0 {
		return 1.0
	}
	return s.launcherZoomScale
}

func (s *Settings) SetLauncherZoomScale(scale float64) {
	s.launcherZoomScale = scale
}

func (s *Settings) ModuleNames() []string {
	return s.moduleNames
}

func (s *Settings) SetModuleNames(mods []string) {
	s.moduleNames = mods
}
