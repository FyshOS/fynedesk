package x11

import (
	"fyne.io/fyne"

	"fyne.io/fynedesk"
)

type clientProperties struct {
	c *client
}

func (c *client) Properties() fynedesk.WindowProperties {
	if c.props == nil {
		c.props = &clientProperties{c: c}
	}

	return c.props
}

func (c *clientProperties) Class() []string {
	return windowClass(c.c.wm.x, c.c.win)
}

func (c *clientProperties) Command() string {
	return windowCommand(c.c.wm.x, c.c.win)
}

func (c clientProperties) Decorated() bool {
	return !windowBorderless(c.c.wm.x, c.c.win)
}

func (c *clientProperties) Icon() fyne.Resource {
	settings := fynedesk.Instance().Settings()
	iconSize := int(float64(settings.LauncherIconSize()) * settings.LauncherZoomScale())
	xIcon := windowIcon(c.c.wm.x, c.c.win, iconSize, iconSize)
	if len(xIcon.Bytes()) != 0 {
		return fyne.NewStaticResource(c.Title(), xIcon.Bytes())
	}
	return nil
}

func (c *clientProperties) IconName() string {
	return windowIconName(c.c.wm.x, c.c.win)
}

func (c *clientProperties) SkipTaskbar() bool {
	extendedHints := windowExtendedHintsGet(c.c.wm.x, c.c.win)
	if extendedHints == nil {
		return false
	}
	for _, hint := range extendedHints {
		if hint == "_NET_WM_STATE_SKIP_TASKBAR" {
			return true
		}
	}
	return false
}

func (c *clientProperties) Title() string {
	return windowName(c.c.wm.x, c.c.win)
}
