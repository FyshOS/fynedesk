//go:generate fyne bundle -o about_bundled.go -package ui -name resourceAuthors ../../AUTHORS

package ui

import (
	"image/color"
	"net/url"
	"runtime/debug"
	"strings"

	theme2 "fyshos.com/fynedesk/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Authors contains the list of the toolkit authors, extracted from ../../../AUTHORS.
var Authors = resourceAuthors

func newURLButton(label, link string) *widget.Hyperlink {
	u, _ := url.Parse(link)
	return widget.NewHyperlink(label, u)
}

// showAbout opens a parallax about screen mimicking (and using code from)
// the [fyne_demo](https://github.com/fyne-io/fyne/tree/develop/cmd/fyne_demo)
// welcome panel.
func (w *widgetPanel) showAbout() {
	if w.about != nil {
		w.about.CenterOnScreen()
		w.about.Show()

		for _, win := range w.desk.WindowManager().Windows() {
			if win.Properties().Title() == w.about.Title() {
				win.SetDesktop(w.desk.Desktop())
				w.desk.WindowManager().RaiseToTop(win)
				break
			}
		}
		return
	}
	win := fyne.CurrentApp().NewWindow("About FyneDesk")

	logo := canvas.NewImageFromResource(theme2.FyshOSLogo)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(256, 256))

	footer := container.NewHBox(
		layout.NewSpacer(),
		newURLButton("Home Page", "https://fyshos.com/fynedesk"),
		widget.NewLabel("-"),
		newURLButton("Report Issue", "https://github.com/FyshOS/fynedesk/issues/new"),
		widget.NewLabel("-"),
		newURLButton("Sponsor", "https://github.com/sponsors/fyne-io"),
		layout.NewSpacer(),
	)

	authors := widget.NewRichTextFromMarkdown(formatAuthors(string(Authors.Content())))
	content := container.NewVBox(
		container.NewCenter(
			widget.NewRichTextFromMarkdown("**Version:** "+version())),
		logo,
		container.NewCenter(authors),
		widget.NewLabelWithStyle("\nWith great thanks to our many kind contributors\n", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
	scroll := container.NewScroll(content)

	bgColor := withAlpha(theme.BackgroundColor(), 0xe0)
	shadowColor := withAlpha(theme.BackgroundColor(), 0x33)

	underlay := canvas.NewImageFromResource(theme2.FyshOSLogo)
	bg := canvas.NewRectangle(bgColor)
	underlayer := underLayout{}
	slideBG := container.New(underlayer, underlay)
	footerBG := canvas.NewRectangle(shadowColor)

	listen := make(chan fyne.Settings)
	fyne.CurrentApp().Settings().AddChangeListener(listen)
	go func() {
		for range listen {
			bgColor = withAlpha(theme.BackgroundColor(), 0xe0)
			bg.FillColor = bgColor
			bg.Refresh()

			shadowColor = withAlpha(theme.BackgroundColor(), 0x33)
			footerBG.FillColor = bgColor
			footer.Refresh()
		}
	}()

	underlay.Resize(fyne.NewSize(512, 512))
	scroll.OnScrolled = func(p fyne.Position) {
		underlayer.offset = -p.Y / 3
		underlayer.Layout(slideBG.Objects, slideBG.Size())
	}

	bgClip := container.NewScroll(slideBG)
	bgClip.Direction = container.ScrollNone
	win.SetContent(container.NewStack(container.New(unpad{top: true}, bgClip, bg),
		container.NewBorder(nil,
			container.NewStack(footerBG, footer), nil, nil,
			container.New(unpad{top: true, bottom: true}, scroll))))
	win.SetCloseIntercept(func() {
		win.Hide()
	})

	w.about = win
	win.Resize(fyne.NewSize(340, 280))
	win.Show()
}

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}

	return "(devel)"
}

func withAlpha(c color.Color, alpha uint8) color.Color {
	r, g, b, _ := c.RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: alpha}
}

type underLayout struct {
	offset float32
}

func (u underLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	under := objs[0]
	left := size.Width/2 - under.Size().Width/2
	under.Move(fyne.NewPos(left, u.offset-50))
}

func (u underLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.Size{}
}

type unpad struct {
	top, bottom bool
}

func (u unpad) Layout(objs []fyne.CanvasObject, s fyne.Size) {
	pad := theme.Padding()
	var pos fyne.Position
	if u.top {
		pos = fyne.NewPos(0, -pad)
	}
	size := s
	if u.top {
		size = size.AddWidthHeight(0, pad)
	}
	if u.bottom {
		size = size.AddWidthHeight(0, pad)
	}
	for _, o := range objs {
		o.Move(pos)
		o.Resize(size)
	}
}

func (u unpad) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 100)
}

func formatAuthors(lines string) string {
	markdown := &strings.Builder{}
	markdown.WriteString("### Authors\n\n")

	for _, line := range strings.Split(lines, "\n") {
		if len(line) == 0 {
			continue
		}

		markdown.WriteString("* ")
		markdown.WriteString(line)
		markdown.WriteByte('\n')
	}

	return markdown.String()
}
