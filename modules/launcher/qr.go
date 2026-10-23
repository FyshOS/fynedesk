package launcher

import (
	_ "embed"
	"image/color"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyshos.com/tyde/internal/modhelper"

	"github.com/FyshOS/fyqr/pkg/qrgen"

	"fyshos.com/tyde"
)

var qrAliases = []string{"qr", "qrcode"}

var qrMeta = tyde.ModuleMetadata{
	Name:        "Launcher: QR Codes",
	NewInstance: newQR,
}

//go:embed qr.svg
var resourceQRSvgData []byte

var resourceQR = &fyne.StaticResource{
	StaticName:    "qr.svg",
	StaticContent: resourceQRSvgData,
}

type qr struct{}

func (q *qr) Destroy() {
}

func (q *qr) Metadata() tyde.ModuleMetadata {
	return qrMeta
}

// Commands offers the QR code display, under the first of its aliases.
func (q *qr) Commands() []tyde.Command {
	c := tyde.Command{
		Name:        qrAliases[0],
		Description: "Show a scannable QR code for a URL or piece of text",
		Args:        "the URL or text to encode",
		Examples:    []string{"qr https://fyshos.com"},
	}

	modhelper.AnswerWith(&c, q.item)
	return []tyde.Command{c}
}

func (q *qr) item(content string) tyde.LaunchSuggestion {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	return &qrItem{content: content}
}

func (q *qr) LaunchSuggestions(input string) []tyde.LaunchSuggestion {
	lower := strings.ToLower(input)

	for _, alias := range qrAliases {
		prefix := alias + " "
		if strings.HasPrefix(lower, prefix) {
			content := strings.TrimSpace(input[len(prefix):])
			return []tyde.LaunchSuggestion{&qrItem{content: content}}
		}
		if strings.HasPrefix(alias, lower) {
			return []tyde.LaunchSuggestion{&qrItem{}}
		}
	}

	return nil
}

// newQR creates a new module that turns "qr <url>" into a scannable code.
func newQR() tyde.Module {
	return &qr{}
}

type qrItem struct {
	content string
}

func (i *qrItem) Icon() fyne.Resource {
	return theme.NewThemedResource(resourceQR)
}

func (i *qrItem) Title() string {
	if i.content == "" {
		return "QR Code: type a URL or text"
	}
	return "QR Code: " + i.content
}

func (i *qrItem) Launch() {
	if i.content == "" {
		return
	}

	// Prefer the standalone fyqr app when it is installed, otherwise fall
	// back to rendering the code ourselves as an overlay.
	if path, err := exec.LookPath("fyqr"); err == nil {
		cmd := exec.Command(path, i.content)
		if err := cmd.Start(); err == nil {
			go func() { _ = cmd.Wait() }()
			return
		} else {
			fyne.LogError("Failed to launch fyqr", err)
		}
	}

	i.showOverlay()
}

// showOverlay renders the QR code centred on a dimmed, blurred backdrop.
func (i *qrItem) showOverlay() {
	img, err := qrgen.NewQR(i.content).Image(512)
	if err != nil {
		fyne.LogError("Failed to generate QR code", err)
		return
	}

	desk := tyde.Instance()
	screen := desk.Screens().Primary()
	scale := screen.CanvasScale()

	screenW := float32(screen.Width) / scale
	screenH := float32(screen.Height) / scale

	code := canvas.NewImageFromImage(img)
	code.ScaleMode = canvas.ImageScalePixels
	code.FillMode = canvas.ImageFillContain
	code.SetMinSize(fyne.NewSquareSize(360))

	caption := canvas.NewText(i.content, color.Black)
	caption.Alignment = fyne.TextAlignCenter

	white := canvas.NewRectangle(color.White)
	white.CornerRadius = theme.Padding() * 2
	card := container.NewStack(white,
		container.NewPadded(container.NewVBox(code, caption)))

	r, g, b, _ := theme.Color(theme.ColorNameOverlayBackground).RGBA()
	bg := canvas.NewRectangle(&color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0x80})

	var overlay fyne.CanvasObject
	dismiss := &largeTypeTappable{onTap: func() {
		desk.HideOverlay(overlay)
	}}
	dismiss.ExtendBaseWidget(dismiss)
	overlay = container.NewStack(canvas.NewBlur(5), bg, dismiss, container.NewCenter(card))

	size := fyne.NewSize(screenW, screenH)
	desk.ShowOverlay(overlay, size, fyne.NewPos(0, 0))
	fyne.Do(func() {
		desk.Root().Canvas().Focus(dismiss)
	})
}
