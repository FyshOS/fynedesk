package ui

import (
	"image"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func (l *desktop) screenshot() {
	bg := l.wm.Capture()
	l.showCaptureSave(bg)
}

func (l *desktop) screenshotWindow() {
	win := l.wm.TopWindow()
	if win == nil {
		fyne.LogError("Unable to print window with no window visible", nil)
		return
	}

	if img := win.Capture(); img != nil {
		l.showCaptureSave(img)
	}
}

func (l *desktop) showCaptureSave(img image.Image) {
	w := fyne.CurrentApp().NewWindow("Screenshot")

	save := &widget.Button{Text: "Save...",
		Importance: widget.HighImportance,
		OnTapped: func() {
			saveImage(img, w)
		},
	}

	buttons := container.NewHBox(
		layout.NewSpacer(),
		widget.NewButton("Cancel", w.Close),
		save,
	)

	preview := canvas.NewImageFromImage(img)
	preview.FillMode = canvas.ImageFillContain

	w.SetContent(container.NewBorder(nil, buttons, nil, nil, preview))
	w.Resize(fyne.NewSize(480, 360))
	w.Show()
}

func saveImage(pix image.Image, w fyne.Window) {
	d := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
		if write == nil { // cancelled
			return
		}

		if err != nil {
			dialog.ShowError(err, w)
		} else if err = png.Encode(write, pix); err != nil {
			dialog.ShowError(err, w)
		}

		err = write.Close()
		if err != nil {
			dialog.ShowError(err, w)
		}

		w.Close()
	}, w)

	d.SetFilter(storage.NewMimeTypeFileFilter([]string{"image/png"}))
	d.SetFileName("screenshot.png")

	if dir, err := getPicturesDir(); err == nil {
		d.SetLocation(dir)
	} else {
		fyne.LogError("error finding pictures dir, falling back to home directory", err)
	}

	d.Show()
}
