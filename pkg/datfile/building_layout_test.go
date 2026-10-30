package datfile

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestBuildingGraphicsReferencesAndSounds(t *testing.T) {
	// Fixed building prefix (25), four annexes (40), and sound/garrison tail (33).
	record := make([]byte, 98)
	binary.LittleEndian.PutUint16(record, 101)
	binary.LittleEndian.PutUint16(record[2:], 102)
	binary.LittleEndian.PutUint16(record[16:], 0xffff) // stack unit
	binary.LittleEndian.PutUint16(record[18:], 42)     // foundation terrain
	record[24] = 1
	binary.LittleEndian.PutUint16(record[65:], 0xffff) // head unit
	binary.LittleEndian.PutUint16(record[67:], 83)     // transform unit
	binary.LittleEndian.PutUint16(record[69:], 111)    // transform sound
	binary.LittleEndian.PutUint16(record[71:], 222)    // construction sound
	binary.LittleEndian.PutUint32(record[73:], 0x12345678)
	binary.LittleEndian.PutUint32(record[77:], 0x87654321)
	record[81] = 5
	binary.LittleEndian.PutUint32(record[82:], math.Float32bits(1.5))
	binary.LittleEndian.PutUint32(record[86:], math.Float32bits(2.5))
	binary.LittleEndian.PutUint16(record[90:], 0xffff) // pile unit
	copy(record[92:], []byte{1, 2, 3, 4, 5, 6})
	c := cursor{data: record}
	got, err := parseBuilding(&c)
	if err != nil {
		t.Fatal(err)
	}
	if got.ConstructionGraphicID != 101 || got.SnowGraphicID != 102 || got.FoundationTerrainID != 42 || got.CanBurn != 1 {
		t.Fatalf("building prefix = %+v", got)
	}
	if got.StackUnit != -1 || got.HeadUnit != -1 || got.TransformUnit != 83 || got.PileUnit != -1 {
		t.Fatalf("building unit references = %+v", got)
	}
	if got.TransformSound != 111 || got.ConstructionSound != 222 || got.WwiseTransformSoundID != 0x12345678 || got.WwiseConstructionSoundID != 0x87654321 {
		t.Fatalf("building sounds = %+v", got)
	}
	if got.GarrisonType != 5 || got.GarrisonHealRate != 1.5 || got.GarrisonRepairRate != 2.5 || len(got.LootingTable) != 6 || got.LootingTable[5] != 6 || c.off != len(record) {
		t.Fatalf("building tail = %+v, consumed=%d", got, c.off)
	}
}
