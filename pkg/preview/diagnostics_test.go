package preview

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"aoe2kit/pkg/datfile"
)

func TestAssetDiagnosticsIdentifyAllConsumers(t *testing.T) {
	a, idx := studyFixture()
	for i := range idx.Civs[0].Units {
		idx.Civs[0].Units[i].Name = "Unit " + string(rune('A'+i))
	}
	crop := Crop{0, 0, 8, 8}
	r, err := ArtworkAssets(a, idx, t.TempDir(), &crop)
	if err != nil || len(r.Assets) != 1 || r.Missing != 1 {
		t.Fatalf("manifest=%+v, err=%v", r, err)
	}
	asset := r.Assets[0]
	if len(asset.RequiredBy) != 6 || len(asset.RequiredBy) != len(asset.Units) {
		t.Fatalf("missing shared consumers: %+v", asset)
	}
	for i, u := range asset.RequiredBy {
		if u.UnitID != asset.Units[i] || u.UnitName != idx.Civs[0].Units[u.UnitID-1].Name {
			t.Fatalf("incorrect association: %+v", asset)
		}
	}
	b, err := json.Marshal(asset)
	if err != nil || !strings.Contains(string(b), `"unit_name":"Unit A"`) || !strings.Contains(string(b), `"unit_ids":`) {
		t.Fatalf("JSON lost names or existing IDs: %s, %v", b, err)
	}
	// Fetch reports retain the same identities for assets they cannot copy.
	source := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "resources/_common/drs/graphics"), 0755); err != nil {
		t.Fatal(err)
	}
	fetched, err := FetchMissingSLDs(r, source, t.TempDir())
	if err == nil || len(fetched.Remaining) != 1 || !reflect.DeepEqual(fetched.Remaining[0].RequiredBy, asset.RequiredBy) {
		t.Fatalf("fetch report=%+v, err=%v", fetched, err)
	}
	idx.Graphics[0].ParticleEffectName.Value = "missing-effect"
	r, err = ArtworkAssets(a, idx, t.TempDir(), &crop)
	if err != nil || len(r.Unresolved) == 0 || r.Unresolved[0].UnitName == "" {
		t.Fatalf("unnamed particle dependency: %+v, %v", r, err)
	}
}

func TestSpriteFailureDiagnosticsIncludeDATName(t *testing.T) {
	for _, tc := range []struct {
		name, reason string
		change       func(*testing.T, string, *datfile.Index, *Snapshot, *Options)
	}{
		{"missing", "missing sprite:", func(t *testing.T, root string, idx *datfile.Index, s *Snapshot, opts *Options) {
			idx.Graphics[0].FileName.Value = "u_shp_canoe_x1"
		}},
		{"decode", "SLD decode:", func(t *testing.T, root string, idx *datfile.Index, s *Snapshot, opts *Options) {
			if err := os.WriteFile(filepath.Join(root, "resources/_common/drs/graphics/synthetic.sld"), []byte("broken"), 0600); err != nil {
				t.Fatal(err)
			}
		}},
		{"graphic", "standing graphic absent", func(t *testing.T, root string, idx *datfile.Index, s *Snapshot, opts *Options) {
			idx.Civs[0].Units[0].StandingGraphic1 = 999
		}},
		{"particle", "particle:", func(t *testing.T, root string, idx *datfile.Index, s *Snapshot, opts *Options) {
			idx.Graphics[0].FileName.Value = ""
			idx.Graphics[0].ParticleEffectName.Value = "missing-effect"
		}},
		{"crop", "outside crop", func(t *testing.T, root string, idx *datfile.Index, s *Snapshot, opts *Options) {
			opts.Crop = &Crop{0, 0, 2, 2}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, idx := spriteFixture(t)
			idx.Civs[0].Units[0].Name = "Canoe"
			s := spriteSnapshot()
			s.Objects[0].Name = "stale snapshot name"
			opts := Options{Scale: 4}
			tc.change(t, root, idx, &s, &opts)
			_, r, err := RenderSprites(s, opts, idx, root)
			if err != nil || len(r.Skipped) != 1 {
				t.Fatalf("report=%+v, err=%v", r, err)
			}
			skip := r.Skipped[0]
			if skip.UnitID != 7 || skip.UnitName != "Canoe" || !strings.HasPrefix(skip.Reason, tc.reason) {
				t.Fatalf("diagnostic=%+v", skip)
			}
			if err := CheckSpriteCompleteness(r); (err == nil) != (tc.name == "crop") {
				t.Fatalf("strict semantics changed: %v", err)
			}
		})
	}
}

func TestSkippedNameFallback(t *testing.T) {
	for _, idx := range []*datfile.Index{nil, {}} {
		s := skippedObject(Object{UnitID: 123, Name: "Snapshot name"}, idx, "unknown unit")
		if s.UnitID != 123 || s.UnitName != "Snapshot name" {
			t.Fatalf("fallback=%+v", s)
		}
		if s := skippedObject(Object{UnitID: 123}, idx, "unknown unit"); s.UnitName != "" {
			t.Fatalf("invented name: %+v", s)
		}
	}
}
