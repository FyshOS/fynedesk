package test

import (
	"strings"

	"github.com/FyshOS/appie"

	"fyne.io/fyne/v2"

	"fyshos.com/fynedesk"
	wmTheme "fyshos.com/fynedesk/theme"
)

type testAppData struct {
	categories, mimes []string
	name              string
}

// NewAppData returns a new test app icon with the specified name
func NewAppData(name string) appie.AppData {
	return &testAppData{name: name}
}

func (tad *testAppData) Name() string {
	return tad.name
}

func (tad *testAppData) Run([]string) error {
	return nil
}

func (tad *testAppData) RunWithParameters([]string, []string) error {
	return nil
}

func (tad *testAppData) Categories() []string {
	return tad.categories
}

func (tad *testAppData) Hidden() bool {
	return false
}

func (tad *testAppData) Icon(theme string, size int) fyne.Resource {
	if theme == "" {
		return nil
	} else if theme == "Maximize" {
		return wmTheme.MaximizeIcon
	}
	return wmTheme.IconifyIcon
}

func (tad *testAppData) MimeTypes() []string {
	return tad.mimes
}

func (tad *testAppData) Source() *appie.AppSource {
	return nil
}

type testAppProvider struct {
	apps []appie.AppData
}

// NewAppProvider returns a simple provider of applications from the provided list of app names
func NewAppProvider(appNames ...string) appie.Provider {
	provider := &testAppProvider{}

	for _, name := range appNames {
		provider.apps = append(provider.apps, NewAppData(name))
	}

	return provider
}

func (tap *testAppProvider) AvailableApps() []appie.AppData {
	return tap.apps
}

func (tap *testAppProvider) AvailableThemes() []string {
	return nil
}

func (tap *testAppProvider) ClearCache() {
}

func (tap *testAppProvider) FindAppFromName(appName string) appie.AppData {
	return &testAppData{name: appName}
}

func (tap *testAppProvider) FindAppFromWinInfo(win fynedesk.Window) appie.AppData {
	return &testAppData{}
}

func (tap *testAppProvider) FindAppsMatching(pattern string) []appie.AppData {
	var ret []appie.AppData
	for _, app := range tap.apps {
		if !strings.Contains(strings.ToLower(app.Name()), strings.ToLower(pattern)) {
			continue
		}

		ret = append(ret, app)
	}

	return ret
}

func (tap *testAppProvider) DefaultApps() []appie.AppData {
	return nil
}

func (tap *testAppProvider) CategorizedApps() map[string][]appie.AppData {
	return nil
}
