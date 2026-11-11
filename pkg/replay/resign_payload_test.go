package replay

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAction11FlagAndEnvelope(t *testing.T) {
	for _, tt := range []struct {
		payload []byte
		kind    string
		flag    int
	}{
		{[]byte{1, 1, 0, 0}, "resign", 0},
		{[]byte{1, 1, 0, 1}, "action_11_candidate", 1},
		{[]byte{1, 1, 0, 2}, "action_11_candidate", 2},
		{[]byte{1, 1, 0}, "action_11_candidate", 0},
		{[]byte{1, 1, 1}, "action_11_candidate", 1},
		{[]byte{1, 2, 0, 0}, "action_11_candidate", -1},
		{[]byte{1}, "action_11_candidate", -1},
		{[]byte{}, "action_11_candidate", -1},
		{[]byte{9, 1, 0, 0}, "action_11_candidate", 0},
	} {
		e, ok := DecodeActionEvent(actionResign, tt.payload, 1000, 42, false)
		if !ok || e.Type != tt.kind {
			t.Fatalf("payload%x: %+v / %v", tt.payload, e, ok)
		}
		if tt.flag >= 0 && (e.Action11Flag == nil || *e.Action11Flag != tt.flag) {
			t.Fatal(e)
		}
		if tt.kind == "action_11_candidate" && e.RawHex != fmt.Sprintf("%x", tt.payload) {
			t.Fatalf("raw not preserved: %+v", e)
		}
	}
}

func TestCandidates11DoNotInferResignation(t *testing.T) {
	t.Setenv("AOE2KIT_REPLAY_CACHE", t.TempDir())
	var b bytes.Buffer
	healthTestMeta(&b)
	healthTestSync(&b, 1000, 1)
	for i := 0; i < 3; i++ {
		writeU32(&b, 1)
		writeU32(&b, 5)
		b.Write([]byte{11, 1, 1, 0, 1})
		writeU32(&b, uint32(i))
	}
	healthTestSync(&b, 1000, 2)
	healthTestChat(&b, "continuation")
	path := cacheFixture(t, 1024, b.Bytes())
	cache, err := OpenReplayCache(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, include := range []bool{false, true} {
		opts := EventOptions{IncludeUntypedAction: include}
		events, names, counts, warnings := ExtractActionStreamEvents(b.Bytes(), opts)
		got, gn, gc, gw, err := extractCachedEvents(cache, opts)
		if err != nil || !reflect.DeepEqual(events, got) || !reflect.DeepEqual(names, gn) || !reflect.DeepEqual(counts, gc) || !reflect.DeepEqual(warnings, gw) {
			t.Fatalf("cache parity: %v %+v %+v", err, counts, gc)
		}
		if counts.Resigns != 0 || counts.Candidates11 != 3 || counts.Actions != 3 || counts.Chat != 1 || counts.DurationMS != 2000 || len(warnings) != 1 {
			t.Fatalf("counts=%+v warnings=%v", counts, warnings)
		}
		result := InferResult([]FeedbackPlayer{{PlayerID: 1}, {PlayerID: 2}}, events, counts, warnings)
		if len(result.Resigned) != 0 || len(result.ResignEvents) != 0 || len(result.Losers) != 0 || result.WinnerKnown || result.Completed {
			t.Fatalf("candidate affected outcome: %+v", result)
		}
	}
	health, err := buildHealthBody(bytes.NewReader(b.Bytes()), HealthOptions{})
	if err != nil || health.EndAnchor.LastResign != nil {
		t.Fatalf("candidate affected last-resign anchor: %+v %v", health, err)
	}
}

func TestAction11PrivateCorpus(t *testing.T) {
	root := os.Getenv("AOE2KIT_ACTION11_CORPUS")
	if root == "" {
		t.Skip("AOE2KIT_ACTION11_CORPUS unset")
	}
	for _, tc := range []struct {
		name, hash                    string
		resigns, candidates, duration int
	}{
		{"mp-225157.aoe2record", "c2ff821096d67ae2e3c21d51b9b619f50378d439175fc5bd50cf1b6bdb2bb910", 1, 36, 311878},
		{"mp-230133.aoe2record", "ade2fa6445c36521e222a66fe2e1d9bc9485d6b4060595c16261adfc77cf88f1", 1, 9, 1883730},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, tc.name))
			if os.IsNotExist(err) {
				t.Skip("corpus recording missing")
			}
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%x", sha256.Sum256(raw)) != tc.hash {
				t.Fatal("corpus fingerprint mismatch")
			}
			_, _, counts, warnings := ExtractActionStreamEvents(raw[binary.LittleEndian.Uint32(raw[:4]):], EventOptions{})
			if counts.Resigns != tc.resigns || counts.Candidates11 != tc.candidates || counts.DurationMS != tc.duration || len(warnings) != 1 {
				t.Fatalf("counts=%+v warnings=%v", counts, warnings)
			}
		})
	}
}
