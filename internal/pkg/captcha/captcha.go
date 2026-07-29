package captcha

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"strings"

	"github.com/spf13/cast"
	"github.com/wenlng/go-captcha/v2/slide"

	"github.com/mlogclub/simple/common/strs"
)

// Generate creates a slide captcha with programmatic random backgrounds and
// puzzle-piece shapes. Every call produces unique images — no fixed asset
// library that an attacker could pre-learn.
func Generate() (*CaptchaData, error) {
	bgImages := generateBackgrounds(6)
	tileGraphs := generateTileGraphs(4)

	builder := slide.NewBuilder(
		slide.WithEnableGraphVerticalRandom(true),
	)
	builder.SetResources(
		slide.WithGraphImages(tileGraphs),
		slide.WithBackgrounds(bgImages),
	)
	capt := builder.Make()

	data, err := capt.Generate()
	if err != nil {
		return nil, err
	}

	imageBase64, err := data.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}
	thumbBase64, err := data.GetTileImage().ToBase64()
	if err != nil {
		return nil, err
	}

	block := data.GetData()
	id := strs.UUID()
	Set(id, block)

	return &CaptchaData{
		Id:          id,
		ImageBase64: imageBase64,
		ThumbBase64: thumbBase64,
		ThumbX:      block.TileX,
		ThumbY:      block.TileY,
		ThumbWidth:  block.Width,
		ThumbHeight: block.Height,
	}, nil
}

// Verify checks the captcha answer and destroys the captcha on success to
// prevent replay attacks. A failed attempt does not destroy the captcha.
// captchaCode format: "x,y" — the user's slide end position.
func Verify(captchaId string, captchaCode string) bool {
	data := Get(captchaId)
	if data == nil {
		return false
	}

	parts := strings.Split(captchaCode, ",")
	if len(parts) != 2 {
		return false
	}
	sx := cast.ToInt64(strings.TrimSpace(parts[0]))
	sy := cast.ToInt64(strings.TrimSpace(parts[1]))

	ok := slide.CheckPoint(sx, sy, int64(data.X), int64(data.Y), 5)
	if ok {
		Delete(captchaId)
	}
	return ok
}

// ---------------------------------------------------------------------------
// Programmatic background generation (same approach as the old rotate captcha)
// ---------------------------------------------------------------------------

func generateBackgrounds(count int) []image.Image {
	images := make([]image.Image, count)
	for i := range images {
		images[i] = generateBackground()
	}
	return images
}

func generateBackground() image.Image {
	const w, h = 300, 220
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	baseR := uint8(rand.Intn(80) + 50)
	baseG := uint8(rand.Intn(80) + 70)
	baseB := uint8(rand.Intn(80) + 90)

	for y := range h {
		for x := range w {
			r := baseR + uint8(rand.Intn(50))
			g := baseG + uint8(rand.Intn(50))
			b := baseB + uint8(rand.Intn(50))
			img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Random polygons
	for range rand.Intn(6) + 2 {
		drawRandomPolygon(img, w, h)
	}

	// Random lines
	for range rand.Intn(15) + 5 {
		drawRandomLine(img, w, h)
	}

	// Random dots/noise
	for range rand.Intn(150) + 50 {
		x := rand.Intn(w)
		y := rand.Intn(h)
		r := uint8(rand.Intn(256))
		g := uint8(rand.Intn(256))
		b := uint8(rand.Intn(256))
		img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
	}

	return img
}

// ---------------------------------------------------------------------------
// Programmatic puzzle-piece (tile) generation
// ---------------------------------------------------------------------------

// generateTileGraphs creates multiple puzzle-piece shapes with random
// tab positions and sizes so the captcha is not trivially matchable.
func generateTileGraphs(count int) []*slide.GraphImage {
	graphs := make([]*slide.GraphImage, count)
	for i := range graphs {
		graphs[i] = generateTileGraph()
	}
	return graphs
}

// generateTileGraph creates a single puzzle-piece shape: a rectangle with a
// semi-circular "tab" protruding from one edge. Returns the three required
// images: Overlay (draggable piece), Shadow (notch shadow), Mask (cutout).
func generateTileGraph() *slide.GraphImage {
	const size = 65
	tabR := rand.Intn(10) + 8 // tab radius 8-17

	// Choose a random edge for the tab
	edge := rand.Intn(4) // 0=right, 1=left, 2=bottom, 3=top

	var imgW, imgH int
	var tabCX, tabCY int
	switch edge {
	case 0: // right
		imgW, imgH = size+tabR, size
		tabCX, tabCY = size, size/2
	case 1: // left
		imgW, imgH = size+tabR, size
		tabCX, tabCY = 0, size/2
	case 2: // bottom
		imgW, imgH = size, size+tabR
		tabCX, tabCY = size/2, size
	default: // top
		imgW, imgH = size, size+tabR
		tabCX, tabCY = size/2, 0
	}

	mask := image.NewNRGBA(image.Rect(0, 0, imgW, imgH))

	// Draw filled rectangle
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var mx, my int
			switch edge {
			case 0: // tab on right, rect at left
				mx, my = x, y
			case 1: // tab on left, rect shifted right
				mx, my = x+tabR, y
			case 2: // tab on bottom, rect at top
				mx, my = x, y
			default: // tab on top, rect shifted down
				mx, my = x, y+tabR
			}
			mask.Set(mx, my, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	// Draw semi-circular tab
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			dx := x - tabCX
			dy := y - tabCY
			dist := int(math.Sqrt(float64(dx*dx + dy*dy)))
			if dist <= tabR {
				mask.Set(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}

	// Overlay: near-transparent tint applied on top of the cropped background.
	// The tile already shows the background through the mask; the overlay just
	// adds a faint highlight so the piece is distinguishable from the notch.
	overlay := copyAndTint(mask, color.NRGBA{R: 255, G: 255, B: 255, A: 40})

	// Shadow: darker version for the notch cutout on the main image.
	shadow := copyAndTint(mask, color.NRGBA{R: 0, G: 0, B: 0, A: 100})

	return &slide.GraphImage{
		OverlayImage: overlay,
		ShadowImage:  shadow,
		MaskImage:    mask,
	}
}

// copyAndTint creates a copy of the mask image, replacing every non-
// transparent pixel with the given tint color.
func copyAndTint(src *image.NRGBA, tint color.NRGBA) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.NRGBAAt(x, y)
			if c.A > 0 {
				dst.Set(x, y, tint)
			}
		}
	}
	return dst
}

// ---------------------------------------------------------------------------
// Drawing helpers (from the old rotate captcha)
// ---------------------------------------------------------------------------

func drawRandomPolygon(img *image.NRGBA, maxW, maxH int) {
	vertices := rand.Intn(4) + 3
	pts := make([]image.Point, vertices)
	for i := range vertices {
		pts[i] = image.Point{X: rand.Intn(maxW), Y: rand.Intn(maxH)}
	}
	r := uint8(rand.Intn(120))
	g := uint8(rand.Intn(120))
	b := uint8(rand.Intn(120))
	a := uint8(rand.Intn(80) + 30)

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
	a := uint8(rand.Intn(80) + 50)

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
