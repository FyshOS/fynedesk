package wm

import (
	"fyne.io/fyne"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xinerama"
)

func getHeads(x *xgbutil.XUtil) xinerama.Heads {
	if !x.ExtInitialized("XINERAMA") {
		return nil
	}
	heads, err := xinerama.PhysicalHeads(x)
	if err != nil || len(heads) == 0 {
		fyne.LogError("Error findings heads", err)
		return nil
	}
	return heads
}

func getHeadGeometry(headIndex int, heads xinerama.Heads) (int16, int16, uint16, uint16) {
	if headIndex < len(heads) {
		i := headIndex
		return int16(heads[i].X()), int16(heads[i].Y()), uint16(heads[i].Width()), uint16(heads[i].Height())
	}
	return 0, 0, 0, 0
}
