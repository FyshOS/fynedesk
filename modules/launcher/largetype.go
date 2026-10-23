package launcher

import (
	_ "embed"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyshos.com/tyde/internal/modhelper"

	"fyshos.com/tyde"
)

var largeTypeAliases = []string{"largetype", "large", "big", "bigtype", "type"}

var largeTypeMeta = tyde.ModuleMetadata{
	Name:        "Launcher: Large Type",
	NewInstance: newLargeType,
}

//go:embed largetype.svg
var resourceLargeTypeSvgData []byte

var resourceLargeType = &fyne.StaticResource{
	StaticName:    "largetype.svg",
	StaticContent: resourceLargeTypeSvgData,
}

type largeType struct{}

func (l *largeType) Destroy() {
}

func (l *largeType) Metadata() tyde.ModuleMetadata {
	return largeTypeMeta
}

// Commands offers the large type display. Only the first alias is listed: the
// rest are typing shortcuts, not separate things to do.
func (l *largeType) Commands() []tyde.Command {
	c := tyde.Command{
		Name:        largeTypeAliases[0],
		Description: "Show a short piece of text across the screen in large type",
		Args:        "the text to display",
		Examples:    []string{"largetype Hello World"},
	}

	modhelper.AnswerWith(&c, l.item)
	return []tyde.Command{c}
}

func (l *largeType) item(text string) tyde.LaunchSuggestion {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return &largeTypeItem{text: text}
}

func (l *largeType) LaunchSuggestions(input string) []tyde.LaunchSuggestion {
	lower := strings.ToLower(input)

	for _, alias := range largeTypeAliases {
		prefix := alias + " "
		if strings.HasPrefix(lower, prefix) {
			text := input[len(prefix):]
			return []tyde.LaunchSuggestion{&largeTypeItem{text: text}}
		}
		if strings.HasPrefix(alias, lower) {
			return []tyde.LaunchSuggestion{&largeTypeItem{}}
		}
	}

	return nil
}

func newLargeType() tyde.Module {
	return &largeType{}
}

type largeTypeItem struct {
	text string
}

func (i *largeTypeItem) Icon() fyne.Resource {
	return theme.NewThemedResource(resourceLargeType)
}

func (i *largeTypeItem) Title() string {
	return "Large Type: " + i.text
}

func (i *largeTypeItem) Launch() {
	desk := tyde.Instance()
	screen := desk.Screens().Primary()
	scale := screen.CanvasScale()

	screenW := float32(screen.Width) / scale
	screenH := float32(screen.Height) / scale

	label := canvas.NewText(i.text, theme.Color(theme.ColorNameForeground))
	label.TextSize = 120

	// just 3 emoji/symbols or less
	if len([]rune(i.text)) <= 3 && len([]rune(i.text)) != len(i.text) {
		label.TextSize = 250
	}

	label.Alignment = fyne.TextAlignCenter
	label.TextStyle = fyne.TextStyle{Bold: true}

	r, g, b, _ := theme.Color(theme.ColorNameOverlayBackground).RGBA()
	bg := canvas.NewRectangle(&color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0x80})

	var overlay fyne.CanvasObject
	dismiss := &largeTypeTappable{onTap: func() {
		desk.HideOverlay(overlay)
	}}
	dismiss.ExtendBaseWidget(dismiss)
	overlay = container.NewStack(canvas.NewBlur(5), bg, dismiss, container.NewCenter(label))

	size := fyne.NewSize(screenW, screenH)
	pos := fyne.NewPos(0, 0)
	desk.ShowOverlay(overlay, size, pos)
	fyne.Do(func() {
		desk.Root().Canvas().Focus(dismiss)
	})
}

// largeTypeTappable is a transparent widget that dismisses the overlay on tap.
type largeTypeTappable struct {
	widget.BaseWidget
	onTap func()
}

func (t *largeTypeTappable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (t *largeTypeTappable) Tapped(_ *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}

func (t *largeTypeTappable) FocusGained()   {}
func (t *largeTypeTappable) FocusLost()     {}
func (t *largeTypeTappable) TypedRune(rune) {}

func (t *largeTypeTappable) TypedKey(_ *fyne.KeyEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}
