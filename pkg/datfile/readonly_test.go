package datfile

import (
	"crypto/sha256"
	"encoding/gob"
	"os"
	"runtime"
	"testing"
)

func TestReadOnlyIndexMatchesVerifiedReader(t *testing.T) {
	path := os.Getenv("AOE2KIT_TEST_DAT")
	if path == "" {
		t.Skip("set AOE2KIT_TEST_DAT for full DAT reader equivalence")
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("DAT unavailable: %v", err)
	}
	// Stream each index separately instead of retaining two full DATs at once.
	fingerprint := func(open func(string) (*Index, error)) [32]byte {
		idx, err := open(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, civ := range idx.Civs {
			for _, u := range civ.Units {
				if len(u.FullFields) > 0 {
					t.Fatal("raw parser unexpectedly populated unordered full_fields")
				}
			}
		}
		h := sha256.New()
		if err := gob.NewEncoder(h).Encode(idx); err != nil {
			t.Fatal(err)
		}
		var sum [32]byte
		copy(sum[:], h.Sum(nil))
		return sum
	}
	a := fingerprint(Open)
	runtime.GC()
	b := fingerprint(OpenReadOnly)
	if a != b {
		t.Fatal("read-only reader changed decoded index")
	}
}

func TestReadOnlyStillRejectsInvalidPayload(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("not a DAT"), []byte("VER 8.9\x00")} {
		if _, err := parseIndex(data, false); err == nil {
			t.Fatal("invalid payload accepted")
		}
	}
}
