package captcha

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/rotate"

	"github.com/mlogclub/simple/common/strs"
	"github.com/spf13/cast"
)

// Generate creates a fresh captcha with programmatic random backgrounds.
// Each call produces a unique, non-repeating background.
func Generate() (*CaptchaData, error) {
	// Build a new captcha instance with freshly generated backgrounds.
	// This guarantees infinite variety — no fixed image library.
	bgImages := generateBackgrounds(6)

	builder := rotate.NewBuilder(
		rotate.WithRangeAnglePos([]option.RangeVal{
			{Min: 20, Max: 330},
		}),
	)
	builder.SetResources(rotate.WithImages(bgImages))
	capt := builder.Make()

	data, err := capt.Generate()
	if err != nil {
		return nil, err
	}

	imageBase64, err := data.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}
	thumbBase64, err := data.GetThumbImage().ToBase64()
	if err != nil {
		return nil, err
	}

	id := strs.UUID()
	Set(id, data.GetData())

	return &CaptchaData{
		Id:          id,
		ImageBase64: imageBase64,
		ThumbBase64: thumbBase64,
		ThumbSize:   data.GetData().Width,
	}, nil
}

// Verify checks the captcha answer and destroys the captcha on success to
// prevent replay attacks. A failed attempt does not destroy the captcha.
func Verify(captchaId string, captchaCode string) bool {
	data := Get(captchaId)
	if data == nil {
		return false
	}
	angle := cast.ToFloat64(captchaCode)
	ok := rotate.CheckAngle(int64(angle), int64(data.Angle), 2)
	if ok {
		// Destroy after successful verification — one-time use.
		Delete(captchaId)
	}
	return ok
}

// generateBackgrounds creates multiple programmatic noise background images.
func generateBackgrounds(count int) []image.Image {
	images := make([]image.Image, count)
	for i := range images {
		images[i] = generateBackground()
	}
	return images
}

// generateBackground creates a single programmatic noise image.
// Each call produces a unique randomized pattern with polygons, lines, and noise.
func generateBackground() image.Image {
	const w, h = 320, 240
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	// Random gradient base
	baseR := uint8(rand.Intn(80) + 60)  // 60-139
	baseG := uint8(rand.Intn(80) + 80)  // 80-159
	baseB := uint8(rand.Intn(80) + 100) // 100-179

	for y := range h {
		for x := range w {
			r := baseR + uint8(rand.Intn(40))
			g := baseG + uint8(rand.Intn(40))
			b := baseB + uint8(rand.Intn(40))
			img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Random polygons
	polygonCount := rand.Intn(8) + 3 // 3-10 polygons
	for range polygonCount {
		drawRandomPolygon(img, w, h)
	}

	// Random lines
	lineCount := rand.Intn(20) + 10 // 10-29 lines
	for range lineCount {
		drawRandomLine(img, w, h)
	}

	// Random dots/noise
	dotCount := rand.Intn(200) + 100 // 100-299 dots
	for range dotCount {
		x := rand.Intn(w)
		y := rand.Intn(h)
		r := uint8(rand.Intn(256))
		g := uint8(rand.Intn(256))
		b := uint8(rand.Intn(256))
		img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
	}

	return img
}

func drawRandomPolygon(img *image.NRGBA, maxW, maxH int) {
	vertices := rand.Intn(4) + 3 // 3-6 vertices
	pts := make([]image.Point, vertices)
	for i := range vertices {
		pts[i] = image.Point{X: rand.Intn(maxW), Y: rand.Intn(maxH)}
	}
	r := uint8(rand.Intn(120))
	g := uint8(rand.Intn(120))
	b := uint8(rand.Intn(120))
	a := uint8(rand.Intn(80) + 40) // semi-transparent

	// Draw filled polygon (simple scanline approach)
	minX, minY := maxW, maxH
	maxX, maxY := 0, 0
	for _, pt := range pts {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInPolygon(pts, x, y) {
				orig := img.NRGBAAt(x, y)
				img.Set(x, y, blendNRGBA(orig, r, g, b, a))
			}
		}
	}
}

func drawRandomLine(img *image.NRGBA, maxW, maxH int) {
	x1 := rand.Intn(maxW)
	y1 := rand.Intn(maxH)
	x2 := rand.Intn(maxW)
	y2 := rand.Intn(maxH)
	r := uint8(rand.Intn(200))
	g := uint8(rand.Intn(200))
	b := uint8(rand.Intn(200))
	a := uint8(rand.Intn(100) + 60)

	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		if x1 >= 0 && x1 < maxW && y1 >= 0 && y1 < maxH {
			orig := img.NRGBAAt(x1, y1)
			img.Set(x1, y1, blendNRGBA(orig, r, g, b, a))
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func pointInPolygon(pts []image.Point, x, y int) bool {
	inside := false
	j := len(pts) - 1
	for i := range pts {
		if (pts[i].Y > y) != (pts[j].Y > y) &&
			x < (pts[j].X-pts[i].X)*(y-pts[i].Y)/(pts[j].Y-pts[i].Y)+pts[i].X {
			inside = !inside
		}
		j = i
	}
	return inside
}

func blendNRGBA(orig color.NRGBA, r, g, b, a uint8) color.NRGBA {
	alpha := uint16(a)
	invAlpha := uint16(255 - a)
	return color.NRGBA{
		R: uint8((uint16(r)*alpha + uint16(orig.R)*invAlpha) / 255),
		G: uint8((uint16(g)*alpha + uint16(orig.G)*invAlpha) / 255),
		B: uint8((uint16(b)*alpha + uint16(orig.B)*invAlpha) / 255),
		A: 255,
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
