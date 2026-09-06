package gfx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"aoe2kit/pkg/aoe2"
)

const Version = "aoe2kit.gfx.v1"

type SLD struct {
	Path     string     `json:"path,omitempty"`
	Size     int        `json:"size"`
	Version  uint16     `json:"version"`
	Frames   []SLDFrame `json:"frames"`
	raw      []byte
	warnings []string
}

type SLDFrame struct {
	Ordinal  int        `json:"ordinal"`
	Index    uint16     `json:"index"`
	Width    int        `json:"width"`
	Height   int        `json:"height"`
	HotspotX int        `json:"hotspot_x"`
	HotspotY int        `json:"hotspot_y"`
	Type     uint8      `json:"frame_type"`
	Unknown  uint8      `json:"unknown"`
	Start    int        `json:"start"`
	End      int        `json:"end"`
	Layers   []SLDLayer `json:"layers,omitempty"`
}

type SLDLayer struct {
	Name          string `json:"name"`
	Compression   string `json:"compression,omitempty"`
	Start         int    `json:"start"`
	End           int    `json:"end"`
	ContentLength int    `json:"content_length"`
	PaddedLength  int    `json:"padded_length"`
	OffsetX1      int    `json:"offset_x1,omitempty"`
	OffsetY1      int    `json:"offset_y1,omitempty"`
	OffsetX2      int    `json:"offset_x2,omitempty"`
	OffsetY2      int    `json:"offset_y2,omitempty"`
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	Flag1         uint8  `json:"flag1,omitempty"`
	Unknown       uint8  `json:"unknown,omitempty"`
	CommandCount  int    `json:"command_count,omitempty"`
	DrawBlocks    int    `json:"draw_blocks,omitempty"`
}

type SLDExportOptions struct {
	OutDir string
	Limit  int
}

type SLDExportReport struct {
	Version      string                 `json:"version"`
	OK           bool                   `json:"ok"`
	Verification aoe2.VerificationClaim `json:"verification"`
	Source       string                 `json:"source"`
	OutputDir    string                 `json:"output_dir"`
	Frames       int                    `json:"frames"`
	Exported     []SLDExportedFrame     `json:"exported,omitempty"`
	Warnings     []string               `json:"warnings,omitempty"`
}

type SLDExportedFrame struct {
	FrameOrdinal int    `json:"frame_ordinal"`
	FrameIndex   uint16 `json:"frame_index"`
	Layer        string `json:"layer"`
	Path         string `json:"path"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	HotspotX     int    `json:"hotspot_x"`
	HotspotY     int    `json:"hotspot_y"`
}

func OpenSLD(path string) (*SLD, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	file, err := ParseSLD(data)
	if err != nil {
		return nil, err
	}
	file.Path = path
	return file, nil
}

func ParseSLD(data []byte) (*SLD, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("SLD too short: %d bytes", len(data))
	}
	if string(data[:4]) != "SLDX" {
		return nil, fmt.Errorf("bad SLD signature %q", string(data[:4]))
	}
	file := &SLD{
		Size:    len(data),
		Version: binary.LittleEndian.Uint16(data[4:6]),
		raw:     append([]byte(nil), data...),
	}
	frameCount := int(binary.LittleEndian.Uint16(data[6:8]))
	pos := 16
	for i := 0; i < frameCount; i++ {
		if pos+12 > len(data) {
			return nil, fmt.Errorf("frame %d header exceeds file at %d", i, pos)
		}
		frame := SLDFrame{
			Ordinal:  i,
			Width:    int(binary.LittleEndian.Uint16(data[pos : pos+2])),
			Height:   int(binary.LittleEndian.Uint16(data[pos+2 : pos+4])),
			HotspotX: int(int16(binary.LittleEndian.Uint16(data[pos+4 : pos+6]))),
			HotspotY: int(int16(binary.LittleEndian.Uint16(data[pos+6 : pos+8]))),
			Type:     data[pos+8],
			Unknown:  data[pos+9],
			Index:    binary.LittleEndian.Uint16(data[pos+10 : pos+12]),
			Start:    pos,
		}
		pos += 12
		for _, kind := range sldLayerKinds {
			if frame.Type&kind.Mask == 0 {
				continue
			}
			layer, next, err := parseSLDLayer(data, pos, frame, kind)
			if err != nil {
				return nil, fmt.Errorf("frame %d %s layer: %w", i, kind.Name, err)
			}
			frame.Layers = append(frame.Layers, layer)
			pos = next
		}
		frame.End = pos
		file.Frames = append(file.Frames, frame)
	}
	if pos != len(data) {
		file.warnings = append(file.warnings, fmt.Sprintf("%d trailing bytes after final frame", len(data)-pos))
	}
	return file, nil
}

func ExportSLD(path string, options SLDExportOptions) (SLDExportReport, error) {
	file, err := OpenSLD(path)
	if err != nil {
		return SLDExportReport{}, err
	}
	outDir := options.OutDir
	if outDir == "" {
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		outDir = base + "_png"
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return SLDExportReport{}, err
	}
	report := SLDExportReport{
		Version:      Version,
		OK:           true,
		Verification: aoe2.StructureVerification(true),
		Source:       path,
		OutputDir:    outDir,
		Frames:       len(file.Frames),
		Warnings:     append([]string(nil), file.warnings...),
	}
	limit := len(file.Frames)
	if options.Limit > 0 && options.Limit < limit {
		limit = options.Limit
	}
	var previousMain *image.RGBA
	for i := 0; i < limit; i++ {
		frame := file.Frames[i]
		layer, ok := frame.layer("main")
		if !ok {
			report.Warnings = append(report.Warnings, fmt.Sprintf("frame %d has no main layer", frame.Ordinal))
			previousMain = nil
			continue
		}
		img, err := file.decodeMainLayer(frame, layer, previousMain)
		if err != nil {
			return report, fmt.Errorf("frame %d main layer: %w", frame.Ordinal, err)
		}
		previousMain = img
		name := fmt.Sprintf("%s_frame_%04d_main.png", strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), frame.Ordinal)
		outPath := filepath.Join(outDir, name)
		if err := writePNG(outPath, img); err != nil {
			return report, err
		}
		report.Exported = append(report.Exported, SLDExportedFrame{
			FrameOrdinal: frame.Ordinal,
			FrameIndex:   frame.Index,
			Layer:        "main",
			Path:         outPath,
			Width:        frame.Width,
			Height:       frame.Height,
			HotspotX:     frame.HotspotX,
			HotspotY:     frame.HotspotY,
		})
	}
	if limit < len(file.Frames) {
		report.Warnings = append(report.Warnings, fmt.Sprintf("export limited to %d of %d frames", limit, len(file.Frames)))
	}
	return report, nil
}

func (f *SLD) decodeMainLayer(frame SLDFrame, layer SLDLayer, previous *image.RGBA) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, frame.Width, frame.Height))
	if previous != nil && len(previous.Pix) == len(img.Pix) && (layer.Flag1&0x80 != 0 || layer.Flag1&0x01 != 0) {
		copy(img.Pix, previous.Pix)
	}
	contentStart := layer.Start + 4
	headerStart := contentStart
	if headerStart+12 > layer.End {
		return nil, fmt.Errorf("layer data too short")
	}
	commandCount := int(binary.LittleEndian.Uint16(f.raw[headerStart+10 : headerStart+12]))
	commandsStart := headerStart + 12
	blocksStart := commandsStart + commandCount*2
	if blocksStart > layer.End {
		return nil, fmt.Errorf("command array exceeds layer")
	}
	blockWidth := divCeil(layer.Width, 4)
	blockHeight := divCeil(layer.Height, 4)
	totalBlocks := blockWidth * blockHeight
	blockIndex := 0
	dataPos := blocksStart
	for i := 0; i < commandCount; i++ {
		cmd := f.raw[commandsStart+i*2 : commandsStart+i*2+2]
		skip := int(cmd[0])
		draw := int(cmd[1])
		blockIndex += skip
		for j := 0; j < draw; j++ {
			if blockIndex >= totalBlocks {
				return nil, fmt.Errorf("draw command exceeds %d layer blocks", totalBlocks)
			}
			if dataPos+8 > layer.End {
				return nil, fmt.Errorf("compressed block data truncated")
			}
			block := decodeBC1Block(f.raw[dataPos : dataPos+8])
			dataPos += 8
			bx := blockIndex % blockWidth
			by := blockIndex / blockWidth
			paintBlock(img, layer, bx, by, block)
			blockIndex++
		}
	}
	if dataPos != layer.End {
		return nil, fmt.Errorf("unused compressed block bytes: %d", layer.End-dataPos)
	}
	return img, nil
}

func parseSLDLayer(data []byte, start int, frame SLDFrame, kind sldLayerKind) (SLDLayer, int, error) {
	if start+4 > len(data) {
		return SLDLayer{}, start, fmt.Errorf("missing content length at %d", start)
	}
	contentLength := int(binary.LittleEndian.Uint32(data[start : start+4]))
	if contentLength < 4 {
		return SLDLayer{}, start, fmt.Errorf("invalid content length %d", contentLength)
	}
	paddedLength := align4(contentLength)
	end := start + contentLength
	paddedEnd := start + paddedLength
	if end > len(data) || paddedEnd > len(data) {
		return SLDLayer{}, start, fmt.Errorf("content length %d exceeds file", contentLength)
	}
	layer := SLDLayer{
		Name:          kind.Name,
		Compression:   kind.Compression,
		Start:         start,
		End:           end,
		ContentLength: contentLength,
		PaddedLength:  paddedLength,
	}
	body := data[start+4 : end]
	switch kind.Name {
	case "main", "shadow":
		if len(body) < 12 {
			return SLDLayer{}, start, fmt.Errorf("graphics layer body too short")
		}
		layer.OffsetX1 = int(binary.LittleEndian.Uint16(body[0:2]))
		layer.OffsetY1 = int(binary.LittleEndian.Uint16(body[2:4]))
		layer.OffsetX2 = int(binary.LittleEndian.Uint16(body[4:6]))
		layer.OffsetY2 = int(binary.LittleEndian.Uint16(body[6:8]))
		layer.Flag1 = body[8]
		layer.Unknown = body[9]
		layer.CommandCount = int(binary.LittleEndian.Uint16(body[10:12]))
	case "damage", "playercolor":
		if len(body) < 4 {
			return SLDLayer{}, start, fmt.Errorf("mask layer body too short")
		}
		layer.OffsetX2 = frame.Width
		layer.OffsetY2 = frame.Height
		layer.Width = frame.Width
		layer.Height = frame.Height
		layer.Flag1 = body[0]
		layer.Unknown = body[1]
		layer.CommandCount = int(binary.LittleEndian.Uint16(body[2:4]))
	default:
		return layer, paddedEnd, nil
	}
	if layer.Width == 0 {
		layer.Width = layer.OffsetX2 - layer.OffsetX1
	}
	if layer.Height == 0 {
		layer.Height = layer.OffsetY2 - layer.OffsetY1
	}
	if layer.Width < 0 || layer.Height < 0 {
		return SLDLayer{}, start, fmt.Errorf("negative layer dimensions %dx%d", layer.Width, layer.Height)
	}
	commandBytes := layer.CommandCount * 2
	headerBytes := 12
	if kind.Name == "damage" || kind.Name == "playercolor" {
		headerBytes = 4
	}
	if len(body) < headerBytes+commandBytes {
		return SLDLayer{}, start, fmt.Errorf("command array exceeds layer body")
	}
	drawBlocks := 0
	commands := body[headerBytes : headerBytes+commandBytes]
	for i := 0; i < len(commands); i += 2 {
		drawBlocks += int(commands[i+1])
	}
	layer.DrawBlocks = drawBlocks
	blockBytes := drawBlocks * 8
	if len(body) != headerBytes+commandBytes+blockBytes {
		return SLDLayer{}, start, fmt.Errorf("body has %d bytes, expected %d", len(body), headerBytes+commandBytes+blockBytes)
	}
	return layer, paddedEnd, nil
}

type sldLayerKind struct {
	Name        string
	Mask        uint8
	Compression string
}

var sldLayerKinds = []sldLayerKind{
	{Name: "main", Mask: 0x01, Compression: "BC1"},
	{Name: "shadow", Mask: 0x02, Compression: "BC4"},
	{Name: "unknown", Mask: 0x04},
	{Name: "damage", Mask: 0x08, Compression: "BC1"},
	{Name: "playercolor", Mask: 0x10, Compression: "BC4"},
}

func (f SLDFrame) layer(name string) (SLDLayer, bool) {
	for _, layer := range f.Layers {
		if layer.Name == name {
			return layer, true
		}
	}
	return SLDLayer{}, false
}

func decodeBC1Block(data []byte) [16]color.RGBA {
	c0 := binary.LittleEndian.Uint16(data[0:2])
	c1 := binary.LittleEndian.Uint16(data[2:4])
	table := [4]color.RGBA{rgb565(c0), rgb565(c1)}
	if c0 > c1 {
		table[2] = mixColor(table[0], table[1], 2, 1, 3, 255)
		table[3] = mixColor(table[0], table[1], 1, 2, 3, 255)
	} else {
		table[2] = mixColor(table[0], table[1], 1, 1, 2, 255)
		table[3] = color.RGBA{}
	}
	indices := binary.LittleEndian.Uint32(data[4:8])
	var out [16]color.RGBA
	for i := 0; i < 16; i++ {
		out[i] = table[indices&0x03]
		indices >>= 2
	}
	return out
}

func rgb565(v uint16) color.RGBA {
	r := uint8((v >> 11) & 0x1f)
	g := uint8((v >> 5) & 0x3f)
	b := uint8(v & 0x1f)
	return color.RGBA{
		R: (r << 3) | (r >> 2),
		G: (g << 2) | (g >> 4),
		B: (b << 3) | (b >> 2),
		A: 255,
	}
}

func mixColor(a, b color.RGBA, aw, bw, div int, alpha uint8) color.RGBA {
	return color.RGBA{
		R: uint8((int(a.R)*aw + int(b.R)*bw) / div),
		G: uint8((int(a.G)*aw + int(b.G)*bw) / div),
		B: uint8((int(a.B)*aw + int(b.B)*bw) / div),
		A: alpha,
	}
}

func paintBlock(img *image.RGBA, layer SLDLayer, bx, by int, block [16]color.RGBA) {
	for py := 0; py < 4; py++ {
		for px := 0; px < 4; px++ {
			lx := bx*4 + px
			ly := by*4 + py
			if lx >= layer.Width || ly >= layer.Height {
				continue
			}
			img.SetRGBA(layer.OffsetX1+lx, layer.OffsetY1+ly, block[py*4+px])
		}
	}
}

func writePNG(path string, img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func divCeil(n, d int) int {
	if n <= 0 {
		return 0
	}
	return (n + d - 1) / d
}

func align4(n int) int {
	return n + ((4 - (n % 4)) % 4)
}
