package preview

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"aoe2kit/pkg/datfile"
)

const renderIndexSchema = "aoe2kit.preview.index.v1"
const maxRenderIndexBytes = 32 << 20

type renderIndexCache struct {
	Schema       string
	SourceSHA256 string
	Index        *datfile.Index
}

// LoadRenderIndex reuses a content-keyed, render-only DAT projection when a
// cache directory is supplied. The first request builds it with the normal DAT
// reader; servers should prewarm it before accepting render requests. The result
// is not a complete DAT and must never be used for editing or gameplay queries.
func LoadRenderIndex(datPath, cacheDir string) (*datfile.Index, error) {
	if cacheDir == "" {
		return datfile.OpenReadOnly(datPath)
	}
	key, err := renderDATKey(datPath)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("preview cache: %w", err)
	}
	path := filepath.Join(cacheDir, "render-index-v1-"+key+".gob")
	if idx, err := readRenderIndex(path, key); err == nil {
		idx.Path = datPath
		return idx, nil
	}
	full, err := datfile.OpenReadOnly(datPath)
	if err != nil {
		return nil, err
	}
	after, err := renderDATKey(datPath)
	if err != nil {
		return nil, err
	}
	if after != key {
		return nil, fmt.Errorf("DAT changed while building preview cache; retry")
	}
	idx := compactRenderIndex(full)
	entry := renderIndexCache{Schema: renderIndexSchema, SourceSHA256: key, Index: idx}
	tmp, err := os.CreateTemp(cacheDir, ".render-index-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	err = gob.NewEncoder(tmp).Encode(entry)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return nil, err
	}
	idx.Path = datPath
	return idx, nil
}

func renderDATKey(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func readRenderIndex(path, key string) (*datfile.Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() > maxRenderIndexBytes {
		return nil, fmt.Errorf("preview index exceeds size limit")
	}
	var entry renderIndexCache
	if err := gob.NewDecoder(io.LimitReader(f, maxRenderIndexBytes)).Decode(&entry); err != nil {
		return nil, err
	}
	if entry.Schema != renderIndexSchema || entry.SourceSHA256 != key || entry.Index == nil || len(entry.Index.Civs) != 1 {
		return nil, fmt.Errorf("invalid preview index cache")
	}
	return entry.Index, nil
}

func compactRenderIndex(src *datfile.Index) *datfile.Index {
	idx := &datfile.Index{Version: src.Version, Compressed: src.Compressed, Inflated: src.Inflated, PlayerColours: append([]datfile.PlayerColour(nil), src.PlayerColours...)}
	for _, t := range src.Terrains {
		idx.Terrains = append(idx.Terrains, datfile.Terrain{Index: t.Index, Name: datfile.DebugString{Value: t.Name.Value}, Name2: datfile.DebugString{Value: t.Name2.Value}, OverlayMaskName: datfile.DebugString{Value: t.OverlayMaskName.Value}, BlendPriority: t.BlendPriority, BlendType: t.BlendType, MinimapColorHigh: t.MinimapColorHigh, MinimapColorMedium: t.MinimapColorMedium, MinimapColorLow: t.MinimapColorLow, MinimapColorCliffLeft: t.MinimapColorCliffLeft, MinimapColorCliffRight: t.MinimapColorCliffRight})
	}
	idx.Civs = make([]datfile.Civ, 1)
	if len(src.Civs) == 0 {
		return idx
	}
	idx.Graphics = make([]datfile.Graphic, len(src.Graphics))
	for _, u := range src.Civs[0].Units {
		v := datfile.UnitSummary{CivIndex: 0, Index: u.Index, ID: u.ID, Present: u.Present, Type: u.Type, Name: u.Name, Class: u.Class, ClassName: u.ClassName, StandingGraphic1: u.StandingGraphic1, ClearanceSizeX: u.ClearanceSizeX, ClearanceSizeY: u.ClearanceSizeY, ObstructionTypeID: u.ObstructionTypeID, ObstructionClassID: u.ObstructionClassID, MapDrawLevel: u.MapDrawLevel, MapColor: u.MapColor}
		for _, a := range u.Attributes {
			if a.Active && a.AttributeType == 144 {
				v.Attributes = append(v.Attributes, a)
				break
			}
		}
		idx.Civs[0].Units = append(idx.Civs[0].Units, v)
		if g, ok := src.Graphic(int(u.StandingGraphic1)); ok {
			idx.Graphics[g.Index] = datfile.Graphic{Index: g.Index, Present: g.Present, Name: datfile.DebugString{Value: g.Name.Value}, FileName: datfile.DebugString{Value: g.FileName.Value}, ParticleEffectName: datfile.DebugString{Value: g.ParticleEffectName.Value}, AngleCount: g.AngleCount, FrameCount: g.FrameCount}
		}
	}
	idx.GraphicsSize = len(idx.Graphics)
	return idx
}
