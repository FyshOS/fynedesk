package icon

import (
	"testing"

	"fyne.io/fynedesk/test"

	"github.com/stretchr/testify/assert"
)

func TestLoadAppBundle(t *testing.T) {
	path := "testdata/Test.app"
	app := loadAppBundle("Test", path)

	assert.NotNil(t, app)
	assert.Equal(t, "Test", app.Name())
	assert.NotNil(t, app.Icon("", 0))
}

func TestMacOSAppProvider_FindAppFromName(t *testing.T) {
	provider := NewMacOSAppProvider()
	provider.(*macOSAppProvider).rootDirs = []string{"testdata"}

	app := provider.FindAppFromName("Test")
	assert.NotNil(t, app)
	assert.Equal(t, "Test", app.Name())
}

func TestMacOSAppProvider_FindAppFromWinInfo(t *testing.T) {
	provider := NewMacOSAppProvider()
	provider.(*macOSAppProvider).rootDirs = []string{"testdata"}

	win := test.NewWindow("Test")
	app := provider.FindAppFromWinInfo(win)
	assert.NotNil(t, app)
	assert.Equal(t, "Test", app.Name())
}
