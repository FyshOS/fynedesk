//go:build linux || openbsd || freebsd || netbsd
// +build linux openbsd freebsd netbsd

package composit

import (
	"image"
	"image/color"
	"testing"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/stretchr/testify/assert"
)

// convertNaive is a straightforward, independent reference implementation of
// the BGRA->premultiplied-RGBA conversion, used as a correctness oracle and as
// a benchmark baseline for the optimized convertZPixmap.
func convertNaive(pix, data []byte, argb bool) {
	n := len(pix)
	if len(data) < n {
		n = len(data)
	}
	n -= n % 4
	for i := 0; i < n; i += 4 {
		if argb {
			a := uint32(data[i+3])
			pix[i] = uint8((uint32(data[i+2])*a + 127) / 255)
			pix[i+1] = uint8((uint32(data[i+1])*a + 127) / 255)
			pix[i+2] = uint8((uint32(data[i])*a + 127) / 255)
			pix[i+3] = data[i+3]
		} else {
			pix[i] = data[i+2]
			pix[i+1] = data[i+1]
			pix[i+2] = data[i]
			pix[i+3] = 0xff
		}
	}
}

func sampleZPixmap(w, h int) []byte {
	data := make([]byte, w*h*4)
	for i := range data {
		data[i] = byte(i*7 + 3) // deterministic, varies across all 4 bytes
	}
	return data
}

func TestConvertZPixmapMatchesNaive(t *testing.T) {
	for _, argb := range []bool{true, false} {
		src := sampleZPixmap(64, 48)
		want := make([]byte, len(src))
		got := make([]byte, len(src))

		convertNaive(want, src, argb)
		convertZPixmap(got, src, argb)

		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("argb=%v: byte %d = %d, want %d", argb, i, got[i], want[i])
			}
		}
	}
}

func TestConvertZPixmapShortBuffers(t *testing.T) {
	// dst smaller than src: only dst-many whole pixels are converted, no panic.
	src := sampleZPixmap(10, 1)
	dst := make([]byte, 5*4)
	convertZPixmap(dst, src, false)
	for i := 0; i < len(dst); i += 4 {
		if dst[i] != src[i+2] || dst[i+3] != 0xff {
			t.Fatalf("pixel %d not converted correctly", i/4)
		}
	}

	// Non-pixel-multiple tail is ignored rather than read out of range.
	convertZPixmap(make([]byte, 7), sampleZPixmap(2, 1), true) // must not panic
}

func benchConvert(b *testing.B, fn func(dst, src []byte, argb bool)) {
	const w, h = 1920, 1080 // a typical maximised Chrome window
	src := sampleZPixmap(w, h)
	dst := make([]byte, len(src))
	b.SetBytes(int64(len(src)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(dst, src, false) // opaque path — the common case (e.g. Chrome)
	}
}

func BenchmarkConvertNaive(b *testing.B)   { benchConvert(b, convertNaive) }
func BenchmarkConvertZPixmap(b *testing.B) { benchConvert(b, convertZPixmap) }

func TestConvertZPixmapRegion(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 6, 5))
	before := color.RGBA{R: 1, G: 2, B: 3, A: 4}
	for i := 0; i < len(img.Pix); i += 4 {
		img.SetRGBA(i/4%6, i/4/6, before)
	}

	r := image.Rect(2, 1, 5, 3)
	src := sampleZPixmap(r.Dx(), r.Dy())
	want := make([]byte, len(src))
	convertNaive(want, src, false)
	convertZPixmapRegion(img, r, src, false)

	for y := 0; y < 5; y++ {
		for x := 0; x < 6; x++ {
			if !image.Pt(x, y).In(r) {
				assert.Equal(t, before, img.RGBAAt(x, y), "pixel %d,%d outside the region", x, y)
				continue
			}
			i := ((y-r.Min.Y)*r.Dx() + x - r.Min.X) * 4
			assert.Equal(t, color.RGBA{R: want[i], G: want[i+1], B: want[i+2], A: want[i+3]}, img.RGBAAt(x, y),
				"pixel %d,%d", x, y)
		}
	}
}

func TestCopyRegion(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 6, 5))
	for i := range src.Pix {
		src.Pix[i] = byte(i)
	}
	dst := image.NewRGBA(src.Rect)

	r := image.Rect(1, 2, 4, 5)
	copyRegion(dst, src, r)
	for y := 0; y < 5; y++ {
		for x := 0; x < 6; x++ {
			want := color.RGBA{}
			if image.Pt(x, y).In(r) {
				want = src.RGBAAt(x, y)
			}
			assert.Equal(t, want, dst.RGBAAt(x, y), "pixel %d,%d", x, y)
		}
	}

	copyRegion(dst, src, image.Rectangle{}) // nothing to copy, must not panic
}

func TestDamagedRegion(t *testing.T) {
	area := xproto.Rectangle{X: 10, Y: 20, Width: 30, Height: 40}
	assert.Equal(t, image.Rect(10, 20, 40, 60), damagedRegion(area, 0))
	assert.Equal(t, image.Rect(12, 22, 42, 62), damagedRegion(area, 2))
}
