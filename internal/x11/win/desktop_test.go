//go:build linux || openbsd || freebsd || netbsd

package win

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeskForY(t *testing.T) {
	const h = 1080
	assert.Equal(t, 0, deskForY(540, h, 0))
	assert.Equal(t, 0, deskForY(1079, h, 0))
	assert.Equal(t, 1, deskForY(1080, h, 0))
	assert.Equal(t, 2, deskForY(2160+540, h, 0))
	assert.Equal(t, 3, deskForY(540, h, 3), "centred on screen is the current desktop")
	assert.Equal(t, 1, deskForY(-540, h, 2), "one root height above the current desktop")
	assert.Equal(t, 0, deskForY(-540, h, 0), "never below the first desktop")
}

func TestDeskTop(t *testing.T) {
	const h = 1080
	assert.Equal(t, 0, deskTop(0, h, 0))
	assert.Equal(t, 2160, deskTop(2, h, 0))
	assert.Equal(t, -1080, deskTop(1, h, 2))
}

func TestFloorDivMod(t *testing.T) {
	assert.Equal(t, 0, floorDiv(5, 10))
	assert.Equal(t, -1, floorDiv(-5, 10))
	assert.Equal(t, -1, floorDiv(-10, 10))
	assert.Equal(t, -2, floorDiv(-11, 10))
	assert.Equal(t, 5, floorMod(5, 10))
	assert.Equal(t, 5, floorMod(-5, 10))
	assert.Equal(t, 0, floorMod(-10, 10))
	assert.Equal(t, 9, floorMod(-11, 10))
}
