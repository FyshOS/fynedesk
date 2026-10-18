package ui

import (
	"bufio"
	_ "embed"
	"fmt"
	"image/color"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const zoneTabFile = "/usr/share/zoneinfo/zone.tab"

var (
	//go:embed worldmap.svg
	resourceWorldmapSvgData []byte

	// worldMap is an equirectangular (plate carrée) world map used by the
	// timezone picker.
	worldMapSvg = &fyne.StaticResource{
		StaticName:    "assets/worldmap.svg",
		StaticContent: resourceWorldmapSvgData,
	}
)

// zoneInfo is a single IANA timezone with the map coordinates of a
// representative location, as listed in zone1970.tab.
type zoneInfo struct {
	name    string
	comment string
	lat     float64
	lon     float64
}

var cachedZones []zoneInfo

// loadZones reads the system zone table once and returns the list of timezones
// sorted by name. Each entry carries a lat/long so it can be plotted on a map.
func loadZones() []zoneInfo {
	if cachedZones != nil {
		return cachedZones
	}

	f, err := os.Open(zoneTabFile)
	if err != nil {
		fyne.LogError("Failed to open timezone table", err)
		return nil
	}
	defer f.Close()

	var zones []zoneInfo
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// columns: country-codes, ISO-6709 coordinates, TZ name, [comment]
		cols := strings.Split(line, "\t")
		if len(cols) < 3 {
			continue
		}
		lat, lon, ok := parseCoord(cols[1])
		if !ok {
			continue
		}
		z := zoneInfo{name: cols[2], lat: lat, lon: lon}
		if len(cols) > 3 {
			z.comment = cols[3]
		}
		zones = append(zones, z)
	}

	sort.Slice(zones, func(i, j int) bool { return zones[i].name < zones[j].name })
	cachedZones = zones
	return zones
}

// parseCoord decodes an ISO-6709 coordinate such as "+4042-07400" or
// "+404251-0740023" into decimal degrees of latitude and longitude.
func parseCoord(s string) (lat, lon float64, ok bool) {
	if len(s) < 2 {
		return 0, 0, false
	}
	// the longitude sign marks where the latitude field ends
	split := strings.IndexAny(s[1:], "+-")
	if split < 0 {
		return 0, 0, false
	}
	split++
	lat, ok1 := parseDMS(s[:split], 2)
	lon, ok2 := parseDMS(s[split:], 3)
	return lat, lon, ok1 && ok2
}

// parseDMS parses a signed degrees/minutes[/seconds] field where the degrees
// component is degDigits long (2 for latitude, 3 for longitude).
func parseDMS(s string, degDigits int) (float64, bool) {
	if len(s) < 1+degDigits+2 {
		return 0, false
	}
	sign := 1.0
	if s[0] == '-' {
		sign = -1.0
	}
	body := s[1:]
	deg, err1 := strconv.Atoi(body[:degDigits])
	min, err2 := strconv.Atoi(body[degDigits : degDigits+2])
	if err1 != nil || err2 != nil {
		return 0, false
	}
	val := float64(deg) + float64(min)/60
	if len(body) >= degDigits+4 { // optional seconds
		if sec, err := strconv.Atoi(body[degDigits+2 : degDigits+4]); err == nil {
			val += float64(sec) / 3600
		}
	}
	return sign * val, true
}

// offsetString returns the current UTC offset of a zone (DST aware), e.g. "UTC+1".
func offsetString(name string) string {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return ""
	}
	_, off := time.Now().In(loc).Zone()
	sign := "+"
	if off < 0 {
		sign = "-"
		off = -off
	}
	h := off / 3600
	m := (off % 3600) / 60
	if m == 0 {
		return fmt.Sprintf("UTC%s%d", sign, h)
	}
	return fmt.Sprintf("UTC%s%d:%02d", sign, h, m)
}

// worldMap is a visual, tappable equirectangular world map that plots a marker
// for every timezone and highlights the selected one.
type worldMap struct {
	widget.BaseWidget

	zones      []zoneInfo
	selected   int
	OnSelected func(zoneInfo)
}

func newWorldMap(zones []zoneInfo) *worldMap {
	w := &worldMap{zones: zones, selected: -1}
	w.ExtendBaseWidget(w)
	return w
}

// SetSelected highlights the zone with the given name without firing OnSelected.
func (w *worldMap) SetSelected(name string) {
	for i, z := range w.zones {
		if z.name == name {
			if w.selected != i {
				w.selected = i
				w.Refresh()
			}
			return
		}
	}
}

// project converts a zone's lat/long into a pixel position within the map area,
// matching the plate carrée image drawn with ImageFillContain inside size.
func project(z zoneInfo, size fyne.Size) fyne.Position {
	offX, offY, dw, dh := mapRect(size)
	x := offX + float32((z.lon+180)/360)*dw
	y := offY + float32((90-z.lat)/180)*dh
	return fyne.NewPos(x, y)
}

// mapRect returns the rectangle the 2:1 map image actually occupies once
// centred within size by ImageFillContain.
func mapRect(size fyne.Size) (offX, offY, dw, dh float32) {
	const aspect = 2.0 // 360 / 180
	if size.Width/size.Height > aspect {
		dh = size.Height
		dw = dh * aspect
	} else {
		dw = size.Width
		dh = dw / aspect
	}
	offX = (size.Width - dw) / 2
	offY = (size.Height - dh) / 2
	return
}

// Tapped selects the timezone whose marker is nearest the tap position.
func (w *worldMap) Tapped(ev *fyne.PointEvent) {
	size := w.Size()
	best := -1
	bestDist := float32(math.MaxFloat32)
	for i, z := range w.zones {
		p := project(z, size)
		dx := p.X - ev.Position.X
		dy := p.Y - ev.Position.Y
		d := dx*dx + dy*dy
		if d < bestDist {
			bestDist = d
			best = i
		}
	}
	if best < 0 {
		return
	}
	w.selected = best
	w.Refresh()
	if w.OnSelected != nil {
		w.OnSelected(w.zones[best])
	}
}

func (w *worldMap) CreateRenderer() fyne.WidgetRenderer {
	img := canvas.NewImageFromResource(worldMapSvg)
	img.FillMode = canvas.ImageFillContain

	dots := make([]*canvas.Circle, len(w.zones))
	objs := []fyne.CanvasObject{img}
	for i := range w.zones {
		c := canvas.NewCircle(color.Transparent)
		dots[i] = c
		objs = append(objs, c)
	}
	highlight := canvas.NewCircle(color.Transparent)
	highlight.StrokeWidth = 2
	objs = append(objs, highlight)

	r := &worldMapRenderer{m: w, img: img, dots: dots, highlight: highlight, objects: objs}
	r.applyColors()
	return r
}

const (
	dotRadius       = float32(1.6)
	highlightRadius = float32(5)
)

type worldMapRenderer struct {
	m         *worldMap
	img       *canvas.Image
	dots      []*canvas.Circle
	highlight *canvas.Circle
	objects   []fyne.CanvasObject
}

func (r *worldMapRenderer) Layout(size fyne.Size) {
	r.img.Resize(size)
	r.img.Move(fyne.NewPos(0, 0))
	for i, z := range r.m.zones {
		p := project(z, size)
		r.dots[i].Move(p.SubtractXY(dotRadius, dotRadius))
		r.dots[i].Resize(fyne.NewSize(dotRadius*2, dotRadius*2))
	}
	r.positionHighlight(size)
}

func (r *worldMapRenderer) positionHighlight(size fyne.Size) {
	if r.m.selected < 0 {
		r.highlight.Resize(fyne.NewSize(0, 0))
		return
	}
	p := project(r.m.zones[r.m.selected], size)
	r.highlight.Move(p.SubtractXY(highlightRadius, highlightRadius))
	r.highlight.Resize(fyne.NewSize(highlightRadius*2, highlightRadius*2))
}

func (r *worldMapRenderer) MinSize() fyne.Size {
	return fyne.NewSize(320, 160)
}

// applyColors sets marker colours from the current theme.
func (r *worldMapRenderer) applyColors() {
	dot := theme.Color(theme.ColorNameForeground)
	rr, gg, bb, _ := dot.RGBA()
	faint := color.NRGBA{R: uint8(rr >> 8), G: uint8(gg >> 8), B: uint8(bb >> 8), A: 110}
	for _, c := range r.dots {
		c.FillColor = faint
	}
	r.highlight.FillColor = theme.Color(theme.ColorNamePrimary)
	r.highlight.StrokeColor = theme.Color(theme.ColorNameBackground)
}

func (r *worldMapRenderer) Refresh() {
	r.applyColors()
	for _, c := range r.dots {
		c.Refresh()
	}
	r.positionHighlight(r.m.Size())
	r.highlight.Refresh()
	canvas.Refresh(r.m)
}

func (r *worldMapRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *worldMapRenderer) Destroy() {}

// loadTimeScreen builds the Time/Date settings panel: a time area with an
// automatic (NTP) toggle and manual fallback, plus a visual timezone picker.
func (d *settingsUI) loadTimeScreen() fyne.CanvasObject {
	zones := loadZones()

	// --- time controls ---
	now := time.Now()
	dateEntry := widget.NewEntry()
	dateEntry.SetText(now.Format("2006-01-02"))
	timeEntry := widget.NewEntry()
	timeEntry.SetText(now.Format("15:04"))

	// Automatic (network) time is the default. We track the inverse ("manual")
	// internally so its zero value means automatic, but present it as a positive
	// "set automatically" toggle that is checked whenever manual is off.
	// Assume automatic (network) time until timedatectl is queried below (off the
	// render thread); the real value is applied to these widgets asynchronously.
	manual := false
	auto := widget.NewCheck("Set time automatically (network time)", nil)
	auto.SetChecked(!manual)
	updateManual := func() {
		manual = !auto.Checked
		if manual {
			dateEntry.Enable()
			timeEntry.Enable()
		} else {
			dateEntry.Disable()
			timeEntry.Disable()
		}
	}
	auto.OnChanged = func(bool) { updateManual() }
	updateManual()

	manualRow := container.NewBorder(nil, nil, widget.NewLabel("Date / Time"), nil,
		container.NewGridWithColumns(2, dateEntry, timeEntry))
	timeSection := container.NewVBox(auto, manualRow, widget.NewSeparator())

	// --- timezone picker ---
	names := make([]string, len(zones))
	byName := make(map[string]zoneInfo, len(zones))
	for i, z := range zones {
		names[i] = z.name
		byName[z.name] = z
	}

	mapWidget := newWorldMap(zones)
	selectedLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	search := widget.NewSelectEntry(names)
	search.SetPlaceHolder("Search for a city or region")

	current := ""
	updating := false
	selectZone := func(name string) {
		if updating {
			return
		}
		z, ok := byName[name]
		if !ok {
			return
		}
		updating = true
		current = name
		mapWidget.SetSelected(name)
		if search.Text != name {
			search.SetText(name)
		}
		label := name
		if off := offsetString(name); off != "" {
			label += "   (" + off + ")"
		}
		if z.comment != "" {
			label += "  –  " + z.comment
		}
		selectedLabel.SetText("Selected: " + label)
		updating = false
	}
	mapWidget.OnSelected = func(z zoneInfo) { selectZone(z.name) }
	search.OnChanged = func(s string) { selectZone(s) }

	// network calls run in the background.
	go func() {
		ntp := currentNTP()
		tz := currentTimezone()
		fyne.Do(func() {
			auto.SetChecked(ntp)
			updateManual()
			if tz != "" {
				selectZone(tz)
			}
		})
	}()

	tzSection := container.NewBorder(widget.NewRichTextFromMarkdown("## Time Zone"), nil, nil, nil,
		container.NewBorder(search, selectedLabel, nil, nil, mapWidget))

	// --- apply ---
	var applyButton *widget.Button
	applyButton = &widget.Button{Text: "Apply", Importance: widget.HighImportance, OnTapped: func() {
		applyButton.Disable()
		wantNTP := !manual
		wantZone := current
		dateText, timeText := dateEntry.Text, timeEntry.Text
		go func() {
			err := d.applyTimeSettings(wantZone, wantNTP, dateText, timeText)
			fyne.Do(func() {
				applyButton.Enable()
				if err != nil {
					dialog.ShowError(err, d.win)
				}
			})
		}()
	}}
	apply := container.NewHBox(layout.NewSpacer(), applyButton)

	return container.NewBorder(timeSection, apply, nil, nil, tzSection)
}

// applyTimeSettings pushes the requested timezone, NTP and (when manual) clock
// values to the system via timedatectl. Each privileged call triggers tyde's
// PolicyKit auth prompt. This blocks, so it must run off the UI goroutine.
func (d *settingsUI) applyTimeSettings(zone string, ntp bool, dateText, timeText string) error {
	if zone != "" && zone != currentTimezone() {
		if err := runTimedatectl("set-timezone", zone); err != nil {
			return err
		}
	}
	if ntp != currentNTP() {
		val := "false"
		if ntp {
			val = "true"
		}
		if err := runTimedatectl("set-ntp", val); err != nil {
			return err
		}
	}
	if !ntp { // timedatectl refuses set-time while NTP is active
		when, err := time.ParseInLocation("2006-01-02 15:04", dateText+" "+timeText, time.Local)
		if err != nil {
			return fmt.Errorf("invalid date or time: %w", err)
		}
		if err := runTimedatectl("set-time", when.Format("2006-01-02 15:04:05")); err != nil {
			return err
		}
	}
	return nil
}

func runTimedatectl(args ...string) error {
	out, err := exec.Command("timedatectl", args...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%s: %s", strings.Join(args, " "), msg)
		}
		return err
	}
	return nil
}

func currentTimezone() string {
	out, err := exec.Command("timedatectl", "show", "-p", "Timezone", "--value").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func currentNTP() bool {
	out, err := exec.Command("timedatectl", "show", "-p", "NTP", "--value").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "yes"
}
