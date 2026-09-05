package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The system entry is a preview with no theme.json beside it, so it has to
// write a theme Fyne can parse rather than copy a file that is not there.
func TestWriteTheme_System(t *testing.T) {
	config := t.TempDir()
	require.NoError(t, writeTheme(themeNameSystem, config, t.TempDir()))

	data, err := os.ReadFile(filepath.Join(config, "theme.json"))
	require.NoError(t, err)
	assert.Equal(t, "{}", string(data))

	var parsed map[string]any
	assert.NoError(t, json.Unmarshal(data, &parsed), "Fyne must be able to parse the result")
}

func TestWriteTheme_Bundled(t *testing.T) {
	config := t.TempDir()
	require.NoError(t, writeTheme(themeNameDefault, config, t.TempDir()))

	written, err := os.ReadFile(filepath.Join(config, "theme.json"))
	require.NoError(t, err)

	bundled, err := bundledThemes.ReadFile("themes/fyshos/theme.json")
	require.NoError(t, err)
	assert.Equal(t, string(bundled), string(written))
}

func TestWriteTheme_Custom(t *testing.T) {
	config, custom := t.TempDir(), t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(custom, "mine"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(custom, "mine", "theme.json"),
		[]byte(`{"Colors":{"primary":"#ff0000"}}`), 0o644))

	require.NoError(t, writeTheme("mine", config, custom))
	data, err := os.ReadFile(filepath.Join(config, "theme.json"))
	require.NoError(t, err)
	assert.Equal(t, `{"Colors":{"primary":"#ff0000"}}`, string(data))
}

// A theme that cannot be found must leave the one in use alone - truncating it
// would drop the desktop back to unthemed on the next start.
func TestWriteTheme_MissingKeepsCurrent(t *testing.T) {
	config := t.TempDir()
	current := filepath.Join(config, "theme.json")
	require.NoError(t, os.WriteFile(current, []byte(`{"Colors":{"primary":"#00ff00"}}`), 0o644))

	assert.Error(t, writeTheme("nope", config, t.TempDir()))

	data, err := os.ReadFile(current)
	require.NoError(t, err)
	assert.Equal(t, `{"Colors":{"primary":"#00ff00"}}`, string(data))
}

// Every bundled theme needs an entry of its own: the rename left "fyshos" to be
// title cased into "Fyshos", and a missing one would read as "Description...".
func TestThemeListEntry_Bundled(t *testing.T) {
	dirs, err := bundledThemes.ReadDir("themes")
	require.NoError(t, err)
	require.NotEmpty(t, dirs)

	for _, dir := range dirs {
		info := themeListEntry(dir.Name())
		assert.Contains(t, bundledThemeInfo, dir.Name(), "bundled theme is undescribed")
		assert.NotEmpty(t, info.title)
		assert.NotEmpty(t, info.description)
	}

	assert.Equal(t, "FyshOS", themeListEntry(themeNameDefault).title)
	assert.Equal(t, "System", themeListEntry(themeNameSystem).title)
}

func TestThemeListEntry_Custom(t *testing.T) {
	info := themeListEntry("mine")

	assert.Equal(t, "Mine", info.title)
	assert.NotEmpty(t, info.description)
}

func TestThemeMarkdown(t *testing.T) {
	assert.Equal(t, "## Neon\n\nFunky orange, blues and purples",
		themeMarkdown(themeListEntry("neon")))
}
