package test

import (
	"strings"

	"fyne.io/fyne"

	"fyne.io/fynedesk"
	wmTheme "fyne.io/fynedesk/theme"
)

type testAppData struct {
	name string
}

// NewAppData returns a new test app icon with the specified name
func NewAppData(name string) fynedesk.AppData {
	return &testAppData{name: name}
}

func (tad *testAppData) Name() string {
	return tad.name
}

func (tad *testAppData) Run([]string) error {
	return nil
}

func (tad *testAppData) Icon(theme string, size int) fyne.Resource {
	if theme == "" {
		return nil
	} else if theme == "Maximize" {
		return wmTheme.MaximizeIcon
	}
	return wmTheme.IconifyIcon
}

type testAppProvider struct {
	screens []*fynedesk.Screen
	apps    []fynedesk.AppData
}

// NewAppProvider returns a simple provider of applications from the provided list of app names
func NewAppProvider(appNames ...string) fynedesk.ApplicationProvider {
	provider := &testAppProvider{}

	for _, name := range appNames {
		provider.apps = append(provider.apps, NewAppData(name))
	}

	return provider
}

func (tap *testAppProvider) AvailableApps() []fynedesk.AppData {
	return tap.apps
}

func (tap *testAppProvider) AvailableThemes() []string {
	return nil
}

func (tap *testAppProvider) FindAppFromName(appName string) fynedesk.AppData {
	return &testAppData{name: appName}
}

func (tap *testAppProvider) FindAppFromWinInfo(win fynedesk.Window) fynedesk.AppData {
	return &testAppData{}
}

func (tap *testAppProvider) FindAppsMatching(pattern string) []fynedesk.AppData {
	var ret []fynedesk.AppData
	for _, app := range tap.apps {
		if !strings.Contains(strings.ToLower(app.Name()), strings.ToLower(pattern)) {
			continue
		}

		ret = append(ret, app)
	}

	return ret
}

func (tap *testAppProvider) DefaultApps() []fynedesk.AppData {
	return nil
}
