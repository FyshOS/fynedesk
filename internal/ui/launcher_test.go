package ui

import (
	"testing"

	"fyne.io/fyne/test"
	"fyne.io/fyne/widget"
	"github.com/stretchr/testify/assert"
)

func TestLauncher_ListMatches(t *testing.T) {
	names := []string{"App 1", "App 2", "Another"}
	desk := &testDesk{icons: newTestAppProvider(names), settings: &testSettings{}}
	launcher := newAppLauncher(desk)

	apps := launcher.appButtonListMatching("App")
	assert.Equal(t, 2, len(apps))
	assert.Equal(t, "App 1", apps[0].(*widget.Button).Text)
	assert.Equal(t, "App 2", apps[1].(*widget.Button).Text)

	apps = launcher.appButtonListMatching("ano")
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, "Another", apps[0].(*widget.Button).Text)

	apps = launcher.appButtonListMatching("miss")
	assert.Equal(t, 0, len(apps))
}

func TestLauncher_ListTyped(t *testing.T) {
	names := []string{"App 1", "App 2", "Another"}
	desk := &testDesk{icons: newTestAppProvider(names), settings: &testSettings{}}
	launcher := newAppLauncher(desk)

	assert.Equal(t, 0, len(launcher.appList.Objects))
	test.Type(launcher.entry, "App")
	assert.Equal(t, 2, len(launcher.appList.Objects))
	test.Type(launcher.entry, "Appy")
	assert.Equal(t, 0, len(launcher.appList.Objects))
}
