package builtin

import "fyne.io/fynedesk"

func init() {
	fynedesk.RegisterModule(batteryMeta)
	fynedesk.RegisterModule(soundMeta)
	fynedesk.RegisterModule(brightnessMeta)
}
