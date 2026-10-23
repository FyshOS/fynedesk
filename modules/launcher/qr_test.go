package launcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The command and the launcher entry encode the same content, and neither is
// asked to draw a code for nothing.
func TestQR_Command(t *testing.T) {
	q := newQR().(*qr)

	cmds := q.Commands()
	if !assert.Len(t, cmds, 1) {
		return
	}
	assert.Equal(t, qrAliases[0], cmds[0].Name)

	item := q.item("https://fyshos.com")
	if assert.NotNil(t, item) {
		assert.Equal(t, "https://fyshos.com", item.(*qrItem).content)
	}

	assert.Nil(t, q.item(" "))
	_, err := cmds[0].Run("")
	assert.Error(t, err)
}
