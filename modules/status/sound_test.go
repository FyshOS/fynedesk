package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartsWith(t *testing.T) {
	assert.True(t, startsWith("vol", "volume"))
	assert.True(t, startsWith("volume", "volume"))
	assert.True(t, startsWith("volume up", "volume"))

	assert.False(t, startsWith("", "volume"))
	assert.False(t, startsWith("voo", "volume"))
}
