package wm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwapPixel(t *testing.T) {
	in := []byte{0xff, 0x99, 0x33, 0x00}
	out := []uint8{0xff, 0xff, 0xff, 0xff}
	swapPixels(in, out, 0)

	assert.Equal(t, uint8(0x33), out[0])
	assert.Equal(t, uint8(0x99), out[1])
	assert.Equal(t, uint8(0xff), out[2])
	assert.Equal(t, uint8(0xff), out[3])
}
