package gfx

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSLDMainLayer(t *testing.T) {
	file, err := ParseSLD(syntheticSLD())
	if err != nil {
		t.Fatalf("ParseSLD: %v", err)
	}
	if file.Version != 4 || len(file.Frames) != 1 {
		t.Fatalf("header = version %d frames %d", file.Version, len(file.Frames))
	}
	frame := file.Frames[0]
	if frame.Width != 4 || frame.Height != 4 || frame.HotspotX != 2 || frame.HotspotY != 3 {
		t.Fatalf("frame = %+v", frame)
	}
	layer, ok := frame.layer("main")
	if !ok {
		t.Fatal("missing main layer")
	}
	if layer.Width != 4 || layer.Height != 4 || layer.CommandCount != 1 || layer.DrawBlocks != 1 {
		t.Fatalf("layer = %+v", layer)
	}
	img, err := file.decodeMainLayer(frame, layer, nil)
	if err != nil {
		t.Fatalf("decodeMainLayer: %v", err)
	}
	got := img.RGBAAt(0, 0)
	if got.R < 248 || got.G != 0 || got.B != 0 || got.A != 255 {
		t.Fatalf("pixel = %#v, want opaque red", got)
	}
}

func TestRenderFramePreservesHotspotCanvas(t *testing.T) {
	file, err := ParseSLD(syntheticSLD())
	if err != nil {
		t.Fatal(err)
	}
	img, frame, err := file.RenderFrame(0, color.RGBA{})
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds() != image.Rect(0, 0, 4, 4) || frame.HotspotX != 2 || frame.HotspotY != 3 {
		t.Fatalf("render bounds=%v frame=%+v", img.Bounds(), frame)
	}
	if got := img.RGBAAt(0, 0); got.R < 248 || got.A != 255 {
		t.Fatalf("render pixel=%+v", got)
	}
}

func TestRenderFramePlayerMaskUsesBC4(t *testing.T) {
	b := syntheticSLD()
	b[24] |= 0x10
	// BC4 endpoints 0,255, indices: first pixel 0, second pixel 1.
	body := []byte{0, 1, 1, 0, 0, 1, 0, 255, 8, 0, 0, 0, 0, 0}
	b = binary.LittleEndian.AppendUint32(b, uint32(len(body)+4))
	b = append(b, body...)
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	s, err := ParseSLD(b)
	if err != nil {
		t.Fatal(err)
	}
	pc := color.RGBA{G: 255, A: 255}
	im, _, err := s.RenderFrame(0, pc)
	if err != nil {
		t.Fatal(err)
	}
	if im.RGBAAt(0, 0) != (color.RGBA{R: 255, A: 255}) || im.RGBAAt(1, 0) != pc {
		t.Fatalf("mask colors = %v, %v", im.RGBAAt(0, 0), im.RGBAAt(1, 0))
	}
}

func TestBC4EndpointModes(t *testing.T) {
	for _, tc := range []struct {
		a, b uint8
		want [8]uint8
	}{
		{255, 0, [8]uint8{255, 0, 218, 182, 145, 109, 72, 36}},
		{0, 255, [8]uint8{0, 255, 51, 102, 153, 204, 0, 255}},
	} {
		var bits uint64
		for i := 0; i < 16; i++ {
			bits |= uint64(i%8) << (3 * i)
		}
		data := []byte{tc.a, tc.b}
		for i := 0; i < 6; i++ {
			data = append(data, uint8(bits>>(8*i)))
		}
		out := decodeBC4Block(data)
		for i, v := range out {
			if v != tc.want[i%8] {
				t.Fatalf("%d,%d pixel%d=%d want%d", tc.a, tc.b, i, v, tc.want[i%8])
			}
		}
	}
}

func TestRenderShadowFrame(t *testing.T) {
	b := syntheticSLD()
	b[24] |= 2
	// Shadow graphics have a 12-byte header, unlike the player-color mask.
	body := []byte{1, 0, 1, 0, 4, 0, 4, 0, 0, 1, 1, 0, 0, 1, 80, 0, 0, 0, 0, 0, 0, 0}
	b = binary.LittleEndian.AppendUint32(b, uint32(len(body)+4))
	b = append(b, body...)
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	s, err := ParseSLD(b)
	if err != nil {
		t.Fatal(err)
	}
	img, frame, err := s.RenderShadowFrame(0)
	if err != nil {
		t.Fatal(err)
	}
	if frame.HotspotX != 2 || frame.HotspotY != 3 || img.Bounds() != image.Rect(0, 0, 4, 4) {
		t.Fatalf("canvas/hotspot: %+v", frame)
	}
	if img.RGBAAt(0, 0).A != 0 || img.RGBAAt(1, 1) != (color.RGBA{A: 80}) || img.RGBAAt(3, 3).A != 80 {
		t.Fatal("shadow offsets/coverage lost")
	}
	if _, _, err := s.RenderShadowFrame(1); err == nil {
		t.Fatal("accepted invalid ordinal")
	}
	l, _ := frame.layer("shadow")
	previous := image.NewRGBA(img.Bounds())
	previous.SetRGBA(0, 0, color.RGBA{A: 40})
	l.Flag1 = 0x80
	delta, err := s.decodeGraphicsLayer(frame, l, previous, true)
	if err != nil || delta.RGBAAt(0, 0).A != 40 {
		t.Fatalf("delta preservation: %v", err)
	}
	l.End--
	if _, err := s.decodeGraphicsLayer(frame, l, nil, true); err == nil {
		t.Fatal("accepted truncated shadow")
	}
	s, _ = ParseSLD(syntheticSLD())
	img, _, err = s.RenderShadowFrame(0)
	if err != nil || img != nil {
		t.Fatal("invented missing shadow")
	}
}

func TestParseSLDShortFileHeader(t *testing.T) {
	file, err := ParseSLD(syntheticShortFileHeaderSLD())
	if err != nil {
		t.Fatalf("ParseSLD: %v", err)
	}
	if len(file.Frames) != 1 {
		t.Fatalf("frames = %d, want 1", len(file.Frames))
	}
	frame := file.Frames[0]
	if frame.Width != 4 || frame.Height != 4 || frame.HotspotX != 2 || frame.HotspotY != 3 || frame.Type != 0x01 || frame.Start != 14 {
		t.Fatalf("short-file-header frame = %+v", frame)
	}
	if len(frame.Layers) != 1 || frame.Layers[0].Name != "main" || frame.Layers[0].ContentLength != 26 {
		t.Fatalf("layers = %+v", frame.Layers)
	}
}

func TestExportSLDMainLayerPNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "one.sld")
	if err := os.WriteFile(path, syntheticSLD(), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := ExportSLD(path, SLDExportOptions{OutDir: filepath.Join(dir, "out")})
	if err != nil {
		t.Fatalf("ExportSLD: %v", err)
	}
	if !report.OK || len(report.Exported) != 1 {
		t.Fatalf("report = %+v", report)
	}
	if filepath.Base(report.Exported[0].Path) != "frame_00.png" {
		t.Fatalf("exported path = %s, want frame_00.png", report.Exported[0].Path)
	}
	if report.Exported[0].OpaquePixels == 0 {
		t.Fatalf("opaque pixels = 0")
	}
	data, err := os.ReadFile(report.Exported[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode exported png: %v", err)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Fatalf("png bounds = %v", img.Bounds())
	}
	manifestData, err := os.ReadFile(filepath.Join(dir, "out", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest SLDManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if manifest.FrameCount != 1 || len(manifest.Frames) != 1 {
		t.Fatalf("manifest frames = %d/%d, want 1/1", manifest.FrameCount, len(manifest.Frames))
	}
	if manifest.Frames[0].PNG != "frame_00.png" || manifest.Frames[0].AnchorX != 2 || manifest.Frames[0].AnchorY != 3 {
		t.Fatalf("manifest frame = %+v", manifest.Frames[0])
	}
	if _, err := os.Stat(filepath.Join(dir, "out", "contact_sheet.png")); err != nil {
		t.Fatal(err)
	}
}

func TestParseSLDAlignsOddLengthLayersForward(t *testing.T) {
	file, err := ParseSLD(syntheticSLDWithOddUnknownLayer())
	if err != nil {
		t.Fatalf("ParseSLD: %v", err)
	}
	frame := file.Frames[0]
	if len(frame.Layers) != 3 {
		t.Fatalf("layers = %d, want 3", len(frame.Layers))
	}
	if frame.Layers[1].Name != "unknown" || frame.Layers[1].ContentLength != 6 || frame.Layers[1].PaddedLength != 8 {
		t.Fatalf("unknown layer = %+v", frame.Layers[1])
	}
	if frame.Layers[2].Name != "playercolor" || frame.Layers[2].ContentLength != 8 {
		t.Fatalf("playercolor layer = %+v", frame.Layers[2])
	}
}

func TestExportRealSLDSamples(t *testing.T) {
	samples := []string{
		"s_medi_courtyard_wall_x1.sld",
		"s_hedge_garden_x1.sld",
	}
	for _, sample := range samples {
		t.Run(sample, func(t *testing.T) {
			path := filepath.Join("..", "..", "testdata", "sld-samples", sample)
			if _, err := os.Stat(path); err != nil {
				t.Skipf("SLD sample unavailable: %v", err)
			}
			outDir := filepath.Join(t.TempDir(), "out")
			report, err := ExportSLD(path, SLDExportOptions{OutDir: outDir})
			if err != nil {
				t.Fatalf("ExportSLD: %v", err)
			}
			if !report.OK || report.Frames != 10 || len(report.Exported) != 10 {
				t.Fatalf("report = %+v", report)
			}
			if report.ManifestPath == "" || report.ContactSheet == "" {
				t.Fatalf("manifest/contact paths missing: %+v", report)
			}
			for i := 0; i < 2; i++ {
				frame := report.Exported[i]
				if frame.Width != 300 || frame.Height != 300 || frame.HotspotX != 150 || frame.HotspotY != 150 {
					t.Fatalf("frame %d geometry = %+v", i, frame)
				}
				if frame.OpaquePixels == 0 {
					t.Fatalf("frame %d has no opaque pixels", i)
				}
				data, err := os.ReadFile(frame.Path)
				if err != nil {
					t.Fatal(err)
				}
				img, err := png.Decode(bytes.NewReader(data))
				if err != nil {
					t.Fatalf("decode frame %d: %v", i, err)
				}
				if img.Bounds().Dx() != frame.Width || img.Bounds().Dy() != frame.Height {
					t.Fatalf("frame %d png bounds = %v, want %dx%d", i, img.Bounds(), frame.Width, frame.Height)
				}
			}
			if _, err := os.Stat(filepath.Join(outDir, "frame_09.png")); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func syntheticSLD() []byte {
	var out []byte
	out = append(out, []byte("SLDX")...)
	out = binary.LittleEndian.AppendUint16(out, 4)
	out = binary.LittleEndian.AppendUint16(out, 1)
	out = binary.LittleEndian.AppendUint16(out, 0)
	out = binary.LittleEndian.AppendUint16(out, 16)
	out = binary.LittleEndian.AppendUint32(out, 0xff000000)
	out = binary.LittleEndian.AppendUint16(out, 4)
	out = binary.LittleEndian.AppendUint16(out, 4)
	out = binary.LittleEndian.AppendUint16(out, uint16(int16(2)))
	out = binary.LittleEndian.AppendUint16(out, uint16(int16(3)))
	out = append(out, 0x01, 0x01)
	out = binary.LittleEndian.AppendUint16(out, 0)

	var body []byte
	body = binary.LittleEndian.AppendUint16(body, 0)
	body = binary.LittleEndian.AppendUint16(body, 0)
	body = binary.LittleEndian.AppendUint16(body, 4)
	body = binary.LittleEndian.AppendUint16(body, 4)
	body = append(body, 0x00, 0x01)
	body = binary.LittleEndian.AppendUint16(body, 1)
	body = append(body, 0x00, 0x01)
	body = binary.LittleEndian.AppendUint16(body, 0xf800)
	body = binary.LittleEndian.AppendUint16(body, 0x0000)
	body = binary.LittleEndian.AppendUint32(body, 0)

	contentLength := uint32(len(body) + 4)
	out = binary.LittleEndian.AppendUint32(out, contentLength)
	out = append(out, body...)
	for len(out)%4 != 0 {
		out = append(out, 0)
	}
	return out
}

func syntheticShortFileHeaderSLD() []byte {
	out := syntheticSLD()
	out[10] = 0x0e
	out[11] = 0x00
	return append(append([]byte{}, out[:14]...), out[16:]...)
}

func syntheticSLDWithOddUnknownLayer() []byte {
	out := syntheticSLD()
	frameTypeOffset := 24
	out[frameTypeOffset] = 0x15

	var unknown []byte
	unknown = binary.LittleEndian.AppendUint32(unknown, 6)
	unknown = append(unknown, 0xaa, 0xbb)
	unknown = append(unknown, 0x00, 0x00)

	var playerColor []byte
	playerColor = binary.LittleEndian.AppendUint32(playerColor, 8)
	playerColor = append(playerColor, 0x00, 0x00)
	playerColor = binary.LittleEndian.AppendUint16(playerColor, 0)

	out = append(out, unknown...)
	out = append(out, playerColor...)
	return out
}

func TestShortFileHeaderKeepsRectangularCanvasAndDeltaPixels(t *testing.T) {
	b := syntheticShortFileHeaderSLD()
	binary.LittleEndian.PutUint16(b[6:8], 2)
	// A rectangular canvas and asymmetric hotspot expose a two-byte shift.
	for i, v := range []uint16{8, 12, 2, 3} {
		binary.LittleEndian.PutUint16(b[14+2*i:], v)
	}
	for i, v := range []uint16{2, 8, 6, 12} {
		binary.LittleEndian.PutUint16(b[30+2*i:], v)
	}
	b = append(b, b[14:26]...)
	// Frame one retains its predecessor's bottom block without drawing.
	body := []byte{}
	for _, v := range []uint16{2, 8, 6, 12} {
		body = binary.LittleEndian.AppendUint16(body, v)
	}
	body = append(body, 0x80, 1, 1, 0, 1, 0)
	b = binary.LittleEndian.AppendUint32(b, uint32(len(body)+4))
	b = append(b, body...)
	b = append(b, 0, 0)
	f, err := ParseSLD(b)
	if err != nil {
		t.Fatal(err)
	}
	for ordinal := 0; ordinal < 2; ordinal++ {
		img, frame, err := f.RenderFrame(ordinal, color.RGBA{})
		if err != nil {
			t.Fatal(err)
		}
		if frame.Width != 8 || frame.Height != 12 || frame.HotspotX != 2 || frame.HotspotY != 3 || img.Bounds() != image.Rect(0, 0, 8, 12) {
			t.Fatalf("frame%d geometry: %+v %v", ordinal, frame, img.Bounds())
		}
		if got := img.RGBAAt(5, 11); got != (color.RGBA{255, 0, 0, 255}) {
			t.Fatalf("frame%d lower block clipped: %v", ordinal, got)
		}
	}
}

func TestShadowVariantsAreIndependentAndDeltaBitRetains(t *testing.T) {
	b := []byte("SLDX")
	for _, v := range []uint16{4, 3, 0, 16} {
		b = binary.LittleEndian.AppendUint16(b, v)
	}
	b = binary.LittleEndian.AppendUint32(b, 0)
	for i, tc := range []struct {
		x, y, w, h  uint16
		flag, alpha byte
	}{
		{0, 0, 4, 4, 1, 80}, {0, 0, 1, 1, 1, 160}, {1, 1, 1, 1, 0x81, 240},
	} {
		for _, v := range []uint16{4, 4, 2, 3} {
			b = binary.LittleEndian.AppendUint16(b, v)
		}
		b = append(b, 2, 0)
		b = binary.LittleEndian.AppendUint16(b, uint16(i))
		body := []byte{}
		for _, v := range []uint16{tc.x, tc.y, tc.x + tc.w, tc.y + tc.h} {
			body = binary.LittleEndian.AppendUint16(body, v)
		}
		body = append(body, tc.flag, 1, 1, 0, 0, 1, tc.alpha, 0, 0, 0, 0, 0, 0, 0)
		b = binary.LittleEndian.AppendUint32(b, uint32(len(body)+4))
		b = append(b, body...)
		b = append(b, 0, 0)
	}
	f, err := ParseSLD(b)
	if err != nil {
		t.Fatal(err)
	}
	static, _, err := f.RenderShadowFrame(1)
	if err != nil {
		t.Fatal(err)
	}
	if static.RGBAAt(0, 0).A != 160 || static.RGBAAt(3, 3).A != 0 {
		t.Fatal("independent facing retained previous silhouette")
	}
	delta, _, err := f.RenderShadowFrame(2)
	if err != nil {
		t.Fatal(err)
	}
	if delta.RGBAAt(0, 0).A != 160 || delta.RGBAAt(1, 1).A != 240 || delta.RGBAAt(3, 3).A != 0 {
		t.Fatal("delta did not retain only its immediate predecessor")
	}
}
