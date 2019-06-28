package driver

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"image"
	"image/draw"
)

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
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, image.NewUniform(theme.BackgroundColor()), image.ZP, draw.Src)

	paint := func(obj fyne.CanvasObject, pos fyne.Position) {
		if text, ok := obj.(*canvas.Text); ok {
			bounds := text.MinSize()
			width := bounds.Width   //textureScaleInt(c, bounds.Width)
			height := bounds.Height //textureScaleInt(c, bounds.Height)
			txtImg := image.NewRGBA(image.Rect(0, 0, width, height))

			var opts truetype.Options
			fontSize := float64(text.TextSize) * float64(c.Scale())
			opts.Size = fontSize
			opts.DPI = 78.0
			f, _ := truetype.Parse(theme.TextFont().Content())
			face := truetype.NewFace(f, &opts)

			d := font.Drawer{}
			d.Dst = txtImg
			d.Src = &image.Uniform{C: text.Color}
			d.Face = face
			d.Dot = freetype.Pt(0, height-face.Metrics().Descent.Ceil())
			d.DrawString(text.Text)

			imgBounds := image.Rect(pos.X, pos.Y, text.Size().Width+pos.X, text.Size().Height+pos.Y)
			draw.Draw(img, imgBounds, txtImg, image.ZP, draw.Over)
		} else if rect, ok := obj.(*canvas.Rectangle); ok {
			bounds = image.Rect(pos.X, pos.Y, rect.Size().Width+pos.X, rect.Size().Height+pos.Y)
			draw.Draw(img, bounds, image.NewUniform(rect.FillColor), image.ZP, draw.Over)
		} else if wid, ok := obj.(fyne.Widget); ok {
			bounds = image.Rect(pos.X, pos.Y, wid.Size().Width+pos.X, wid.Size().Height+pos.Y)
			draw.Draw(img, bounds, image.NewUniform(widget.Renderer(wid).BackgroundColor()), image.ZP, draw.Over)
		}
	}

	walkObjectTree(c.content, paint)
	return img
}

func NewSoftwareCanvas(content fyne.CanvasObject) WindowlessCanvas {
	return &softwareCanvas{content: content}
}
