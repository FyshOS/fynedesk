package pacman

import (
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type resultWidget struct {
	widget.Label

	onTapped func()
}

func (r *resultWidget) Tapped(*fyne.PointEvent) {
	r.onTapped()
}

func (r *resultWidget) TappedSecondary(*fyne.PointEvent) {
}

func (u *updater) setResult(item *Package) {
	u.description.SetText(item.Description)
}

func (u *updater) fillResults(text string, list *widget.Box) {
	list.Children = nil

	var rows []fyne.CanvasObject
	for _, item := range u.manager.Search(text) {
		pkg := item // capture
		label := fmt.Sprintf("%s (%s)", item.Name, item.Version)
		result := &resultWidget{onTapped: func() {
			u.setResult(pkg)
		}}
		result.Text = label
		rows = append(rows, result)
	}

	list.Children = rows
	list.Refresh()
}

func (u *updater) loadSearchUI() fyne.CanvasObject {
	list := widget.NewVBox()
	input := widget.NewEntry()
	input.PlaceHolder = "Enter app name"
	search := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		go u.fillResults(input.Text, list)
	})
	u.description = widget.NewLabel("details")
	content := fyne.NewContainerWithLayout(layout.NewGridLayout(1),
		widget.NewScrollContainer(list), u.description)

	inputRow := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, search),
		search, input)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(inputRow, nil, nil, nil),
		inputRow, content)
}
