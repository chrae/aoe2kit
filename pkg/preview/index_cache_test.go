package preview

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"aoe2kit/pkg/datfile"
)

func TestCompactRenderIndexPreservesRenderedPixels(t *testing.T) {
	game, idx := spriteFixture(t)
	idx.Civs[0].Units[0].Attributes = []datfile.UnitAttribute{{AttributeType: 144, Active: true, Amount: 1}, {AttributeType: 0, Active: true, Amount: 999}}
	idx.Civs = append(idx.Civs, datfile.Civ{Index: 1, Name: "not used for Gaia previews"})
	compact := compactRenderIndex(idx)
	if len(compact.Civs) != 1 || len(compact.Civs[0].Units[0].Attributes) != 1 {
		t.Fatal("non-render fields retained")
	}
	for _, opts := range []Options{{Scale: 13}, {Scale: 26, Crop: &Crop{X1: 3, Y1: 3, X2: 7, Y2: 7}, Shadows: &ShadowOptions{}}} {
		a, ar, err := RenderSprites(spriteSnapshot(), opts, idx, game)
		if err != nil {
			t.Fatal(err)
		}
		b, br, err := RenderSprites(spriteSnapshot(), opts, compact, game)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a.Pix, b.Pix) || !reflect.DeepEqual(ar, br) {
			t.Fatal("compact index changed rendering")
		}
	}
}

func TestRenderIndexCacheKeyAndCorruption(t *testing.T) {
	dir := t.TempDir()
	dat := filepath.Join(dir, "source.dat")
	if err := os.WriteFile(dat, []byte("opaque source for warm-cache test"), 0600); err != nil {
		t.Fatal(err)
	}
	key, err := renderDATKey(dat)
	if err != nil {
		t.Fatal(err)
	}
	_, idx := spriteFixture(t)
	cachePath := filepath.Join(dir, "render-index-v1-"+key+".gob")
	f, err := os.Create(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	err = gob.NewEncoder(f).Encode(renderIndexCache{renderIndexSchema, key, compactRenderIndex(idx)})
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadRenderIndex(dat, dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != dat || got.Civs[0].Units[0].Name != "Synthetic" {
		t.Fatal("warm index differs")
	}
	if _, err := readRenderIndex(cachePath, "wrong key"); err == nil {
		t.Fatal("wrong source accepted")
	}
	if err := os.WriteFile(dat, []byte("source changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRenderIndex(dat, dir); err == nil {
		t.Fatal("changed source reused stale cache")
	}
	if err := os.WriteFile(cachePath, []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readRenderIndex(cachePath, key); err == nil {
		t.Fatal("corrupt cache accepted")
	}
}
