package test

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fynedesk"
)

// Window is an in-memory virtual window for test purposes
type Window struct {
	props dummyProperties

	iconic, focused, fullscreen, maximized, raised bool

	parent        fynedesk.Window
	x, y, desk    int
	width, height uint
}

// NewWindow creates a virtual window with the given title ("" is acceptable)
func NewWindow(title string) *Window {
	win := &Window{width: 10, height: 10}
	win.props.name = title
	return win
}

// Capture the contents of the window. Our test code cowardly refuses to do this.
func (w *Window) Capture() image.Image {
	return nil // we can add this if required for testing
}

// Close this test window
func (w *Window) Close() {
	// no-op
}

// Desktop returns the id of the desktop this window is associated with
func (w *Window) Desktop() int {
	return w.desk
}

// Fullscreened returns true if this window has been made full screen
func (w *Window) Fullscreened() bool {
	return w.fullscreen
}

// Focused returns true if this window has requested focus
func (w *Window) Focused() bool {
	return w.focused
}

// Focus sets this window to have input focus
func (w *Window) Focus() {
	w.focused = true
}

// Fullscreen simulates this window becoming full screen
func (w *Window) Fullscreen() {
	w.fullscreen = true
}

// Iconic returns true if this window has been iconified
func (w *Window) Iconic() bool {
	return w.iconic
}

// Iconify sets this window to be reduced to an icon
func (w *Window) Iconify() {
	w.iconic = true
}

// Maximize simulates this window becoming maximized
func (w *Window) Maximize() {
	w.maximized = true
}

// Maximized returns true if this window has been made maximized
func (w *Window) Maximized() bool {
	return false
}

// Parent returns a window that this should be positioned within, if set.
func (w *Window) Parent() fynedesk.Window {
	return w.parent
}

// Position returns 0, 0 for test windows
func (w *Window) Position() fyne.Position {
	return fyne.NewPos(0, 0)
}

// Properties obtains the window properties currently set
func (w *Window) Properties() fynedesk.WindowProperties {
	return w.props
}

// RaiseAbove sets this window to be above the passed window
func (w *Window) RaiseAbove(fynedesk.Window) {
	// no-op (this is instructing the window after stack changes)
}

// RaiseToTop sets this window to be the topmost
func (w *Window) RaiseToTop() {
	w.raised = true
}

// SetClass is a test utility to set the class property of this window
func (w *Window) SetClass(class []string) {
	w.props.class = class
}

// SetCommand is a test utility to set the command property of this window
func (w *Window) SetCommand(cmd string) {
	w.props.cmd = cmd
}

// SetDesktop sets the index of a desktop this window would associate with
func (w *Window) SetDesktop(id int) {
	w.desk = id
}

// SetIconName is a test utility to set the icon-name property of this window
func (w *Window) SetIconName(name string) {
	w.props.iconName = name
}

// SetGeometry is a test utility to set the position and size of this window
func (w *Window) SetGeometry(x, y int, width, height uint) {
	w.x, w.y = x, y
	w.width, w.height = width, height
}

// SetParent is a test utility to set a parent of this window
func (w *Window) SetParent(p fynedesk.Window) {
	w.parent = p
}

// TopWindow returns true if this window has been raised above all others
func (w *Window) TopWindow() bool {
	return w.raised
}

// Unfullscreen removes the fullscreen state of this window
func (w *Window) Unfullscreen() {
	w.fullscreen = false
}

// Uniconify returns this window to its normal state
func (w *Window) Uniconify() {
	w.iconic = false
}

// Unmaximize removes the maximized state of this window
func (w *Window) Unmaximize() {
	w.maximized = false
}
