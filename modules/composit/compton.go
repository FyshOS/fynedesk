package composit

import (
	"os/exec"

	"fyne.io/fyne/v2"
	"fyshos.com/fynedesk"
)

var compizMeta = fynedesk.ModuleMetadata{
	Name:        "Compositor",
	NewInstance: newCompiz,
}

type comp struct {
}

func (c *comp) Destroy() {
	c.disable()
}

func (c *comp) Metadata() fynedesk.ModuleMetadata {
	return compizMeta
}

func (c *comp) disable() {
	_ = exec.Command("killall", "compton").Start()
}

func (c *comp) enable() {
	path, err := exec.LookPath("compton")
	if err != nil {
		fyne.LogError("Compositor requires compton binary present", err)
		return
	}

	_ = exec.Command(path, "--vsync", "drm", "-c", "-C", "-r", "20", "-f", "-i", "1.0").Start()
}

// newCompiz creates a new module that will manage composition of the windows.
func newCompiz() fynedesk.Module {
	c := &comp{}
	c.enable()
	return c
}
