package preview

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestBlendMaskSelection(t *testing.T) {
	for bits, want := range map[uint8][]int{
		0: nil, 1: {18}, 2: {6}, 4: {16}, 8: {3}, 16: {17}, 32: {10}, 64: {19}, 128: {15},
		0xaa: {30}, 0xff: {30}, 0x22: {20}, 0x88: {21}, 0x0a: {25}, 0xa0: {22},
		0x82: {23}, 0x28: {24}, 0x2a: {26}, 0xa8: {27}, 0xa2: {28}, 0x8a: {29},
		3: {6}, 0x12: {6, 17}, 0x55: {18, 16, 17, 19},
	} {
		ids, n := blendMaskTiles(bits, 2, 3)
		if n != len(want) {
			t.Fatalf("mask %02x: %v want %v", bits, ids[:n], want)
		}
		for i, id := range want {
			if ids[i] != id {
				t.Fatalf("mask %02x: %v want %v", bits, ids[:n], want)
			}
		}
	}
	for bits := 0; bits < 256; bits++ {
		ids, n := blendMaskTiles(uint8(bits), 7, 9)
		if (bits == 0) != (n == 0) {
			t.Fatal("lost influence", bits)
		}
		seen := map[int]bool{}
		for _, id := range ids[:n] {
			if id < 0 || id >= 31 || seen[id] {
				t.Fatal("invalid or duplicate mask", bits, id)
			}
			seen[id] = true
		}
	}
	for _, pair := range [][3]int32{{0, 3, 1}, {0, 5, 6}, {0, 7, 4}, {2, 6, 1}, {3, 4, 7}, {4, 4, 7}, {3, 3, 0}, {1, 7, 3}, {1, 2, 3}} {
		if blendMode(pair[0], pair[1]) != int(pair[2]) || blendMode(pair[1], pair[0]) != int(pair[2]) {
			t.Fatal("pair mapping", pair)
		}
	}
	if blendMode(8, 0) != -1 || blendMode(-1, 0) != -1 {
		t.Fatal("guessed unknown pair")
	}
}

func TestBlendMaskSampling(t *testing.T) {
	m := image.NewGray(image.Rect(0, 0, 97, 49))
	for y := 0; y < 49; y++ {
		for x := 0; x < 97; x++ {
			m.SetGray(x, y, color.Gray{128})
		}
	}
	// Four scenario corners project to the four diamond tips. Stored weight
	// zero replaces the base; full weight retains it, independent of projection.
	for _, p := range []image.Point{{0, 24}, {48, 0}, {96, 24}, {48, 48}} {
		m.SetGray(p.X, p.Y, color.Gray{255})
	}
	for _, p := range [][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}} {
		if blendAlpha(m, p[0], p[1]) != 0 {
			t.Fatal("diamond tip sampling", p)
		}
	}
	m.SetGray(48, 24, color.Gray{0})
	if blendAlpha(m, .5, .5) != 1 {
		t.Fatal("stored weight polarity")
	}
}

func TestWaterTransitionsShareEdgesAndCorners(t *testing.T) {
	for _, kind := range []int32{0, 2, 3, 4} {
		mats := map[int]terrainMaterial{
			0: {blendType: 3, priority: 180, fallback: color.RGBA{0, 40, 200, 255}},
			1: {blendType: kind, priority: 1, fallback: color.RGBA{220, 180, 80, 255}},
		}
		tiles := make([]Tile, 16)
		for i := range tiles {
			tiles[i] = Tile{TerrainID: (i + i/4) % 2}
		}
		opts := TerrainTextureOptions{Repeat: 10, TextureRepeat: 9}
		near := func(a, b color.RGBA) {
			t.Helper()
			for _, d := range []int{int(a.R) - int(b.R), int(a.G) - int(b.G), int(a.B) - int(b.B)} {
				if math.Abs(float64(d)) > 1 {
					t.Fatalf("type%d seam: %v != %v", kind, a, b)
				}
			}
		}
		for edge := 1; edge < 4; edge++ {
			for p := 1; p < 80; p++ {
				x, y := float64(edge), float64(p)/20
				near(terrainEdgePixel(tiles, 4, 4, mats, x-1e-8, y, opts), terrainEdgePixel(tiles, 4, 4, mats, x+1e-8, y, opts))
				near(terrainEdgePixel(tiles, 4, 4, mats, y, x-1e-8, opts), terrainEdgePixel(tiles, 4, 4, mats, y, x+1e-8, opts))
			}
		}
		uniform := make([]Tile, 16)
		if got := terrainEdgePixel(uniform, 4, 4, mats, .5, .5, opts); got != mats[0].fallback {
			t.Fatal("uniform water changed", got)
		}
		for id, m := range mats {
			m.priority = 500 - m.priority
			mats[id] = m
		}
		a := terrainEdgePixel(tiles, 4, 4, mats, 1, .5, opts)
		if a != (color.RGBA{110, 110, 140, 255}) {
			t.Fatal("priority changed boundary mixture", a)
		}
	}
}

func TestShoreTransitionsToLandWithoutMaskSeams(t *testing.T) {
	mats := map[int]terrainMaterial{0: {blendType: 2, fallback: color.RGBA{220, 180, 80, 255}}, 1: {blendType: 0, fallback: color.RGBA{20, 80, 20, 255}}}
	tiles := []Tile{{TerrainID: 0}, {TerrainID: 1}}
	for _, x := range []float64{1 - 1e-8, 1, 1 + 1e-8} {
		got := terrainEdgePixel(tiles, 2, 1, mats, x, .5, TerrainTextureOptions{})
		if got != (color.RGBA{120, 130, 50, 255}) {
			t.Fatal("shore/land seam", x, got)
		}
	}
}

func TestWaterTransitionLeavesDryLandPolicyAlone(t *testing.T) {
	mats := map[int]terrainMaterial{0: {blendType: 0, fallback: color.RGBA{255, 0, 0, 255}}, 1: {blendType: 5, fallback: color.RGBA{0, 0, 255, 255}}}
	tiles := []Tile{{TerrainID: 0}, {TerrainID: 1}}
	for x := 0.0; x < 2; x += .01 {
		_, w := waterTransitionPixel(tiles, 2, 1, mats, x, .5, TerrainTextureOptions{})
		if w != 0 {
			t.Fatal("dry land entered water policy", w)
		}
	}
}

func TestWaterBaseWithBeachOverlayStillBlends(t *testing.T) {
	mats := map[int]terrainMaterial{
		0: {blendType: 3, fallback: color.RGBA{0, 40, 200, 255}},
		1: {blendType: 2, fallback: color.RGBA{220, 180, 80, 255}},
	}
	top := 1
	// An unavailable beach mask retains the visible water base; classifying
	// only its top layer used to leave this boundary as a hard sand diamond.
	tiles := []Tile{{TerrainID: 0, Layer: &top}, {TerrainID: 1}}
	opts := TerrainTextureOptions{Repeat: 10, TextureRepeat: 9}
	for _, x := range []float64{1 - 1e-8, 1, 1 + 1e-8} {
		got := terrainEdgePixel(tiles, 2, 1, mats, x, .5, opts)
		if got != (color.RGBA{110, 110, 140, 255}) {
			t.Fatal("layered water seam", x, got)
		}
	}
}
