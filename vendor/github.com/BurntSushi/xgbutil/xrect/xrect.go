package xrect

import "fmt"

// Define a base and simple Rect interface.
type Rect interface {
	X() int
	Y() int
	Width() int
	Height() int
	XSet(x int)
	YSet(y int)
	WidthSet(width int)
	HeightSet(height int)
	Pieces() (int, int, int, int)
}

// RectPieces just returns a four-tuple of x, y, width and height
func RectPieces(xr Rect) (int, int, int, int) {
	return xr.X(), xr.Y(), xr.Width(), xr.Height()
}
func Pieces(xr Rect) (int, int, int, int) {
	return RectPieces(xr)
}

// Provide a simple implementation of a rect.
// Maybe this will be all we need?
type XRect struct {
	x, y          int
	width, height int
}

// Provide the ability to construct an XRect.
func New(x, y, w, h int) *XRect {
	return &XRect{x, y, w, h}
}

func (r *XRect) String() string {
	return fmt.Sprintf("[(%d, %d) %dx%d]", r.x, r.y, r.width, r.height)
}

// Satisfy the Rect interface
func (r *XRect) X() int {
	return r.x
}

func (r *XRect) Y() int {
	return r.y
}

func (r *XRect) Width() int {
	return r.width
}

func (r *XRect) Height() int {
	return r.height
}

func (r *XRect) XSet(x int) {
	r.x = x
}

func (r *XRect) YSet(y int) {
	r.y = y
}

func (r *XRect) WidthSet(width int) {
	r.width = width
}

func (r *XRect) HeightSet(height int) {
	r.height = height
}

// Pieces just returns a four-tuple of x, y, width and height
func (r *XRect) Pieces() (int, int, int, int) {
	return r.X(), r.Y(), r.Width(), r.Height()
}

// Valid returns whether a rectangle is valid or not. i.e., a width AND height
// not equal to zero.
func Valid(r Rect) bool {
	return r.Width() != 0 && r.Height() != 0
}

// Subtract subtracts r2 from r1 and returns the result as a
// new slice of Rects.
// Basically, rectangle subtraction works by cutting r2 out of r1, and returning
// the resulting rectangles.
// If r1 does not overlap r2, then only one rectangle is returned and is
// equivalent to r1.
// If r2 covers r1, then no rectangles are returned.
// If r1 covers r2, then four rectangles are returned.
// If r2 partially overlaps r1, then one, two or three rectangles are returned.
func Subtract(r1 Rect, r2 Rect) []Rect {
	r1x1, r1y1, r1w, r1h := r1.Pieces()
	r2x1, r2y1, r2w, r2h := r2.Pieces()

	r1x2, r1y2 := r1x1+r1w, r1y1+r1h
	r2x2, r2y2 := r2x1+r2w, r2y1+r2h

	// No intersection; return r1.
	if r2x1 >= r1x2 || r1x1 >= r2x2 || r2y1 >= r1y2 || r1y1 >= r2y2 {
		return []Rect{New(r1x1, r1y1, r1w, r1h)}
	}

	// r2 covers r1; so subtraction yields no rectangles.
	if r1x1 >= r2x1 && r1y1 >= r2y1 && r1x2 <= r2x2 && r1y2 <= r2y2 {
		return []Rect{}
	}

	// Now generate each of the four possible rectangles and add them only
	// if they are valid (i.e., width/height >= 1)
	result := make([]Rect, 0, 4)

	rect1 := New(r1x1, r1y1, r1w, r2y1-r1y1)
	rect2 := New(r1x1, r1y1, r2x1-r1x1, r1h)
	rect3 := New(r1x1, r2y2, r1w, r1h-((r2y1-r1y1)+r2h))
	rect4 := New(r2x2, r1y1, r1w-((r2x1-r1x1)+r2w), r1h)

	if Valid(rect1) {
		result = append(result, rect1)
	}
	if Valid(rect2) {
		result = append(result, rect2)
	}
	if Valid(rect3) {
		result = append(result, rect3)
	}
	if Valid(rect4) {
		result = append(result, rect4)
	}

	return result
}

// IntersectArea takes two rectangles satisfying the Rect interface and
// returns the area of their intersection. If there is no intersection, return
// 0 area.
func IntersectArea(r1 Rect, r2 Rect) int {
	x1, y1, w1, h1 := RectPieces(r1)
	x2, y2, w2, h2 := RectPieces(r2)
	if x2 < x1+w1 && x2+w2 > x1 && y2 < y1+h1 && y2+h2 > y1 {
		iw := min(x1+w1-1, x2+w2-1) - max(x1, x2) + 1
		ih := min(y1+h1-1, y2+h2-1) - max(y1, y2) + 1
		return iw * ih
	}

	return 0
}

// LargestOverlap returns the index of the rectangle in 'haystack' that has the
// largest overlap with the rectangle 'needle'.
// This is commonly used to find which monitor a window should belong on.
// (Since it can technically be partially displayed on more than one monitor
// at a time.)
// Be careful, the return value can be -1 if there is no overlap.
func LargestOverlap(needle Rect, haystack []Rect) int {
	biggestArea := 0
	reti := -1

	var area int
	for i, possible := range haystack {
		area = IntersectArea(needle, possible)
		if area > biggestArea {
			biggestArea = area
			reti = i
		}
	}
	return reti
}

// ApplyStrut takes a list of Rects (typically the rectangles that represent
// each physical head in this case), the root window geometry,
// and a set of parameters representing a
// strut, and modifies the list of Rects to account for that strut.
// That is, it shrinks each rect.
// Note that if struts overlap, the *most restrictive* one is used. This seems
// like the most sensible response to a weird scenario.
// (If you don't have a partial strut, just use '0' for the extra fields.)
// See xgbutil/examples/workarea-struts for an example of how to use this to
// get accurate workarea for each physical head.
func ApplyStrut(rects []Rect, rootWidth, rootHeight uint,
	left, right, top, bottom,
	left_start_y, left_end_y, right_start_y, right_end_y,
	top_start_x, top_end_x, bottom_start_x, bottom_end_x uint) {

	var nx, ny uint // 'n*' are new values that may or may not be used
	var nw, nh uint
	var x_, y_, w_, h_ int
	var x, y, w, h uint
	var bt, tp, lt, rt bool
	rWidth, rHeight := rootWidth, rootHeight

	// The essential idea of struts, and particularly partial struts, is that
	// one piece of a border of the screen can be "reserved" for some
	// special windows like docks, panels, taskbars and system trays.
	// Since we assume that one window can only reserve one piece of a border
	// (either top, left, right or bottom), we iterate through each rect
	// in our list and check if that rect is affected by the given strut.
	// If it is, we modify the current rect appropriately.
	// TODO: Fix this so old school _NET_WM_STRUT can work too. It actually
	// should be pretty simple: change conditions like 'if tp' to
	// 'if tp || (top_start_x == 0 && top_end_x == 0 && top != 0)'.
	// Thus, we would end up changing every rect, which is what old school
	// struts should do. We may also make a conscious choice to ignore them
	// when 'rects' has more than one rect, since the old school struts will
	// typically result in undesirable behavior.
	for _, rect := range rects {
		x_, y_, w_, h_ = RectPieces(rect)
		x, y, w, h = uint(x_), uint(y_), uint(w_), uint(h_)

		bt = bottom_start_x != bottom_end_x &&
			(xInRect(bottom_start_x, rect) || xInRect(bottom_end_x, rect))
		tp = top_start_x != top_end_x &&
			(xInRect(top_start_x, rect) || xInRect(top_end_x, rect))
		lt = left_start_y != left_end_y &&
			(yInRect(left_start_y, rect) || yInRect(left_end_y, rect))
		rt = right_start_y != right_end_y &&
			(yInRect(right_start_y, rect) || yInRect(right_end_y, rect))

		if bt {
			nh = h - (bottom - ((rHeight - h) - y))
			if nh < uint(rect.Height()) {
				rect.HeightSet(int(nh))
			}
		} else if tp {
			nh = h - (top - y)
			if nh < uint(rect.Height()) {
				rect.HeightSet(int(nh))
			}

			ny = top
			if ny > uint(rect.Y()) {
				rect.YSet(int(ny))
			}
		} else if rt {
			nw = w - (right - ((rWidth - w) - x))
			if nw < uint(rect.Width()) {
				rect.WidthSet(int(nw))
			}
		} else if lt {
			nw = w - (left - x)
			if nw < uint(rect.Width()) {
				rect.WidthSet(int(nw))
			}

			nx = left
			if nx > uint(rect.X()) {
				rect.XSet(int(nx))
			}
		}
	}
}

// xInRect is whether a particular x-coordinate is vertically constrained by
// a rectangle.
func xInRect(xtest uint, rect Rect) bool {
	x, _, w, _ := RectPieces(rect)
	return int(xtest) >= x && int(xtest) < (x+w)
}

// yInRect is whether a particular y-coordinate is horizontally constrained by
// a rectangle.
func yInRect(ytest uint, rect Rect) bool {
	_, y, _, h := RectPieces(rect)
	return int(ytest) >= y && int(ytest) < (y+h)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
