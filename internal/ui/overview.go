package ui

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	xdraw "golang.org/x/image/draw"
)

// deskOverview drives the "reveal all" desktop overview: a zoom-out to show every
// virtual desktop laid out in a grid, an interactive layer to pick one (or escape
// to cancel), then a zoom back in to the chosen desktop. The zoom itself is played
// by per-screen overview shaders over a stitched snapshot of all desktops; the
// interactive layer sits above the shader on the primary screen only and is shown
// once the zoom-out settles, mirroring how startDeskCube hands off to the live
// desktop when its roll completes.
type deskOverview struct {
	desk    *desktop
	count   int
	focus   int // desktop the view was opened from (current desktop)
	active  bool
	shaders []*canvas.Shader  // per-screen shaders currently playing
	anim    *fyne.Animation   // the in-flight zoom animation
	content fyne.CanvasObject // interactive selection layer on the primary overlay
}

// ShowDesktopOverview reveals all virtual desktops laid out in a grid and lets the
// user pick one. count is supplied by the desktops module, which owns the
// desktop count, so the core need not track it. A no-op when fewer than two
// desktops exist or an overview is already on screen.
func (l *desktop) ShowDesktopOverview(count int) {
	if count <= 1 {
		return
	}
	if l.overview != nil && l.overview.active {
		return
	}

	o := &deskOverview{desk: l, count: count, focus: l.desk, active: true}
	l.overview = o
	o.open()
}

// open captures every desktop into a per-screen strip, shows the shaders and plays
// the zoom-out. When it settles the interactive selection layer is revealed.
func (o *deskOverview) open() {
	l := o.desk
	var shaders []*canvas.Shader
	for _, sw := range l.screenWindows {
		if sw.overviewShader == nil || sw.screen == nil {
			continue
		}

		strip := l.buildDeskStrip(sw, o.count)
		if strip == nil {
			continue
		}

		sw.overviewShader.Textures = map[string]image.Image{"strip": strip}
		sw.overviewShader.Uniforms = map[string]float32{
			"progress": 0,
			"count":    float32(o.count),
			"cols":     float32(overviewGridCols(o.count)),
			"gap":      overviewGap,
			"focus":    float32(o.focus),
		}
		sw.overviewShader.Show()
		shaders = append(shaders, sw.overviewShader)
	}

	if len(shaders) == 0 {
		// Nothing could be captured (e.g. no shader/compositor) - abandon quietly.
		o.active = false
		l.overview = nil
		return
	}

	o.shaders = shaders
	o.animate(0, 1, o.focus, func() {
		o.showContent()
	})
}

// showContent reveals the interactive selection layer over the settled zoom-out.
// It lives on the primary overlay, above the overview shader, and is focused so it
// receives the Escape key.
func (o *deskOverview) showContent() {
	l := o.desk
	if !o.active || l.primaryWin == nil || l.primaryWin.win == nil {
		return
	}

	size := l.primaryWin.win.Canvas().Size()
	content := newOverviewContent(o.count, l.desk, size,
		func(d int) { o.selectDesk(d) },
		func() { o.cancel() })
	o.content = content
	l.showOverlay(content, size, fyne.NewPos(0, 0), content)
}

// selectDesk zooms in to the tapped desktop and switches to it. The window slide
// is started without the cube transition so it happens unseen beneath the zoom-in,
// exactly as SetDesktop's slide happens beneath the cube.
func (o *deskOverview) selectDesk(d int) {
	if !o.active || o.anim != nil {
		return
	}

	o.removeContent()
	if d != o.desk.desk {
		o.desk.setDesktop(d, false)
	}
	o.close(d)
}

// cancel zooms back in to the desktop the overview was opened from, changing nothing.
func (o *deskOverview) cancel() {
	if !o.active || o.anim != nil {
		return
	}

	o.removeContent()
	o.close(o.focus)
}

// close plays the zoom-in to the given desktop then tears the overview down.
func (o *deskOverview) close(focus int) {
	o.animate(1, 0, focus, func() {
		for _, s := range o.shaders {
			s.Hide()
		}
		o.shaders = nil
		o.active = false
		if o.desk.overview == o {
			o.desk.overview = nil
		}
	})
}

func (o *deskOverview) removeContent() {
	if o.content != nil {
		o.desk.HideOverlay(o.content)
		o.content = nil
	}
}

// animate drives every screen's shader progress from->to with the given focus,
// invoking onDone once when it completes. A newer animation supersedes an older one.
func (o *deskOverview) animate(from, to float32, focus int, onDone func()) {
	for _, s := range o.shaders {
		s.Uniforms["focus"] = float32(focus)
	}

	var a *fyne.Animation
	a = fyne.NewAnimation(canvas.DurationStandard, func(f float32) {
		if o.anim != a {
			return // superseded
		}

		p := from + (to-from)*f
		for _, s := range o.shaders {
			s.Uniforms["progress"] = p
			s.Refresh()
		}

		if f >= 1.0 {
			o.anim = nil
			if onDone != nil {
				onDone()
			}
		}
	})
	a.Curve = fyne.AnimationLinear
	o.anim = a
	a.Start()
}

// buildDeskStrip stitches every desktop's preview into one tall image for screen
// sw, desk0 at the top, each slice exactly one screen tall. Every preview is scaled
// to fill its slice precisely so all desktops share the screen's resolution and the
// zoom stays smooth right down to the live desktop revealed at the end. Returns nil
// if the screen has no usable size.
func (l *desktop) buildDeskStrip(sw *screenWindow, count int) image.Image {
	if sw.screen == nil {
		return nil
	}
	w, h := sw.screen.Width, sw.screen.Height
	if w <= 0 || h <= 0 {
		return nil
	}

	strip := image.NewRGBA(image.Rect(0, 0, w, count*h))
	for d := 0; d < count; d++ {
		dst := image.Rect(0, d*h, w, (d+1)*h)
		l.drawDeskPreview(strip, dst, sw, d)
	}
	return strip
}

// drawDeskPreview renders desktop id into rect r of strip, scaled to fill exactly.
func (l *desktop) drawDeskPreview(strip *image.RGBA, r image.Rectangle, sw *screenWindow, id int) {
	img := l.deskPreviewImage(sw, id)
	if img == nil {
		return
	}
	xdraw.ApproxBiLinear.Scale(strip, r, img, img.Bounds(), draw.Src, nil)
}

// deskPreviewImage returns the best image of desktop id for screen sw, using exactly
// the sources the cube transition (startDeskCube) uses so the overview matches it:
//
//  1. The current desktop: a live framebuffer capture via captureOpaque - the whole
//     desktop as drawn, including the composited windows AND the bar and widget panel.
//     It is remembered in deskSnapshots (genuine captures only) so a later overview
//     reuses it instead of synthesising.
//  2. A previously shown desktop: its genuine cached capture (full chrome).
//  3. Otherwise: a face synthesised from the compositor's window pixmaps over the
//     wallpaper (no chrome), exactly as the cube's rolling-in face for an unvisited
//     desktop.
//  4. Embedded mode with no compositor and nothing cached: wallpaper plus schematic
//     window rectangles, so the desktop is never blank.
func (l *desktop) deskPreviewImage(sw *screenWindow, id int) image.Image {
	if id == l.desk && sw.win != nil {
		// A blank capture means GL front-buffer readback isn't working in this
		// environment (e.g. tyde nested in Xephyr via "make embed"); fall through to
		// a synthesised face rather than caching or showing a black tile. On a real
		// desktop the capture is the full desktop, windows and chrome included.
		if live := captureOpaque(sw.win.Canvas()); live != nil && !captureIsBlank(live) {
			if sw.deskSnapshots == nil {
				sw.deskSnapshots = make(map[int]image.Image)
			}
			sw.deskSnapshots[id] = live
			return live
		}
	}

	if sw.deskSnapshots != nil {
		if snap := sw.deskSnapshots[id]; snap != nil {
			return snap
		}
	}

	if face := l.synthesizeDeskFace(sw, l.desk, id); face != nil {
		return face
	}

	if sw.screen == nil {
		return nil
	}
	preview := renderWallpaper(sw.screen.Width, sw.screen.Height)
	if preview == nil {
		preview = image.NewRGBA(image.Rect(0, 0, sw.screen.Width, sw.screen.Height))
	}
	l.drawSchematicWindows(preview, sw, id)
	return preview
}

// drawSchematicWindows draws scaled rectangles for the windows on desktop id (or
// pinned windows) onto dst, a screen-sized preview. It is the no-compositor
// fallback - lower fidelity than real pixmaps but enough to show each desktop's
// layout. The geometry matches the pager's normalisation.
func (l *desktop) drawSchematicWindows(dst *image.RGBA, sw *screenWindow, id int) {
	screen := sw.screen
	if screen == nil || screen.Width <= 0 || screen.Height <= 0 {
		return
	}

	scale := screen.CanvasScale()
	w := float32(dst.Bounds().Dx())
	h := float32(dst.Bounds().Dy())
	fill := &image.Uniform{C: theme.Color(theme.ColorNameDisabled)}

	wins := l.wm.Windows()
	// Windows() is top-first; draw back-to-front so upper windows land on top.
	for i := len(wins) - 1; i >= 0; i-- {
		win := wins[i]
		if win.Iconic() || win.Properties().SkipTaskbar() {
			continue
		}
		if win.Desktop() != id && !win.Pinned() {
			continue
		}

		x := win.Position().X * scale / float32(screen.Width) * w
		y := win.Position().Y * scale / float32(screen.Height) * h
		ww := win.Size().Width * scale / float32(screen.Width) * w
		hh := win.Size().Height * scale / float32(screen.Height) * h
		rect := image.Rect(int(x), int(y), int(x+ww), int(y+hh)).Intersect(dst.Bounds())
		if rect.Empty() {
			continue
		}
		draw.Draw(dst, rect, fill, image.Point{}, draw.Over)
	}
}

const (
	// overviewGap is the space between desktop cells in the overview grid, measured
	// in desktop-height units. overviewMargin leaves a little room around the grid
	// at full zoom-out. Both MUST match the constants of the same value baked into
	// overviewshader.go so the interactive panels line up with the rendered grid.
	overviewGap    = float32(0.06)
	overviewMargin = float32(1.08)
)

// overviewGridCols returns the column count for a count-desktop grid: as close to
// square as the count allows.
func overviewGridCols(count int) int {
	rows := math.Ceil(math.Sqrt(float64(count)))
	return int(math.Ceil(float64(count) / rows))
}

// overviewPanelGeometry returns the on-screen position and size of the preview for
// desktop index when all count desktops are framed as a grid. It reproduces exactly
// the camera the overview shader settles on at full zoom-out, so the interactive
// panels line up with the rendered desktops.
func overviewPanelGeometry(count, index int, size fyne.Size) (fyne.Position, fyne.Size) {
	cols := overviewGridCols(count)
	rows := int(math.Ceil(float64(count) / float64(cols)))
	aspect := size.Width / size.Height

	pitchX := aspect + overviewGap
	pitchY := 1 + overviewGap
	totalW := float32(cols)*aspect + float32(cols-1)*overviewGap
	totalH := float32(rows) + float32(rows-1)*overviewGap

	zoomH := maxF(totalH*0.5, totalW/(2*aspect)) * overviewMargin
	cx, cy := totalW*0.5, totalH*0.5

	col := index % cols
	row := index / cols
	wx0 := float32(col) * pitchX
	wy0 := float32(row) * pitchY

	// World point -> screen fraction (0..1), inverse of the shader's camera.
	toFrac := func(wx, wy float32) (float32, float32) {
		return ((wx-cx)/(zoomH*aspect) + 1) / 2, ((wy-cy)/zoomH + 1) / 2
	}
	fx0, fy0 := toFrac(wx0, wy0)
	fx1, fy1 := toFrac(wx0+aspect, wy0+1)

	return fyne.NewPos(fx0*size.Width, fy0*size.Height),
		fyne.NewSize((fx1-fx0)*size.Width, (fy1-fy0)*size.Height)
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// overviewContent is the interactive selection layer: a transparent, full-screen,
// focusable widget hosting one tappable panel per desktop. Tapping a panel selects
// that desktop; tapping elsewhere or pressing Escape cancels.
type overviewContent struct {
	widget.BaseWidget
	count, current int
	size           fyne.Size
	onSelect       func(int)
	onCancel       func()
}

func newOverviewContent(count, current int, size fyne.Size, sel func(int), cancel func()) *overviewContent {
	c := &overviewContent{count: count, current: current, size: size, onSelect: sel, onCancel: cancel}
	c.ExtendBaseWidget(c)
	return c
}

func (c *overviewContent) Tapped(*fyne.PointEvent) {
	if c.onCancel != nil {
		c.onCancel()
	}
}

func (c *overviewContent) TypedKey(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeyEscape && c.onCancel != nil {
		c.onCancel()
	}
}

func (c *overviewContent) TypedRune(rune) {}
func (c *overviewContent) FocusGained()   {}
func (c *overviewContent) FocusLost()     {}

func (c *overviewContent) CreateRenderer() fyne.WidgetRenderer {
	objs := make([]fyne.CanvasObject, 0, c.count)
	for d := 0; d < c.count; d++ {
		p := newOverviewPanel(d, d == c.current, c.onSelect)
		pos, sz := overviewPanelGeometry(c.count, d, c.size)
		p.Move(pos)
		p.Resize(sz)
		objs = append(objs, p)
	}
	return widget.NewSimpleRenderer(container.NewWithoutLayout(objs...))
}

// overviewPanel is one desktop's hit area in the overview. It is transparent so the
// shader-rendered desktop shows through; the current desktop gets a highlight border
// and every panel shows its number.
type overviewPanel struct {
	widget.BaseWidget
	index    int
	current  bool
	onSelect func(int)
}

func newOverviewPanel(index int, current bool, sel func(int)) *overviewPanel {
	p := &overviewPanel{index: index, current: current, onSelect: sel}
	p.ExtendBaseWidget(p)
	return p
}

func (p *overviewPanel) Tapped(*fyne.PointEvent) {
	if p.onSelect != nil {
		p.onSelect(p.index)
	}
}

func (p *overviewPanel) CreateRenderer() fyne.WidgetRenderer {
	border := canvas.NewRectangle(color.Transparent)
	if p.current {
		border.StrokeColor = theme.Color(theme.ColorNamePrimary)
		border.StrokeWidth = 3
	}

	label := canvas.NewText(strconv.Itoa(p.index+1), theme.Color(theme.ColorNameForeground))
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = theme.TextHeadingSize()
	label.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewStack(border, container.NewCenter(label))
	return widget.NewSimpleRenderer(content)
}
