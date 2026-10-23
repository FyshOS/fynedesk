package emoji

import (
	"testing"

	"fyne.io/fyne/v2"

	"fyshos.com/tyde"

	"github.com/stretchr/testify/assert"
)

func newTestModule(t *testing.T) *module {
	t.Helper()

	m := &module{}
	m.picker = &picker{}

	fyne.CurrentApp().Preferences().SetString(prefRecent, "")
	t.Cleanup(func() { fyne.CurrentApp().Preferences().SetString(prefRecent, "") })
	return m
}

func TestModule_Metadata(t *testing.T) {
	m := newTestModule(t)
	assert.Equal(t, ModuleName, m.Metadata().Name)
}

func TestModule_Shortcut(t *testing.T) {
	m := newTestModule(t)

	shortcuts := m.Shortcuts()
	if assert.Len(t, shortcuts, 1) {
		for s := range shortcuts {
			assert.Equal(t, fyne.KeyPeriod, s.KeyName)
			assert.Equal(t, "Show Emoji Picker", s.Name)
		}
	}
}

func TestModule_StatusAreaWidget(t *testing.T) {
	m := newTestModule(t)
	assert.NotNil(t, m.StatusAreaWidget())
}

func TestModule_DestroyHidesPicker(t *testing.T) {
	tyde.SetInstance(newTestDesktop())
	t.Cleanup(func() { tyde.SetInstance(nil) })

	m := newTestModule(t)
	m.ShowPicker()

	m.Destroy()
	assert.False(t, m.picker.shown)
}

func TestModule_LaunchSuggestions_Alias(t *testing.T) {
	m := newTestModule(t)

	for _, input := range []string{"e", "em", "emoji", "SMIL", "symbol"} {
		items := m.LaunchSuggestions(input)
		if assert.Len(t, items, 1, "input %q", input) {
			assert.Equal(t, "Emoji Picker", items[0].Title())
		}
	}
}

func TestModule_LaunchSuggestions_NoMatch(t *testing.T) {
	m := newTestModule(t)

	assert.Nil(t, m.LaunchSuggestions(""))
	assert.Nil(t, m.LaunchSuggestions("   "))
	assert.Nil(t, m.LaunchSuggestions("firefox"))
}

func TestModule_LaunchSuggestions_Search(t *testing.T) {
	m := newTestModule(t)

	items := m.LaunchSuggestions("emoji rocket")
	if assert.NotEmpty(t, items) {
		assert.Contains(t, items[0].Title(), "🚀")
		assert.LessOrEqual(t, len(items), 5, "the launcher should not be flooded")
	}
}

func TestModule_LaunchSuggestions_SearchWithNoMatchOffersPicker(t *testing.T) {
	m := newTestModule(t)

	items := m.LaunchSuggestions("emoji zzzzznotanemoji")
	if assert.Len(t, items, 1) {
		assert.Equal(t, "Emoji Picker", items[0].Title())
	}
}

func TestModule_LaunchSuggestion_LaunchCopies(t *testing.T) {
	m := newTestModule(t)

	items := m.LaunchSuggestions("emoji rocket")
	items[0].Launch()

	assert.Equal(t, "🚀", fyne.CurrentApp().Clipboard().Content())
}

func TestModule_LaunchSuggestion_OpensPicker(t *testing.T) {
	tyde.SetInstance(newTestDesktop())
	t.Cleanup(func() { tyde.SetInstance(nil) })

	m := newTestModule(t)
	m.LaunchSuggestions("emoji")[0].Launch()

	assert.True(t, m.picker.shown)
}

func TestModule_ShowPicker(t *testing.T) {
	tyde.SetInstance(newTestDesktop())
	t.Cleanup(func() { tyde.SetInstance(nil) })

	m := newTestModule(t)
	m.ShowPicker()

	assert.True(t, m.picker.shown)
}

// The command copies the same emoji the launcher would have, because it runs
// the very suggestion the launcher builds.
func TestModule_Command_Copies(t *testing.T) {
	m := newTestModule(t)
	fyne.CurrentApp().Clipboard().SetContent("")

	cmds := m.Commands()
	if !assert.Len(t, cmds, 1) {
		return
	}
	assert.Equal(t, launchAliases[0], cmds[0].Name)

	reply, err := cmds[0].Run("rocket")
	assert.NoError(t, err)
	assert.Contains(t, reply, "🚀")
	assert.Equal(t, "🚀", fyne.CurrentApp().Clipboard().Content())
}
