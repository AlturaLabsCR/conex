package handlers

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"

	cardtemplates "app/templates/cards"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	siteCardWidth  = 1200
	siteCardHeight = 630
)

func renderSiteCardPNG(w io.Writer, name string, url string, footer string, nameLines []cardtemplates.TextLine, tags []cardtemplates.Tag) error {
	urlFace, err := siteCardBoldFace(38)
	if err != nil {
		return err
	}
	titleFace, err := siteCardDisplayFace(128)
	if err != nil {
		return err
	}
	tagFace, err := siteCardBoldFace(40)
	if err != nil {
		return err
	}
	footerFace, err := siteCardRegularFace(30)
	if err != nil {
		return err
	}

	img := image.NewRGBA(image.Rect(0, 0, siteCardWidth, siteCardHeight))
	fillRect(img, image.Rect(0, 0, siteCardWidth, siteCardHeight), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	fillRect(img, image.Rect(0, 0, siteCardWidth, 12), color.RGBA{R: 82, G: 182, B: 154, A: 255})

	drawText(img, url, 96, 122, urlFace, color.RGBA{R: 0, G: 73, B: 82, A: 255})

	for _, line := range nameLines {
		x, _ := strconv.Atoi(line.X)
		y, _ := strconv.Atoi(line.Y)
		drawText(img, line.Text, x, y, titleFace, color.RGBA{R: 17, G: 17, B: 17, A: 255})
	}

	for _, tag := range tags {
		x, _ := strconv.Atoi(tag.X)
		y, _ := strconv.Atoi(tag.Y)
		width, _ := strconv.Atoi(tag.Width)
		if width <= 0 {
			continue
		}
		fillRoundedRect(img, image.Rect(x, y, x+width, y+76), 38, parseSiteCardColor(tag.Background, color.RGBA{R: 31, G: 124, B: 107, A: 255}))
		drawText(img, "#"+tag.Label, x+28, y+50, tagFace, parseSiteCardColor(tag.TextColor, color.RGBA{R: 213, G: 250, B: 243, A: 255}))
	}

	drawText(img, footer, 96, 552, footerFace, color.RGBA{R: 102, G: 102, B: 102, A: 255})
	return png.Encode(w, img)
}

func fillRect(img *image.RGBA, rect image.Rectangle, c color.RGBA) {
	draw.Draw(img, rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func fillRoundedRect(img *image.RGBA, rect image.Rectangle, radius int, c color.RGBA) {
	r2 := radius * radius
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			dx, dy := 0, 0
			if x < rect.Min.X+radius {
				dx = rect.Min.X + radius - x
			} else if x >= rect.Max.X-radius {
				dx = x - (rect.Max.X - radius - 1)
			}
			if y < rect.Min.Y+radius {
				dy = rect.Min.Y + radius - y
			} else if y >= rect.Max.Y-radius {
				dy = y - (rect.Max.Y - radius - 1)
			}
			if dx == 0 || dy == 0 || dx*dx+dy*dy <= r2 {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func parseSiteCardColor(value string, fallback color.RGBA) color.RGBA {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "hsl(") || !strings.HasSuffix(value, ")") {
		return fallback
	}

	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(value, "hsl("), ")"), ",")
	if len(parts) != 3 {
		return fallback
	}

	h, errH := strconv.Atoi(strings.TrimSpace(parts[0]))
	s, errS := parsePercent(parts[1])
	l, errL := parsePercent(parts[2])
	if errH != nil || errS != nil || errL != nil {
		return fallback
	}

	return hslToRGB(float64((h%360+360)%360), s, l)
}

func parsePercent(value string) (float64, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(value), "%"), 64)
	return parsed / 100, err
}

func hslToRGB(h float64, s float64, l float64) color.RGBA {
	c := (1 - absFloat(2*l-1)) * s
	x := c * (1 - absFloat(math.Mod(h/60, 2)-1))
	m := l - c/2

	r, g, b := 0.0, 0.0, 0.0
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func drawText(img *image.RGBA, text string, x int, y int, face font.Face, c color.RGBA) {
	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

func siteCardBoldFace(size float64) (font.Face, error) {
	return siteCardFace(siteCardBoldFont, size)
}

func siteCardDisplayFace(size float64) (font.Face, error) {
	return siteCardFace(siteCardDisplayFont, size)
}

func siteCardRegularFace(size float64) (font.Face, error) {
	return siteCardFace(siteCardRegularFont, size)
}

func siteCardFace(source *siteCardFontSource, size float64) (font.Face, error) {
	ttf, err := source.font()
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	return face, nil
}

type siteCardFontSource struct {
	data []byte
	once sync.Once
	ttf  *opentype.Font
	err  error
}

func (s *siteCardFontSource) font() (*opentype.Font, error) {
	s.once.Do(func() {
		s.ttf, s.err = opentype.Parse(s.data)
		if s.err != nil {
			s.err = fmt.Errorf("parse site card font: %w", s.err)
		}
	})
	return s.ttf, s.err
}

var (
	//go:embed assets/fonts/Inter-Bold.ttf
	siteCardInterBold []byte

	//go:embed assets/fonts/Inter-Regular.ttf
	siteCardInterRegular []byte

	//go:embed assets/fonts/InterDisplay-ExtraBold.ttf
	siteCardInterDisplayExtraBold []byte

	siteCardBoldFont    = &siteCardFontSource{data: siteCardInterBold}
	siteCardDisplayFont = &siteCardFontSource{data: siteCardInterDisplayExtraBold}
	siteCardRegularFont = &siteCardFontSource{data: siteCardInterRegular}
)
