package ui

import (
	"testing"

	"fyshos.com/fynedesk"
	wmTest "fyshos.com/fynedesk/test"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestLauncher_ListMatches(t *testing.T) {
	setupIcons("App 1", "App 2", "Another")
	launcher := newAppPicker("Test", func(data fynedesk.AppData) {})

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
	setupIcons("App 1", "App 2", "Another")
	launcher := newAppPicker("Test", func(data fynedesk.AppData) {})

	assert.Equal(t, 0, len(launcher.appList.Objects))
	test.Type(launcher.entry, "App")
	assert.Equal(t, 2, len(launcher.appList.Objects))
	test.Type(launcher.entry, "Appy")
	assert.Equal(t, 0, len(launcher.appList.Objects))
}

func TestLauncher_ListActive(t *testing.T) {
	setupIcons("App 1", "App 2", "Another")
	launcher := newAppPicker("Test", func(data fynedesk.AppData) {})

	assert.Equal(t, 0, len(launcher.appList.Objects))
	assert.Equal(t, 0, launcher.activeIndex)
	test.Type(launcher.entry, "App")
	launcher.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	assert.Equal(t, 1, launcher.activeIndex)
	assert.Equal(t, widget.MediumImportance, launcher.appList.Objects[0].(*widget.Button).Importance)
	assert.Equal(t, widget.HighImportance, launcher.appList.Objects[1].(*widget.Button).Importance)
}

func TestLauncher_setActiveIndex(t *testing.T) {
	setupIcons("App 1", "App 2", "Another")
	launcher := newAppPicker("Test", func(data fynedesk.AppData) {})

	launcher.appList.Objects = launcher.appButtonListMatching("App")
	assert.Equal(t, 0, launcher.activeIndex)

	launcher.setActiveIndex(1)
	assert.Equal(t, 1, launcher.activeIndex)

	launcher.setActiveIndex(2)
	assert.Equal(t, 1, launcher.activeIndex)

	launcher.setActiveIndex(-1)
	assert.Equal(t, 1, launcher.activeIndex)
}

func setupIcons(icons ...string) {
	desk := wmTest.NewDesktop()
	desk.SetIconProvider(wmTest.NewAppProvider(icons...))
	fynedesk.SetInstance(desk)
}
