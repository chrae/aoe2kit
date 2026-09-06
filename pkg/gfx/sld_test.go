package gfx

import (
	"bytes"
	"encoding/binary"
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
