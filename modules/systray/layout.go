package systray

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
)

type innerLayout fyne.Layout

type minGridWrap struct {
	innerLayout
}

func collapsingGridWrap(s fyne.Size) fyne.Layout {
	w := &minGridWrap{innerLayout: layout.NewGridWrapLayout(s)}
	return w
}

func (m *minGridWrap) MinSize(objs []fyne.CanvasObject) fyne.Size {
	if len(objs) == 0 {
		return fyne.NewSize(0, -theme.Padding())
	}

	return m.innerLayout.MinSize(objs)
}
