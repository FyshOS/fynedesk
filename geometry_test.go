package fynedesk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeometry_Center(t *testing.T) {
	g := Geometry{50, 50, 100, 50}
	cx, cy := g.Center()
	assert.Equal(t, 100, cx)
	assert.Equal(t, 75, cy)

	g = Geometry{-50, -50, 100, 50}
	cx, cy = g.Center()
	assert.Equal(t, 0, cx)
	assert.Equal(t, -25, cy)
}

func TestGeometry_Contains(t *testing.T) {
	g := Geometry{50, 50, 50, 50}

	assert.True(t, g.Contains(75, 75))
	assert.False(t, g.Contains(25, 75))
	assert.False(t, g.Contains(75, 175))
}

func TestGeometry_MovedBy(t *testing.T) {
	g := Geometry{50, 50, 50, 50}
	g = g.MovedBy(25, -25)

	assert.Equal(t, 75, g.X)
	assert.Equal(t, 25, g.Y)
}

func TestGeometry_ResizedBy(t *testing.T) {
	g := Geometry{0, 0, 50, 50}
	g = g.ResizedBy(25, -25)

	assert.Equal(t, uint(75), g.Width)
	assert.Equal(t, uint(25), g.Height)
}
