package win

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstrainToRect(t *testing.T) {
	// already inside: unchanged
	x, y := constrainToRect(100, 50, 200, 100, 36, 0, 1000, 700)
	assert.Equal(t, 100, x)
	assert.Equal(t, 50, y)

	// at the origin, under the bar: pushed to the content edge
	x, y = constrainToRect(0, 0, 200, 100, 36, 0, 1000, 700)
	assert.Equal(t, 36, x)
	assert.Equal(t, 0, y)

	// hanging off the bottom right: pulled back in
	x, y = constrainToRect(900, 650, 200, 100, 36, 0, 1000, 700)
	assert.Equal(t, 836, x)
	assert.Equal(t, 600, y)

	// too large to fit: top-left stays inside
	x, y = constrainToRect(-10, -10, 2000, 100, 36, 0, 1000, 700)
	assert.Equal(t, 36, x)
	assert.Equal(t, 0, y)

	// a screen to the right of the primary
	x, y = constrainToRect(1900, 0, 200, 100, 1920, 0, 1920, 1080)
	assert.Equal(t, 1920, x)
	assert.Equal(t, 0, y)
}
