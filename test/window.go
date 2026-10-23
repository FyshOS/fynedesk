package test

import (
	"fyshos.com/tyde"

	"fyne.io/fyne/v2"
)

// Window is an in-memory virtual window for test purposes
type Window struct {
	props dummyProperties

	iconic, focused, fullscreen, maximized, raised, pinned bool

	parent        tyde.Window
	x, y, desk    int
	width, height uint
}

// NewWindow creates a virtual window with the given title ("" is acceptable)
func NewWindow(title string) *Window {
	win := &Window{width: 10, height: 10}
	win.props.name = title
	return win
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

// Unfocus is a test utility to take input focus away from this window
func (w *Window) Unfocus() {
	w.focused = false
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

// MarkDestroyed would mark the window as gone, we do nothing for tests
func (w *Window) MarkDestroyed() {
}

// Reframe would rebuild the frame X11 window, we do nothing for tests
func (w *Window) Reframe() {
}

// Maximize simulates this window becoming maximized
func (w *Window) Maximize() {
	w.maximized = true
}

// Maximized returns true if this window has been made maximized
func (w *Window) Maximized() bool {
	return w.maximized
}

// Move records the new position of the window
func (w *Window) Move(pos fyne.Position) {
	w.x, w.y = int(pos.X), int(pos.Y)
}

// Parent returns a window that this should be positioned within, if set.
func (w *Window) Parent() tyde.Window {
	return w.parent
}

// Pin requests that the window be visible on all desktops
func (w *Window) Pin() {
	w.pinned = true
}

// Pinned returns true if the window should be visible on all desktops
func (w *Window) Pinned() bool {
	return w.pinned
}

// Position returns the current position.
func (w *Window) Position() fyne.Position {
	return fyne.NewPos(float32(w.x), float32(w.y))
}

// Resize records the new size of the window
func (w *Window) Resize(size fyne.Size) {
	w.width, w.height = uint(size.Width), uint(size.Height)
}

// Size returns the current size.
func (w *Window) Size() fyne.Size {
	return fyne.NewSize(float32(w.width), float32(w.height))
}

// Properties obtains the window properties currently set
func (w *Window) Properties() tyde.WindowProperties {
	return w.props
}

// RaiseAbove sets this window to be above the passed window
func (w *Window) RaiseAbove(tyde.Window) {
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
func (w *Window) SetParent(p tyde.Window) {
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

// Unpin resets the state of being visible on all windows.
// The window will return to being visible on its specified desktop.
func (w *Window) Unpin() {
	w.pinned = false
}
