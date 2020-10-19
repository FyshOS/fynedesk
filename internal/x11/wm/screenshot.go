// +build linux openbsd freebsd netbsd

package wm

import (
	"image"
	"image/png"
	"math"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/widget"
	"github.com/BurntSushi/xgb/xproto"

	"fyne.io/fynedesk/internal/x11"
)

func (x *x11WM) captureWindow(win xproto.Window) *image.NRGBA {
	geom, err := xproto.GetGeometry(x.x.Conn(), xproto.Drawable(win)).Reply()
	if err != nil {
		fyne.LogError("Unable to get screen geometry", err)
		return nil
	}
	pix, err := xproto.GetImage(x.x.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(win),
		0, 0, geom.Width, geom.Height, math.MaxUint32).Reply()
	if err != nil {
		fyne.LogError("Error capturing window content", err)
		return nil
	}

	img := image.NewNRGBA(image.Rect(0, 0, int(geom.Width), int(geom.Height)))
	i := 0
	for y := 0; y < int(geom.Height); y++ {
		for x := 0; x < int(geom.Width); x++ {
			copyPixel(pix.Data, img.Pix, i)
			i += 4
		}
	}
	return img
}

func (x *x11WM) screenshot() {
	root := x.x.RootWin()
	bg := x.captureWindow(root)
	x.showCaptureSave(bg)
}

func (x *x11WM) screenshotWindow() {
	win := x.stack.TopWindow()
	if win == nil {
		fyne.LogError("Unable to print window with no window visible", nil)
		return
	}

	img := x.captureWindow(win.(x11.XWin).FrameID())
	if img == nil {
		return
	}
	x.showCaptureSave(img)
}

func (x *x11WM) showCaptureSave(img image.Image) {
	w := fyne.CurrentApp().NewWindow("Screenshot")
	save := widget.NewButton("Save...", func() {
		saveImage(img, w)
	})
	save.Importance = widget.HighImportance
	buttons := container.NewHBox(
		layout.NewSpacer(),
		widget.NewButton("Cancel", func() {
			w.Close()
		}),
		save)

	preview := canvas.NewImageFromImage(img)
	preview.FillMode = canvas.ImageFillContain
	w.SetContent(container.NewBorder(nil, buttons, nil, nil, preview))
	w.Resize(fyne.NewSize(400, 250))
	w.Show()
}

func copyPixel(in []byte, out []uint8, i int) {
	b := in[i]
	g := in[i+1]
	r := in[i+2]
	a := in[i+3]
	out[i] = r
	out[i+1] = g
	out[i+2] = b
	out[i+3] = a
}

func saveImage(pix image.Image, w fyne.Window) {
	d := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
		}

		err = png.Encode(write, pix)
		if err != nil {
			dialog.ShowError(err, w)
		}

		w.Close()
	}, w)
	d.SetFilter(storage.NewMimeTypeFileFilter([]string{"image/png"}))
	d.Show()
}
