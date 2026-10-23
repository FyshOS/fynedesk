package ai

import (
	"testing"

	"fyne.io/fyne/v2"

	"github.com/stretchr/testify/assert"

	"fyshos.com/tyde"
)

func TestAssistant_Shortcuts(t *testing.T) {
	a := &assistant{}
	var bind tyde.KeyBindModule = a

	shorts := bind.Shortcuts()
	assert.Equal(t, 1, len(shorts))

	for s, fn := range shorts {
		assert.Equal(t, "Show AI Assistant", s.ShortcutName())
		assert.NotNil(t, fn)

		// the assistant key sends Shift+Super+F23
		assert.Equal(t, tyde.KeyAssistant, s.Key())
		assert.Equal(t, fyne.KeyModifierShift|fyne.KeyModifierSuper, s.Mod())
	}
}
