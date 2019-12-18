// Copyright 2010 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.

package truetype

// A Point is a co-ordinate pair plus whether it is ``on'' a contour or an
// ``off'' control point.
type Point struct {
	X, Y int32
	// The Flags' LSB means whether or not this Point is ``on'' the contour.
	// Other bits are reserved for internal use.
	Flags uint32
}

// A GlyphBuf holds a glyph's contours. A GlyphBuf can be re-used to load a
// series of glyphs from a Font.
type GlyphBuf struct {
	// The glyph's bounding box.
	B Bounds
	// Point contains all Points from all contours of the glyph.
	Point []Point
	// The length of End is the number of contours in the glyph. The i'th
	// contour consists of points Point[End[i-1]:End[i]], where End[-1]
	// is interpreted to mean zero.
	End []int
}

// Flags for decoding a glyph's contours. These flags are documented at
// http://developer.apple.com/fonts/TTRefMan/RM06/Chap6glyf.html.
const (
	flagOnCurve = 1 << iota
	flagXShortVector
	flagYShortVector
	flagRepeat
	flagPositiveXShortVector
	flagPositiveYShortVector
)

// The same flag bits (0x10 and 0x20) are overloaded to have two meanings,
// dependent on the value of the flag{X,Y}ShortVector bits.
const (
	flagThisXIsSame = flagPositiveXShortVector
	flagThisYIsSame = flagPositiveYShortVector
)

// decodeFlags decodes a glyph's run-length encoded flags,
// and returns the remaining data.
func (g *GlyphBuf) decodeFlags(d []byte, offset int, np0 int) (offset1 int) {
	for i := np0; i < len(g.Point); {
		c := uint32(d[offset])
		offset++
		g.Point[i].Flags = c
		i++
		if c&flagRepeat != 0 {
			count := d[offset]
			offset++
			for ; count > 0; count-- {
				g.Point[i].Flags = c
				i++
			}
		}
	}
	return offset
}

// decodeCoords decodes a glyph's delta encoded co-ordinates.
func (g *GlyphBuf) decodeCoords(d []byte, offset int, np0 int) int {
	var x int16
	for i := np0; i < len(g.Point); i++ {
		f := g.Point[i].Flags
		if f&flagXShortVector != 0 {
			dx := int16(d[offset])
			offset++
			if f&flagPositiveXShortVector == 0 {
				x -= dx
			} else {
				x += dx
			}
		} else if f&flagThisXIsSame == 0 {
			x += int16(u16(d, offset))
			offset += 2
		}
		g.Point[i].X = int32(x)
	}
	var y int16
	for i := np0; i < len(g.Point); i++ {
		f := g.Point[i].Flags
		if f&flagYShortVector != 0 {
			dy := int16(d[offset])
			offset++
			if f&flagPositiveYShortVector == 0 {
				y -= dy
			} else {
				y += dy
			}
		} else if f&flagThisYIsSame == 0 {
			y += int16(u16(d, offset))
			offset += 2
		}
		g.Point[i].Y = int32(y)
	}
	return offset
}

// Load loads a glyph's contours from a Font, overwriting any previously
// loaded contours for this GlyphBuf. The Hinter is optional; if non-nil, then
// the resulting glyph will be hinted by the Font's bytecode instructions.
func (g *GlyphBuf) Load(f *Font, scale int32, i Index, h *Hinter) error {
	// Reset the GlyphBuf.
	g.B = Bounds{}
	g.Point = g.Point[:0]
	g.End = g.End[:0]
	if err := g.load(f, i, 0); err != nil {
		return err
	}
	g.B.XMin = f.scale(scale * g.B.XMin)
	g.B.YMin = f.scale(scale * g.B.YMin)
	g.B.XMax = f.scale(scale * g.B.XMax)
	g.B.YMax = f.scale(scale * g.B.YMax)
	for i := range g.Point {
		g.Point[i].X = f.scale(scale * g.Point[i].X)
		g.Point[i].Y = f.scale(scale * g.Point[i].Y)
	}
	if h != nil {
		// TODO: invoke h.
	}
	return nil
}

// loadCompound loads a glyph that is composed of other glyphs.
func (g *GlyphBuf) loadCompound(f *Font, glyf []byte, offset, recursion int) error {
	// Flags for decoding a compound glyph. These flags are documented at
	// http://developer.apple.com/fonts/TTRefMan/RM06/Chap6glyf.html.
	const (
		flagArg1And2AreWords = 1 << iota
		flagArgsAreXYValues
		flagRoundXYToGrid
		flagWeHaveAScale
		flagUnused
		flagMoreComponents
		flagWeHaveAnXAndYScale
		flagWeHaveATwoByTwo
		flagWeHaveInstructions
		flagUseMyMetrics
		flagOverlapCompound
	)
	for {
		flags := u16(glyf, offset)
		component := u16(glyf, offset+2)
		var dx, dy int16
		if flags&flagArg1And2AreWords != 0 {
			dx = int16(u16(glyf, offset+4))
			dy = int16(u16(glyf, offset+6))
			offset += 8
		} else {
			dx = int16(int8(glyf[offset+4]))
			dy = int16(int8(glyf[offset+5]))
			offset += 6
		}
		if flags&flagArgsAreXYValues == 0 {
			return UnsupportedError("compound glyph transform vector")
		}
		if flags&(flagWeHaveAScale|flagWeHaveAnXAndYScale|flagWeHaveATwoByTwo) != 0 {
			return UnsupportedError("compound glyph scale/transform")
		}
		b0, i0 := g.B, len(g.Point)
		g.load(f, Index(component), recursion+1)
		for i := i0; i < len(g.Point); i++ {
			g.Point[i].X += int32(dx)
			g.Point[i].Y += int32(dy)
		}
		if flags&flagUseMyMetrics == 0 {
			g.B = b0
		}
		if flags&flagMoreComponents == 0 {
			break
		}
	}
	return nil
}

// load appends a glyph's contours to this GlyphBuf.
func (g *GlyphBuf) load(f *Font, i Index, recursion int) error {
	if recursion >= 4 {
		return UnsupportedError("excessive compound glyph recursion")
	}
	// Find the relevant slice of f.glyf.
	var g0, g1 uint32
	if f.locaOffsetFormat == locaOffsetFormatShort {
		g0 = 2 * uint32(u16(f.loca, 2*int(i)))
		g1 = 2 * uint32(u16(f.loca, 2*int(i)+2))
	} else {
		g0 = u32(f.loca, 4*int(i))
		g1 = u32(f.loca, 4*int(i)+4)
	}
	if g0 == g1 {
		return nil
	}
	glyf := f.glyf[g0:g1]
	// Decode the contour end indices.
	ne := int(int16(u16(glyf, 0)))
	g.B.XMin = int32(int16(u16(glyf, 2)))
	g.B.YMin = int32(int16(u16(glyf, 4)))
	g.B.XMax = int32(int16(u16(glyf, 6)))
	g.B.YMax = int32(int16(u16(glyf, 8)))
	offset := 10
	if ne == -1 {
		return g.loadCompound(f, glyf, offset, recursion)
	} else if ne < 0 {
		// http://developer.apple.com/fonts/TTRefMan/RM06/Chap6glyf.html says that
		// "the values -2, -3, and so forth, are reserved for future use."
		return UnsupportedError("negative number of contours")
	}
	ne0, np0 := len(g.End), len(g.Point)
	ne += ne0
	if ne <= cap(g.End) {
		g.End = g.End[:ne]
	} else {
		g.End = make([]int, ne, ne*2)
	}
	for i := ne0; i < ne; i++ {
		g.End[i] = 1 + np0 + int(u16(glyf, offset))
		offset += 2
	}
	// Skip the TrueType hinting instructions.
	instrLen := int(u16(glyf, offset))
	offset += 2 + instrLen
	// Decode the points.
	np := int(g.End[ne-1])
	if np <= cap(g.Point) {
		g.Point = g.Point[:np]
	} else {
		g.Point = make([]Point, np, np*2)
	}
	offset = g.decodeFlags(glyf, offset, np0)
	g.decodeCoords(glyf, offset, np0)
	return nil
}

// NewGlyphBuf returns a newly allocated GlyphBuf.
func NewGlyphBuf() *GlyphBuf {
	g := new(GlyphBuf)
	g.Point = make([]Point, 0, 256)
	g.End = make([]int, 0, 32)
	return g
}
