package builtin

import "fyne.io/fynedesk"

func init() {
	fynedesk.RegisterModule(batteryMeta)
	fynedesk.RegisterModule(pulseaudioMeta)
	fynedesk.RegisterModule(brightnessMeta)
}
