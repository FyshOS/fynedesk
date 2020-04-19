package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fynedesk"
)

func TestDeskSettings_IsModuleEnabled(t *testing.T) {
	s := &testSettings{moduleNames: []string{"Yes", "maybe"}}

	assert.True(t, isModuleEnabled("Yes", s))
	assert.True(t, isModuleEnabled("maybe", s))
	assert.False(t, isModuleEnabled("Maybe", s))
	assert.False(t, isModuleEnabled("No", s))
}

type testSettings struct {
	background             string
	iconTheme              string
	launcherIcons          []string
	launcherIconSize       int
	launcherZoomScale      float64
	launcherDisableZoom    bool
	launcherDisableTaskbar bool

	moduleNames []string
}

func (ts *testSettings) IconTheme() string {
	return ts.iconTheme
}

func (ts *testSettings) Background() string {
	return ts.background
}

func (ts *testSettings) LauncherIcons() []string {
	return ts.launcherIcons
}

func (ts *testSettings) LauncherIconSize() int {
	if ts.launcherIconSize == 0 {
		return 32
	}
	return ts.launcherIconSize
}

func (ts *testSettings) LauncherDisableTaskbar() bool {
	return ts.launcherDisableTaskbar
}

func (ts *testSettings) LauncherDisableZoom() bool {
	return ts.launcherDisableZoom
}

func (ts *testSettings) LauncherZoomScale() float64 {
	if ts.launcherZoomScale == 0 {
		return 1.0
	}
	return ts.launcherZoomScale
}

func (ts *testSettings) ModuleNames() []string {
	return ts.moduleNames
}

func (*testSettings) AddChangeListener(listener chan fynedesk.DeskSettings) {
	return
}
