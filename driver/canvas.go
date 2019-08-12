package driver

import (
	"image"
	"image/draw"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
)

// WindowlessCanvas represents a canvas with no window to manipulate it
type WindowlessCanvas interface {
	fyne.Canvas
	Resize(fyne.Size)
}

type softwareCanvas struct {
	size fyne.Size

	content, overlay fyne.CanvasObject
	focused          fyne.Focusable

	onTypedRune func(rune)
	onTypedKey  func(*fyne.KeyEvent)

	fyne.ShortcutHandler
}

func (c *softwareCanvas) Content() fyne.CanvasObject {
	return c.content
}

func (c *softwareCanvas) SetContent(content fyne.CanvasObject) {
	c.content = content

	if content == nil {
		return
	}

	theme := fyne.CurrentApp().Settings().Theme()
	padding := fyne.NewSize(theme.Padding(), theme.Padding())
	c.Resize(content.MinSize().Add(padding))
}

func (c *softwareCanvas) Overlay() fyne.CanvasObject {
	return c.overlay
}

func (c *softwareCanvas) SetOverlay(overlay fyne.CanvasObject) {
	c.overlay = overlay
	if overlay != nil {
		overlay.Resize(c.Size())
	}
}

func (c *softwareCanvas) Refresh(fyne.CanvasObject) {
}

func (c *softwareCanvas) Focus(obj fyne.Focusable) {
	if obj == c.focused {
		return
	}

	if c.focused != nil {
		c.focused.FocusLost()
	}

	c.focused = obj

	if obj != nil {
		obj.FocusGained()
	}
}

func (c *softwareCanvas) Unfocus() {
	if c.focused != nil {
		c.focused.FocusLost()
	}
	c.focused = nil
}

func (c *softwareCanvas) Focused() fyne.Focusable {
	return c.focused
}

func (c *softwareCanvas) Size() fyne.Size {
	return c.size
}

func (c *softwareCanvas) Resize(size fyne.Size) {
	c.size = size
}

func (c *softwareCanvas) Scale() float32 {
	return 1.0
}

func (c *softwareCanvas) SetScale(float32) {
}

func (c *softwareCanvas) OnTypedRune() func(rune) {
	return c.onTypedRune
}

func (c *softwareCanvas) SetOnTypedRune(handler func(rune)) {
	c.onTypedRune = handler
}

func (c *softwareCanvas) OnTypedKey() func(*fyne.KeyEvent) {
	return c.onTypedKey
}

func (c *softwareCanvas) SetOnTypedKey(handler func(*fyne.KeyEvent)) {
	c.onTypedKey = handler
}

func (c *softwareCanvas) Capture() image.Image {
	theme := fyne.CurrentApp().Settings().Theme()

	size := c.Size().Union(c.content.MinSize())
	bounds := image.Rect(0, 0, int(float32(size.Width)*c.Scale()), int(float32(size.Height)*c.Scale()))
	base := image.NewRGBA(bounds)
	draw.Draw(base, bounds, image.NewUniform(theme.BackgroundColor()), image.ZP, draw.Src)

	paint := func(obj fyne.CanvasObject, pos fyne.Position) {
		if img, ok := obj.(*canvas.Image); ok {
			c.drawImage(img, pos, size, base)
		} else if text, ok := obj.(*canvas.Text); ok {
			c.drawText(text, pos, size, base)
		} else if rect, ok := obj.(*canvas.Rectangle); ok {
			c.drawRectangle(rect, pos, size, base)
		} else if wid, ok := obj.(fyne.Widget); ok {
			c.drawWidget(wid, pos, size, base)
		}
	}

	walkObjectTree(c.content, paint)
	return base
}

// NewSoftwareCanvas loads a new in-memory fyne canvas for quick rendering
func NewSoftwareCanvas(content fyne.CanvasObject) WindowlessCanvas {
	return &softwareCanvas{content: content}
}
