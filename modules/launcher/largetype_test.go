package launcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLargeType_LaunchSuggestions_NoMatch(t *testing.T) {
	l := newLargeType().(*largeType)

	assert.Nil(t, l.LaunchSuggestions("hello world"))
	assert.Nil(t, l.LaunchSuggestions("xyz"))
}

func TestLargeType_LaunchSuggestions_AliasPrefix(t *testing.T) {
	l := newLargeType().(*largeType)

	for _, alias := range largeTypeAliases {
		res := l.LaunchSuggestions(alias)
		if assert.Len(t, res, 1) {
			item := res[0].(*largeTypeItem)
			assert.Equal(t, "", item.text)
		}
	}

	// partial prefix still matches
	res := l.LaunchSuggestions("lar")
	if assert.Len(t, res, 1) {
		assert.Equal(t, "", res[0].(*largeTypeItem).text)
	}
}

func TestLargeType_LaunchSuggestions_WithText(t *testing.T) {
	l := newLargeType().(*largeType)

	res := l.LaunchSuggestions("largetype Hello World")
	if assert.Len(t, res, 1) {
		item := res[0].(*largeTypeItem)
		assert.Equal(t, "Hello World", item.text)
		assert.Equal(t, "Large Type: Hello World", item.Title())
	}

	res = l.LaunchSuggestions("BIG Fyne")
	if assert.Len(t, res, 1) {
		assert.Equal(t, "Fyne", res[0].(*largeTypeItem).text)
	}
}

// The command and the launcher entry are built from the same item, so what one
// accepts the other does.
func TestLargeType_Command(t *testing.T) {
	l := newLargeType().(*largeType)

	cmds := l.Commands()
	if !assert.Len(t, cmds, 1) {
		return
	}
	assert.Equal(t, largeTypeAliases[0], cmds[0].Name)

	item := l.item("Hello World")
	if assert.NotNil(t, item) {
		assert.Equal(t, "Hello World", item.(*largeTypeItem).text)
	}

	// Nothing to display is refused rather than flashing an empty screen.
	assert.Nil(t, l.item("  "))
	_, err := cmds[0].Run("")
	assert.Error(t, err)
}
