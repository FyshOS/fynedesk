package builtin

import "fyne.io/fynedesk"

func init() {
	// system area (bottom of widget panel) - order is top to bottom
	fynedesk.RegisterModule(networkMeta)
	fynedesk.RegisterModule(batteryMeta)
	fynedesk.RegisterModule(brightnessMeta)
}
