package datfile

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestCreatableTailReferencesAndProjectiles(t *testing.T) {
	record := make([]byte, 117)
	// Three costs and a zero train-location count precede the creatable tail.
	binary.LittleEndian.PutUint32(record[30:], 123456) // garrison graphic
	binary.LittleEndian.PutUint16(record[34:], 321)    // spawning graphic
	binary.LittleEndian.PutUint16(record[36:], 654)    // upgrade graphic
	binary.LittleEndian.PutUint16(record[54:], 9)      // charge target
	binary.LittleEndian.PutUint32(record[89:], math.Float32bits(2.5))
	record[93] = 7
	binary.LittleEndian.PutUint32(record[106:], 0xffffffff) // absent secondary projectile
	c := cursor{data: record}
	got, err := parseCreatable(&c, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.GarrisonGraphic != 123456 || got.SpawningGraphic != 321 || got.UpgradeGraphic != 654 || got.ChargeTarget != 9 {
		t.Fatalf("creatable references = %+v", got)
	}
	if got.TotalProjectiles != 2.5 || got.MaxTotalProjectiles != 7 || got.SecondaryProjectileUnit != -1 || c.off != len(record) {
		t.Fatalf("projectile tail = %+v, consumed=%d", got, c.off)
	}
}

func TestCreatableRejectsTruncatedTail(t *testing.T) {
	// Three six-byte costs, a zero location count, the 45-byte prefix,
	// two two-byte button fields separated by eight bytes, and the 40-byte tail.
	record := make([]byte, 18+2+45+4+8+40)
	for size := 0; size < len(record); size++ {
		c := cursor{data: record[:size]}
		if _, err := parseCreatable(&c, true); err == nil {
			t.Fatalf("accepted creatable truncated to %d bytes", size)
		}
	}
	c := cursor{data: record}
	if _, err := parseCreatable(&c, true); err != nil {
		t.Fatalf("complete creatable: %v", err)
	}
}
