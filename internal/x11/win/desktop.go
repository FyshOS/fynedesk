//go:build linux || openbsd || freebsd || netbsd

package win

// Virtual desktops are stacked vertically in root coordinates: desktop d lies
// (d - current) root heights below the current one, so a window on another
// desktop simply sits outside the visible screens. These helpers map between a
// root-relative y position and the desktop it belongs to.

// deskForY returns the desktop that a window whose vertical centre is at
// centreY root pixels lies on, given the current desktop and the root height.
func deskForY(centreY, rootHeight, current int) int {
	desk := current + floorDiv(centreY, rootHeight)
	if desk < 0 {
		return 0
	}
	return desk
}

// deskTop returns the root-relative y pixel at which desktop desk begins.
func deskTop(desk, rootHeight, current int) int {
	return (desk - current) * rootHeight
}

// floorDiv divides a by b rounding towards negative infinity.
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// floorMod wraps a into the range [0, b).
func floorMod(a, b int) int {
	return a - floorDiv(a, b)*b
}
