package preview

import (
	"image"
	"image/color"
	"testing"
)

func assertWindowPixels(t *testing.T, whole, window *image.RGBA, rect image.Rectangle) {
	t.Helper()
	if window.Bounds() != rect {
		t.Fatalf("bounds %v != %v", window.Bounds(), rect)
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			want := color.RGBA{}
			if image.Pt(x, y).In(whole.Bounds()) {
				want = whole.RGBAAt(x, y)
			}
			if got := window.RGBAAt(x, y); got != want {
				t.Fatalf("pixel (%d,%d): %v != %v", x, y, got, want)
			}
		}
	}
}

func TestFlatWindowsEqualWhole(t *testing.T) {
	s := spriteSnapshot()
	s.Width, s.Height = 13, 9
	s.Terrain = nil
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			s.Terrain = append(s.Terrain, Tile{X: x, Y: y, TerrainID: (x + y) % 25, Elevation: float64(x % 4)})
		}
	}
	s.Objects = append(s.Objects, Object{X: 7.2, Y: 3.8, Class: "building", Blocking: true, FootprintX: 5, FootprintY: 3, Player: 1})
	for _, projection := range []string{"diamond", "topdown"} {
		for _, scale := range []int{8, 16, 32, 64} {
			opts := Options{Scale: scale, Projection: projection, Elevation: true}
			whole, _, err := RenderFlat(s, opts)
			if err != nil {
				t.Fatal(err)
			}
			extent, err := SnapshotImageExtent(s, scale, projection, opts.Elevation)
			if err != nil || extent.Width != whole.Bounds().Dx() || extent.Height != whole.Bounds().Dy() {
				t.Fatal(extent, err)
			}
			for y := -3; y < whole.Bounds().Max.Y+256; y += 256 {
				for x := -5; x < whole.Bounds().Max.X+256; x += 256 {
					w := PixelWindow{X: x, Y: y, Width: 256, Height: 256}
					opts.Window = &w
					got, r, err := RenderFlat(s, opts)
					if err != nil {
						t.Fatal(err)
					}
					rect, _ := w.rectangle()
					assertWindowPixels(t, whole, got, rect)
					if r.ImageWidth != 256 || r.ImageHeight != 256 || r.Extent.Width != extent.Width {
						t.Fatal(r)
					}
				}
			}
		}
	}
}

func TestSpriteWindowIncludesOutsideAnchor(t *testing.T) {
	game, idx := spriteFixture(t)
	s := spriteSnapshot()
	opts := Options{Scale: 8, SpriteTilePixels: 1, Shadows: &ShadowOptions{}}
	whole, _, err := RenderSprites(s, opts, idx, game)
	if err != nil {
		t.Fatal(err)
	}
	x, y := mapPoint(5, 5, Crop{X2: 10, Y2: 10}, 8, "diamond")
	// The sprite extends left/up from the anchor, which is outside this window.
	opts.Window = &PixelWindow{X: x - 16, Y: y - 20, Width: 12, Height: 12}
	got, _, err := RenderSprites(s, opts, idx, game)
	if err != nil {
		t.Fatal(err)
	}
	rect, _ := opts.Window.rectangle()
	assertWindowPixels(t, whole, got, rect)
	if got.RGBAAt(rect.Min.X, rect.Min.Y).R < 248 {
		t.Fatal("outside-anchor sprite omitted")
	}
}

func TestWindowValidationAndVirtualExtent(t *testing.T) {
	for _, w := range []PixelWindow{{Width: 0, Height: 1}, {Width: 1, Height: -1}, {Width: 1 << 30, Height: 1 << 30}, {X: 1 << 30, Width: 1, Height: 1}} {
		if _, err := w.rectangle(); err == nil {
			t.Fatal("accepted", w)
		}
	}
	extent, err := FullImageExtent(480, 480, 64, "diamond")
	if err != nil || extent.Width != 30721 || extent.Height != 15489 {
		t.Fatal(extent, err)
	}
	s := spriteSnapshot()
	if _, _, err := RenderFlat(s, Options{Crop: &Crop{X2: 5, Y2: 5}, Window: &PixelWindow{Width: 1, Height: 1}}); err == nil {
		t.Fatal("crop/window combination accepted")
	}
	large := Snapshot{Schema: SnapshotSchema, Width: 480, Height: 480}
	for y := 0; y < large.Height; y++ {
		for x := 0; x < large.Width; x++ {
			large.Terrain = append(large.Terrain, Tile{X: x, Y: y})
		}
	}
	if _, _, err := RenderFlat(large, Options{Scale: 64}); err == nil {
		t.Fatal("oversized whole frame accepted")
	}
	if got, _, err := RenderFlat(large, Options{Scale: 64, Window: &PixelWindow{X: 15000, Y: 7500, Width: 32, Height: 32}}); err != nil || got.Bounds().Dx() != 32 {
		t.Fatal("virtual full frame should not count against window budget", err)
	}
}

func TestMoodWindowsUseFullOrigin(t *testing.T) {
	s := spriteSnapshot()
	opts := Options{Scale: 8, Projection: "topdown"}
	whole, _, err := RenderFlat(s, opts)
	if err != nil {
		t.Fatal(err)
	}
	origin := whole.RGBAAt(0, 0)
	m := Mood{Brightness: 1.3, Contrast: 1.1, Saturation: .6}
	model := CurrentMoodModel()
	if _, err := ApplyMoodModel(whole, m, "test", model); err != nil {
		t.Fatal(err)
	}
	// Start on the object, whose color must not become a tile-local background.
	opts.Window = &PixelWindow{X: 40, Y: 40, Width: 60, Height: 60}
	got, _, err := RenderFlat(s, opts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyMoodModelBackground(got, m, "test", model, origin); err != nil {
		t.Fatal(err)
	}
	rect, _ := opts.Window.rectangle()
	assertWindowPixels(t, whole, got, rect)
}
