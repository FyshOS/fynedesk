package composit

import (
	"fyne.io/fyne/v2"
	"fyshos.com/fynedesk"
)

var compMeta = fynedesk.ModuleMetadata{
	Name:        "Compositor",
	NewInstance: newCompositor,
}

type comp struct {
	done chan struct{}
}

func (c *comp) Destroy() {
	c.disable()
}

func (c *comp) Metadata() fynedesk.ModuleMetadata {
	return compMeta
}

func (c *comp) disable() {
	close(c.done)
}

func (c *comp) enable() {
	go func() {
		err := run(c.done)
		if err != nil {
			fyne.LogError("Compositor failed", err)
		}
	}()
}

// newCompositor creates a new module that will manage composition of the windows.
func newCompositor() fynedesk.Module {
	c := &comp{done: make(chan struct{})}
	c.enable()
	return c
}
