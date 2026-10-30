package preview

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"aoe2kit/pkg/datfile"
)

type TileLevelExtent struct {
	Z      int `json:"z"`
	Scale  int `json:"scale"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type TileBatchOptions struct {
	Dir               string
	TileSize          int
	Scales            []int
	Sprites           bool
	Strict            bool
	Mood              *Mood
	MoodModel         MoodModel
	MoodSource        string
	CanvasBudgetBytes int64
}

type TileBatchReport struct {
	Metrics      *RenderMetrics    `json:"metrics,omitempty"`
	Schema       string            `json:"schema"`
	OK           bool              `json:"ok"`
	Levels       []TileLevelExtent `json:"levels"`
	TileSize     int               `json:"tile_size"`
	Rendered     int               `json:"rendered"`
	Skipped      int               `json:"skipped"`
	Milliseconds float64           `json:"milliseconds"`
	RenderModes  map[int]string    `json:"render_modes"`
}

// RenderTiles renders each level whole when it fits the canvas budget, or in
// bounded horizontal strips otherwise, then chops fixed-size padded tiles.
func RenderTiles(s Snapshot, opts Options, idx *datfile.Index, gameDir string, batch TileBatchOptions) (report TileBatchReport, err error) {
	report = TileBatchReport{Schema: "aoe2kit.preview.tiles.v1", RenderModes: map[int]string{}, Metrics: opts.Metrics}
	started := time.Now()
	defer func() { report.Milliseconds = float64(time.Since(started)) / float64(time.Millisecond) }()
	if err := validateSnapshot(s); err != nil {
		return report, err
	}
	if batch.Dir == "" {
		return report, fmt.Errorf("tiles directory is required")
	}
	if opts.Crop != nil || opts.Window != nil {
		return report, fmt.Errorf("batch tiles require the full-map frame")
	}
	if batch.TileSize == 0 {
		batch.TileSize = 256
	}
	if batch.TileSize < 1 || batch.TileSize > 2048 {
		return report, fmt.Errorf("tile size must be between 1 and 2048")
	}
	if batch.CanvasBudgetBytes == 0 {
		batch.CanvasBudgetBytes = 32 * 1024 * 1024
	}
	if batch.CanvasBudgetBytes < int64(batch.TileSize)*int64(batch.TileSize)*8 {
		return report, fmt.Errorf("canvas budget must fit at least one tile (8 estimated bytes per pixel)")
	}
	if len(batch.Scales) == 0 {
		batch.Scales = []int{8, 16, 32, 64}
	}
	if len(batch.Scales) > 16 {
		return report, fmt.Errorf("at most 16 scales are allowed")
	}
	if batch.Mood != nil {
		if err := batch.MoodModel.Validate(); err != nil {
			return report, err
		}
	}
	seen := map[int]bool{}
	for z, scale := range batch.Scales {
		if scale < 1 || scale > 128 || seen[scale] {
			return report, fmt.Errorf("scales must be distinct integers between 1 and 128")
		}
		seen[scale] = true
		extent, err := SnapshotImageExtent(s, scale, opts.Projection, opts.Elevation, opts.StoredObjectHeights)
		if err != nil {
			return report, err
		}
		report.Levels = append(report.Levels, TileLevelExtent{Z: z, Scale: scale, Width: extent.Width, Height: extent.Height})
	}
	report.TileSize = batch.TileSize
	if err := os.MkdirAll(batch.Dir, 0755); err != nil {
		return report, err
	}
	extentPath := filepath.Join(batch.Dir, "extent.json")
	writeExtent := true
	if data, err := os.ReadFile(extentPath); err == nil {
		var old []TileLevelExtent
		if json.Unmarshal(data, &old) != nil || !reflect.DeepEqual(old, report.Levels) {
			return report, fmt.Errorf("existing tile extents differ; use a fresh directory")
		}
		writeExtent = false
	} else if !os.IsNotExist(err) {
		return report, err
	}
	if writeExtent {
		if err := writeTileArtifact(extentPath, func(f *os.File) error { return json.NewEncoder(f).Encode(report.Levels) }); err != nil {
			return report, err
		}
	}
	if idx == nil {
		idx = opts.Dat
	}
	opts.Dat = idx
	render := func() (*image.RGBA, Report, error) {
		if batch.Sprites {
			return RenderSprites(s, opts, idx, gameDir)
		}
		return RenderFlat(s, opts)
	}
	for _, level := range report.Levels {
		opts.Scale = level.Scale
		dir := filepath.Join(batch.Dir, fmt.Sprint(level.Z))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return report, err
		}
		var background color.RGBA
		originReady := false
		pending := map[image.Point]string{}
		for y := 0; y < level.Height; y += batch.TileSize {
			for x := 0; x < level.Width; x += batch.TileSize {
				path := filepath.Join(dir, fmt.Sprintf("%d_%d.png", x/batch.TileSize, y/batch.TileSize))
				if info, err := os.Stat(path); err == nil {
					if !info.Mode().IsRegular() {
						return report, fmt.Errorf("tile path is not a regular file: %s", path)
					}
					f, err := os.Open(path)
					if err != nil {
						return report, err
					}
					config, err := png.DecodeConfig(f)
					f.Close()
					if err != nil || config.Width != batch.TileSize || config.Height != batch.TileSize {
						return report, fmt.Errorf("existing tile is invalid or has a different tile size: %s", path)
					}
					report.Skipped++
					continue
				} else if !os.IsNotExist(err) {
					return report, err
				}
				pending[image.Pt(x, y)] = path
			}
		}
		if len(pending) == 0 {
			report.RenderModes[level.Z] = "cached"
			continue
		}
		regions, whole := tileRenderRegions(level, batch.TileSize, batch.CanvasBudgetBytes)
		report.RenderModes[level.Z] = "strips"
		if whole {
			report.RenderModes[level.Z] = "whole"
		}
		for _, region := range regions {
			needed := false
			for point := range pending {
				if point.In(region) {
					needed = true
					break
				}
			}
			if !needed {
				continue
			}
			opts.Window = nil
			if !whole {
				opts.Window = &PixelWindow{X: region.Min.X, Y: region.Min.Y, Width: region.Dx(), Height: region.Dy()}
			}
			img, r, err := render()
			if err != nil {
				return report, err
			}
			if batch.Strict {
				if err := CheckSpriteCompleteness(r); err != nil {
					return report, err
				}
			}
			if batch.Mood != nil {
				if !originReady {
					if image.Pt(0, 0).In(img.Bounds()) {
						background = img.RGBAAt(0, 0)
					} else {
						opts.Window = &PixelWindow{Width: 1, Height: 1}
						origin, _, err := render()
						if err != nil {
							return report, err
						}
						background = origin.RGBAAt(0, 0)
					}
					originReady = true
				}
				moodStarted := opts.Metrics.start()
				_, moodErr := ApplyMoodModelBackground(img, *batch.Mood, batch.MoodSource, batch.MoodModel, background)
				opts.Metrics.Record("mood", moodStarted)
				if moodErr != nil {
					return report, moodErr
				}
			}
			for y := region.Min.Y; y < region.Max.Y; y += batch.TileSize {
				for x := region.Min.X; x < region.Max.X; x += batch.TileSize {
					point := image.Pt(x, y)
					path, ok := pending[point]
					if !ok {
						continue
					}
					tile := image.NewRGBA(image.Rect(0, 0, batch.TileSize, batch.TileSize))
					draw.Draw(tile, tile.Bounds(), img, point, draw.Src)
					pngStarted := opts.Metrics.start()
					err := writeTileArtifact(path, func(f *os.File) error { return png.Encode(f, tile) })
					opts.Metrics.Record("png", pngStarted)
					if err != nil {
						return report, err
					}
					delete(pending, point)
					report.Rendered++
				}
			}
		}
	}
	report.OK = true
	report.Milliseconds = float64(time.Since(started)) / float64(time.Millisecond)
	return report, nil
}

func tileRenderRegions(level TileLevelExtent, size int, budget int64) ([]image.Rectangle, bool) {
	full := image.Rect(0, 0, level.Width, level.Height)
	if int64(level.Width)*int64(level.Height)*8 <= budget && int64(level.Width)*int64(level.Height) <= maxImagePixels {
		return []image.Rectangle{full}, true
	}
	// Usually one full-width, tile-row strip. Extremely wide maps split the
	// strip at tile boundaries too, keeping the same byte bound.
	width := min(level.Width, max(size, int(budget/(int64(size)*8))/size*size))
	var regions []image.Rectangle
	for y := 0; y < level.Height; y += size {
		for x := 0; x < level.Width; x += width {
			regions = append(regions, image.Rect(x, y, min(level.Width, x+width), min(level.Height, y+size)))
		}
	}
	return regions, false
}

// Atomic replacement makes a killed render resumable without mistaking its
// half-written image for a completed tile on the next invocation.
func writeTileArtifact(path string, write func(*os.File) error) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".kit-tile-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	err = write(f)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Chmod(f.Name(), 0644); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
