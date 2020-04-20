package christmas

import (
	"math/rand"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"

	"fyne.io/fynedesk"
)

func init() {
	fynedesk.RegisterModule(christmasMeta)
}

var christmasMeta = fynedesk.ModuleMetadata{
	Name:        "Christmas",
	NewInstance: NewChristmas,
}

type christmas struct {
}

// NewChristmas creates a new module that adds christmas decorations
func NewChristmas() fynedesk.Module {
	return &christmas{}
}

func (c *christmas) Destroy() {
}

func (c *christmas) Metadata() fynedesk.ModuleMetadata {
	return christmasMeta
}

func (c *christmas) newTree() fyne.CanvasObject {
	tree := canvas.NewImageFromResource(resourceTreeSvg)
	tree.SetMinSize(fyne.NewSize(105, 150))
	return tree
}

type forrestLayout struct {
}

func (l *forrestLayout) MinSize(trees []fyne.CanvasObject) fyne.Size {
	return trees[0].MinSize()
}

func (l *forrestLayout) Layout(trees []fyne.CanvasObject, size fyne.Size) {
	space := size.Subtract(l.MinSize(trees))

	rand.Seed(time.Now().Unix())
	for _, tree := range trees {
		pos := fyne.NewPos(int(rand.Float32()*float32(space.Width)), int(rand.Float32()*float32(space.Height)))
		tree.Move(pos)
		tree.Resize(tree.MinSize())
	}
}

func (c *christmas) newForrest() *fyne.Container {
	trees := []fyne.CanvasObject{c.newTree(), c.newTree(), c.newTree(), c.newTree(), c.newTree(), c.newTree(), c.newTree()}
	return fyne.NewContainerWithLayout(&forrestLayout{}, trees...)
}

func (c *christmas) NotificationAreaWidget() fyne.CanvasObject {
	return c.StatusAreaWidget()
}

func (c *christmas) StatusAreaWidget() fyne.CanvasObject {
	lights := canvas.NewImageFromResource(resourceLightsPng)
	lights.FillMode = canvas.ImageFillStretch

	lights.SetMinSize(fyne.NewSize(105, 28))
	return lights
}

func (c *christmas) ScreenAreaWidget() fyne.CanvasObject {
	return c.newForrest()
}
