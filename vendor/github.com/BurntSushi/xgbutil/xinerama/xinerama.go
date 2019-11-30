package xinerama

import "sort"

import (
	"github.com/BurntSushi/xgb/xinerama"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xrect"
)

// Alias so we use it as a receiver to satisfy sort.Interface
type Heads []xrect.Rect

// Len satisfies 'Len' in sort.Interface.
func (hds Heads) Len() int {
	return len(hds)
}

// Less satisfies 'Less' in sort.Interface.
func (hds Heads) Less(i int, j int) bool {
	return hds[i].X() < hds[j].X() || (hds[i].X() == hds[j].X() &&
		hds[i].Y() < hds[j].Y())
}

// Swap does just that. Nothing to see here...
func (hds Heads) Swap(i int, j int) {
	hds[i], hds[j] = hds[j], hds[i]
}

// PhyiscalHeads returns the list of heads in a physical ordering.
// Namely, left to right then top to bottom. (Defined by (X, Y).)
// Xinerama must have been initialized, otherwise the xinerama.QueryScreens
// request will panic.
// PhysicalHeads also checks to make sure each rectangle has a unique (x, y)
// tuple, so as not to return the geometry of cloned displays.
// (At present moment, xgbutil initializes Xinerama automatically during
// initial connection.)
func PhysicalHeads(xu *xgbutil.XUtil) (Heads, error) {
	xinfo, err := xinerama.QueryScreens(xu.Conn()).Reply()
	if err != nil {
		return nil, err
	}

	hds := make(Heads, 0)
	for _, info := range xinfo.ScreenInfo {
		head := xrect.New(int(info.XOrg), int(info.YOrg),
			int(info.Width), int(info.Height))

		// Maybe Xinerama is enabled, but we have cloned displays...
		unique := true
		for _, h := range hds {
			if h.X() == head.X() && h.Y() == head.Y() {
				unique = false
				break
			}
		}

		if unique {
			hds = append(hds, head)
		}
	}

	sort.Sort(hds)
	return hds, nil
}
