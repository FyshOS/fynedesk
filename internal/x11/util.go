package x11

import (
	"fyne.io/fynedesk"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xrect"
)

// GeometryFromRect creates a new geometry struct that has the same values as the given xrect.Rect
func GeometryFromRect(rect xrect.Rect) fynedesk.Geometry {
	return fynedesk.Geometry{X: rect.X(), Y: rect.Y(), Width: uint(rect.Width()), Height: uint(rect.Height())}
}

// GeometryFromGetGeometryReply returns a new geometry struct that has the same values as the
// given xproto.GetGeometryReply
func GeometryFromGetGeometryReply(g *xproto.GetGeometryReply) fynedesk.Geometry {
	return fynedesk.Geometry{X: int(g.X), Y: int(g.Y), Width: uint(g.Width), Height: uint(g.Height)}
}

// GeometryToUint32s returns this geometry in the form of a uint32 slice
func GeometryToUint32s(g fynedesk.Geometry) []uint32 {
	return []uint32{uint32(g.X), uint32(g.Y), uint32(g.Width), uint32(g.Height)}
}
