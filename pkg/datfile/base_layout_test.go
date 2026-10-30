package datfile

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestCurrentDEBaseFlagsAndOutline(t *testing.T) {
	// VER 8.8+ base layout through resource storage, damage-graphics count,
	// sounds, an empty debug name, and the two copy ids. These offsets follow
	// the serialized layout; deliberately distinct flags catch a four-byte drift
	// caused by reading the obsolete base hotkey slot.
	record := make([]byte, 180)
	record[0] = 10
	record[103] = 7 // recyclable
	record[104] = 1 // enable auto gather
	record[108] = 2 // obstruction type
	record[109] = 3 // obstruction class
	record[110] = 4 // trait, not obstruction type
	record[111] = 5 // civilization, not obstruction class
	for i, value := range []float32{1.25, 2.5, 3.75} {
		binary.LittleEndian.PutUint32(record[116+4*i:], math.Float32bits(value))
	}
	for _, value := range []uint16{DebugStringMarker, 0, 0, 0} {
		record = binary.LittleEndian.AppendUint16(record, value)
	}
	unit, err := ParseUnitRecord(record, 0, 0, "VER 8.9", 300)
	if err != nil {
		t.Fatal(err)
	}
	if unit.CanBeGathered != 1 || unit.ObstructionTypeID != 2 || unit.ObstructionClassID != 3 {
		t.Fatalf("base flags = %d/%d/%d", unit.CanBeGathered, unit.ObstructionTypeID, unit.ObstructionClassID)
	}
	if unit.OutlineRadiusX != 1.25 || unit.OutlineRadiusY != 2.5 || unit.OutlineRadiusZ != 3.75 {
		t.Fatalf("outline = %v/%v/%v", unit.OutlineRadiusX, unit.OutlineRadiusY, unit.OutlineRadiusZ)
	}
	if unit.FieldSpans.ObstructionTypeID.Start != 408 {
		t.Fatalf("obstruction span = %+v", unit.FieldSpans.ObstructionTypeID)
	}
}
