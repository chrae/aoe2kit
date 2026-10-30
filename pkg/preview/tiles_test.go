package preview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func readTile(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func assertTileMatchesWindow(t *testing.T, tile image.Image, window *image.RGBA) {
	t.Helper()
	if tile.Bounds().Dx() != window.Bounds().Dx() || tile.Bounds().Dy() != window.Bounds().Dy() {
		t.Fatal("tile/window dimensions differ")
	}
	for y := 0; y < tile.Bounds().Dy(); y++ {
		for x := 0; x < tile.Bounds().Dx(); x++ {
			got := color.RGBAModel.Convert(tile.At(x, y)).(color.RGBA)
			want := window.RGBAAt(window.Bounds().Min.X+x, window.Bounds().Min.Y+y)
			if got != want {
				t.Fatalf("pixel%d,%d: %v != %v", x, y, got, want)
			}
		}
	}
}

func assertTileFileMatchesWindow(t *testing.T, path string, window *image.RGBA) {
	t.Helper()
	assertTileMatchesWindow(t, readTile(t, path), window)
	var want bytes.Buffer
	if err := png.Encode(&want, window); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatal("tile/window PNG bytes differ", path)
	}
}

func TestTileBatchWholeStripsAndResume(t *testing.T) {
	game, idx := spriteFixture(t)
	s := spriteSnapshot()
	for _, budget := range []int64{1024 * 1024, 16 * 16 * 8} {
		dir := t.TempDir()
		metrics := &RenderMetrics{}
		opts := Options{SpriteTilePixels: 1, Metrics: metrics, Shadows: &ShadowOptions{}}
		batch := TileBatchOptions{Dir: dir, Scales: []int{8, 16}, TileSize: 16, Sprites: true, CanvasBudgetBytes: budget}
		r, err := RenderTiles(s, opts, idx, game, batch)
		if err != nil || !r.OK {
			t.Fatal(r, err)
		}
		if metrics.Stages["png"].Operations != r.Rendered {
			t.Fatal("each written tile should record one PNG operation", metrics.Stages, r)
		}
		for _, level := range r.Levels {
			if budget == 1024*1024 && r.RenderModes[level.Z] != "whole" {
				t.Fatal(r.RenderModes)
			}
			if budget == 16*16*8 && r.RenderModes[level.Z] != "strips" {
				t.Fatal(r.RenderModes)
			}
			for y := 0; y < level.Height; y += 16 {
				for x := 0; x < level.Width; x += 16 {
					opts.Scale = level.Scale
					opts.Window = &PixelWindow{X: x, Y: y, Width: 16, Height: 16}
					window, _, err := RenderSprites(s, opts, idx, game)
					if err != nil {
						t.Fatal(err)
					}
					assertTileMatchesWindow(t, readTile(t, filepath.Join(dir, fmt.Sprint(level.Z), fmt.Sprintf("%d_%d.png", x/16, y/16))), window)
				}
			}
		}
		data, err := os.ReadFile(filepath.Join(dir, "extent.json"))
		if err != nil {
			t.Fatal(err)
		}
		var extents []map[string]int
		if err := json.Unmarshal(data, &extents); err != nil || len(extents) != 2 || len(extents[0]) != 4 || extents[1]["scale"] != 16 {
			t.Fatal(string(data), err)
		}
		// A complete rerun opens no sprite assets. A missing tile is rebuilt.
		resumeOpts := Options{SpriteTilePixels: 1, Metrics: &RenderMetrics{}, Shadows: &ShadowOptions{}}
		cached, err := RenderTiles(s, resumeOpts, idx, game, batch)
		if err != nil || cached.Rendered != 0 || cached.Skipped != r.Rendered || resumeOpts.Metrics.Stages["sld_load"].Operations != 0 {
			t.Fatal(cached, err)
		}
		if err := os.Remove(filepath.Join(dir, "1", "0_0.png")); err != nil {
			t.Fatal(err)
		}
		partial, err := RenderTiles(s, resumeOpts, idx, game, batch)
		if err != nil || partial.Rendered != 1 || partial.Skipped != r.Rendered-1 {
			t.Fatal(partial, err)
		}
	}
}

func TestTileRenderRegionsUseByteBudget(t *testing.T) {
	for _, scale := range []int{8, 16, 32, 64} {
		e, _ := FullImageExtent(80, 80, scale, "diamond")
		regions, whole := tileRenderRegions(TileLevelExtent{Width: e.Width, Height: e.Height}, 256, 32*1024*1024)
		if whole != (scale != 64) {
			t.Fatal(scale, whole)
		}
		for _, r := range regions {
			if int64(r.Dx())*int64(r.Dy())*8 > 32*1024*1024 {
				t.Fatal("canvas exceeds budget", r)
			}
		}
	}
	regions, whole := tileRenderRegions(TileLevelExtent{Width: 61441, Height: 30977}, 256, 32*1024*1024)
	if whole {
		t.Fatal("large map must use strips")
	}
	for _, r := range regions {
		if int64(r.Dx())*int64(r.Dy())*8 > 32*1024*1024 || r.Min.X%256 != 0 || r.Min.Y%256 != 0 {
			t.Fatal(r)
		}
	}
}

func TestTileBatchMoodMetricsAndResume(t *testing.T) {
	game, idx := spriteFixture(t)
	mood := neutralTestMood("test")
	for _, budget := range []int64{1024 * 1024, 16 * 16 * 8} {
		metrics := &RenderMetrics{}
		batch := TileBatchOptions{Dir: t.TempDir(), Scales: []int{8}, TileSize: 16, Mood: &mood, MoodModel: CurrentMoodModel(), CanvasBudgetBytes: budget}
		r, err := RenderTiles(spriteSnapshot(), Options{Metrics: metrics}, idx, game, batch)
		if err != nil || !r.OK {
			t.Fatal(r, err)
		}
		level := r.Levels[0]
		regions, _ := tileRenderRegions(level, batch.TileSize, budget)
		if metrics.Stages["mood"].Operations != len(regions) || metrics.Stages["png"].Operations != r.Rendered {
			t.Fatal("mood must count regions, PNG must count written tiles", metrics.Stages, r)
		}
		cachedMetrics := &RenderMetrics{}
		cached, err := RenderTiles(spriteSnapshot(), Options{Metrics: cachedMetrics}, idx, game, batch)
		if err != nil || !cached.OK || cached.Rendered != 0 || cachedMetrics.Stages["mood"].Operations != 0 || cachedMetrics.Stages["png"].Operations != 0 {
			t.Fatal("resuming completed tiles must not grade or encode", cached, cachedMetrics.Stages, err)
		}
	}
}

func TestAtomicTileFailureLeavesNoPartialOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0_0.png")
	if err := writeTileArtifact(path, func(f *os.File) error { _, _ = f.WriteString("partial"); return fmt.Errorf("interrupted") }); err == nil {
		t.Fatal("write should fail")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatal(entries, err)
	}
}
