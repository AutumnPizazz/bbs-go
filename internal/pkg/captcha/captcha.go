package captcha

import (
	"strings"

	"github.com/spf13/cast"
	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"

	"github.com/mlogclub/simple/common/strs"
)

// Generate creates a slide captcha using built-in tile and background assets.
func Generate() (*CaptchaData, error) {
	// Load tile (puzzle piece) graphics from embedded assets
	graphs, err := tiles.GetTiles()
	if err != nil {
		return nil, err
	}
	slideGraphs := make([]*slide.GraphImage, 0, len(graphs))
	for _, g := range graphs {
		slideGraphs = append(slideGraphs, &slide.GraphImage{
			OverlayImage: g.OverlayImage,
			ShadowImage:  g.ShadowImage,
			MaskImage:    g.MaskImage,
		})
	}

	// Load background images from embedded assets
	bgImages, err := imagesv2.GetImages()
	if err != nil {
		return nil, err
	}

	builder := slide.NewBuilder(
		slide.WithEnableGraphVerticalRandom(true),
	)
	builder.SetResources(
		slide.WithGraphImages(slideGraphs),
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
		// Destroy after successful verification — one-time use.
		Delete(captchaId)
	}
	return ok
}
