package fynedesk

// Geometry represents the position and size of a window or screen
type Geometry struct {
	X, Y          int
	Width, Height uint
}

// NewGeometry creates a new instance of the Geometry struct containing the given values
func NewGeometry(x, y int, w, h uint) Geometry {
	return Geometry{X: x, Y: y, Width: w, Height: h}
}

// Contains will return true if the given x, y coordinate is within the geometry
func (g Geometry) Contains(x, y int) bool {
	return x >= g.X && x < g.X+int(g.Width) &&
		y >= g.Y && y < g.Y+int(g.Height)
}

// Center returns the x, y coordinate at the center of this geometry
func (g Geometry) Center() (int, int) {
	return g.X + int(g.Width/2), g.Y + int(g.Height/2)
}

// MovedBy returns a new geometry that is translated right by x and down by y
func (g Geometry) MovedBy(x, y int) Geometry {
	return Geometry{X: g.X + x, Y: g.Y + y, Width: g.Width, Height: g.Height}
}

// ResizedBy returns a new geometry that is wider by w and taller by h
func (g Geometry) ResizedBy(w, h int) Geometry {
	return Geometry{X: g.X, Y: g.Y, Width: uint(int(g.Width) + w), Height: uint(int(g.Height) + h)}
}
