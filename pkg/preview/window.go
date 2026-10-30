package preview

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"math"
)

// PixelWindow is a rectangle in the full image's pixel coordinate frame.
type PixelWindow struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// FullImageExtent describes the virtual canvas without allocating its pixels.
// MapOrigin is tile coordinate (0,0), not the image's top-left origin (0,0).
func FullImageExtent(width, height, scale int, projection string) (ImageExtent, error) {
	projection = defaultProjection(projection)
	if scale == 0 {
		scale = 4
	}
	if width < 1 || height < 1 || width > 480 || height > 480 || scale < 1 || scale > 128 {
		return ImageExtent{}, fmt.Errorf("invalid map dimensions or scale")
	}
	if projection != "diamond" && projection != "topdown" {
		return ImageExtent{}, fmt.Errorf("projection must be topdown or diamond")
	}
	w, h := width*scale, height*scale
	if projection == "diamond" {
		w, h = (width+height)*scale/2+1, (width+height)*scale/4+scale*2+1
	}
	x, y := mapPoint(0, 0, Crop{X2: width, Y2: height}, scale, projection)
	return ImageExtent{Width: w, Height: h, Scale: scale, Projection: projection, MapOriginX: x, MapOriginY: y}, nil
}

func (w PixelWindow) rectangle() (image.Rectangle, error) {
	const limit = 1 << 30
	if w.Width < 1 || w.Height < 1 || w.Width > limit || w.Height > limit || w.X < -limit || w.Y < -limit || w.X > limit-w.Width || w.Y > limit-w.Height || int64(w.Width)*int64(w.Height) > maxImagePixels {
		return image.Rectangle{}, fmt.Errorf("invalid pixel window or window exceeds %d-pixel limit", maxImagePixels)
	}
	return image.Rect(w.X, w.Y, w.X+w.Width, w.Y+w.Height), nil
}

func finishWindow(img *image.RGBA, report *Report, window *PixelWindow) *image.RGBA {
	if window == nil {
		return img
	}
	bounds, _ := window.rectangle()
	if img.Bounds() != bounds {
		padded := image.NewRGBA(bounds)
		draw.Draw(padded, img.Bounds(), img, img.Bounds().Min, draw.Src)
		img = padded
	}
	report.ImageWidth, report.ImageHeight = window.Width, window.Height
	report.PixelSHA256 = fmt.Sprintf("%x", sha256.Sum256(img.Pix))
	return img
}

// windowTerrainCrop is only an asset-selection bound. Projection and texture
// sampling always use the original full-map crop, with a neighbor halo.
func windowTerrainCrop(bounds image.Rectangle, crop Crop, scale int, projection string) Crop {
	if bounds.Empty() {
		return Crop{}
	}
	t := diamondTransform(crop.X2-crop.X1, crop.Y2-crop.Y1, float64(scale), float64(scale), "")
	loX, loY, hiX, hiY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range []image.Point{bounds.Min, bounds.Max, {X: bounds.Min.X, Y: bounds.Max.Y}, {X: bounds.Max.X, Y: bounds.Min.Y}} {
		x, y := float64(p.X)/float64(scale), float64(p.Y)/float64(scale)
		if projection == "diamond" {
			x, y = t.inverse(float64(p.X), float64(p.Y))
		}
		loX, loY, hiX, hiY = math.Min(loX, x), math.Min(loY, y), math.Max(hiX, x), math.Max(hiY, y)
	}
	return Crop{
		X1: max(crop.X1, min(crop.X2, int(math.Floor(loX))+crop.X1-1)),
		Y1: max(crop.Y1, min(crop.Y2, int(math.Floor(loY))+crop.Y1-1)),
		X2: min(crop.X2, max(crop.X1, int(math.Ceil(hiX))+crop.X1+1)),
		Y2: min(crop.Y2, max(crop.Y1, int(math.Ceil(hiY))+crop.Y1+1)),
	}
}
