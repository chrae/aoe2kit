package datcache

import (
	"os"
	"reflect"
	"testing"

	"aoe2kit/pkg/datfile"
)

func TestCachedUnitMatchesDAT(t *testing.T) {
	path := os.Getenv("AOE2KIT_TEST_DAT")
	if path == "" {
		t.Skip("set AOE2KIT_TEST_DAT to compare cached and direct unit decoding")
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("DAT unavailable: %v", err)
	}
	cache, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	got, found, err := cache.Unit(0, 83)
	if err != nil || !found {
		t.Fatalf("cached villager: found=%v, err=%v", found, err)
	}
	if len(got.FullFields) == 0 {
		t.Fatal("cached unit lacks full fields; CLI would reload DAT")
	}
	want, _, err := datfile.DecodeUnitFile(path, 0, 83)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		g, w := reflect.ValueOf(got), reflect.ValueOf(want)
		for i := 0; i < g.NumField(); i++ {
			if !reflect.DeepEqual(g.Field(i).Interface(), w.Field(i).Interface()) {
				t.Errorf("cached %s = %#v, direct = %#v", g.Type().Field(i).Name, g.Field(i).Interface(), w.Field(i).Interface())
			}
		}
	}
}
