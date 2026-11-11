package replay

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aoe2kit/pkg/triggergraph"
)

func modernEmbeddedFixture() []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, math.Float64bits(5))
	b.WriteByte(0)
	writeU32(&b, 2)
	for i := 0; i < 2; i++ {
		prefix := make([]byte, 27)
		binary.LittleEndian.PutUint32(prefix, 1)
		b.Write(prefix)
		for _, s := range []string{strings.Repeat("x", 300000), fmt.Sprintf("Synthetic %d", i), ""} {
			writeU32(&b, uint32(len(s)+1))
			b.WriteString(s)
			b.WriteByte(0)
		}
		writeU32(&b, 0) // effects
		writeU32(&b, 1) // conditions
		words := make([]byte, 36*4)
		binary.LittleEndian.PutUint32(words[4:], 34)
		b.Write(words)
		writeU32(&b, 1)
		b.WriteByte(0)  // xs string
		writeU32(&b, 0) // condition order
	}
	writeU32(&b, 1)
	writeU32(&b, 0) // complete trigger order
	return b.Bytes()
}

func TestModernEmbeddedGraphBeyondStringHeuristics(t *testing.T) {
	section := modernEmbeddedFixture()
	header := append(make([]byte, 128), section...)
	g, region, err := parseTriggerGraphWithEmbeddedFallback(header)
	if err != nil || g.TriggerCount != 2 || g.ConditionCount != 2 || region.End != len(header) || classifyTriggerGraphTier(g, region) != "triggergraph_count_and_order_validated" {
		t.Fatalf("graph=%+v region=%+v err=%v", g, region, err)
	}
	bad := append([]byte(nil), section...)
	binary.LittleEndian.PutUint32(bad[len(bad)-4:], 1)
	g, _, err = parseEmbedded159TriggerGraph(bad)
	if err == nil && classifyTriggerGraphTier(g, nil) == "triggergraph_count_and_order_validated" {
		t.Fatal("invalid permutation validated")
	}
}

func TestSlotNumberColorRemainDistinctInFeedback(t *testing.T) {
	slots := []PlayerSlot{{Slot: 1, Number: 7, Color: 5, Name: "First"}, {Slot: 7, Number: 2, Color: 0, Name: "Second"}}
	players := feedbackPlayers(slots, map[int]string{7: "Second", 2: "Second"}, nil)
	if len(players) != 2 || players[0].PlayerID != 2 || players[0].Name != "Second" || players[1].PlayerID != 7 || players[1].Name != "First" {
		t.Fatal(players)
	}
	events := []ReplayEvent{{PlayerID: 7, PlayerName: "Second"}, {PlayerID: 2}}
	fillEventPlayerNames(events, players)
	if events[0].PlayerName != "First" || events[1].PlayerName != "Second" {
		t.Fatal(events)
	}
}

func TestLobbyTeamContextDoesNotClaimRuntimeAlliances(t *testing.T) {
	rec := &File{TriggerGraphTier: "triggergraph_count_and_order_validated", TriggerGraph: &triggergraph.Graph{Triggers: []map[string]any{{"effects": []map[string]any{{"type_name": "change_diplomacy"}}}}}}
	c := replayTeamContext(rec)
	if c.Source != "lobby_picks" || c.InGameAlliancesVerified || c.DiplomacyEffects != 1 || len(c.Warnings) != 2 {
		t.Fatal(c)
	}
	raw, _ := json.Marshal(rec)
	var cached File
	if err := json.Unmarshal(raw, &cached); err != nil {
		t.Fatal(err)
	}
	if got := replayTeamContext(&cached); got.DiplomacyEffects != 1 {
		t.Fatal(got)
	}
	c = replayTeamContext(&File{TriggerGraphTier: "triggergraph_bounded_candidate"})
	if c.InGameAlliancesVerified || len(c.Warnings) != 2 {
		t.Fatal(c)
	}
}

func TestMPIdentityGraphCorpus(t *testing.T) {
	root := os.Getenv("AOE2KIT_MP_REPLAY_CORPUS")
	if root == "" {
		t.Skip("AOE2KIT_MP_REPLAY_CORPUS unset")
	}
	path := filepath.Join(root, "mp-165753.aoe2record")
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skip("corpus recording missing")
	}
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "ed9b4882e14bcabbd2edcf32e80a9e7bb0c083f87c88d2b731f03534858e6802" {
		t.Fatal("corpus fingerprint mismatch")
	}
	rec, err := OpenMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	if rec.TriggerGraph == nil || rec.TriggerGraph.TriggerCount != 1018 || rec.TriggerGraphTier != "triggergraph_count_and_order_validated" {
		t.Fatalf("graph tier=%s graph=%+v", rec.TriggerGraphTier, rec.TriggerGraph)
	}
	summary, err := BuildSummary(path)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TeamContext.InGameAlliancesVerified || summary.TeamContext.Source != "lobby_picks" {
		t.Fatal(summary.TeamContext)
	}
	feedback, err := ExtractFeedback(path)
	if err != nil {
		t.Fatal(err)
	}
	names := map[int]string{}
	for _, p := range rec.Players {
		names[p.Number] = p.Name
	}
	for _, p := range feedback.Players {
		if name := names[p.PlayerID]; name != "" && p.Name != name {
			t.Fatalf("player-number roster mismatch: %+v", p)
		}
	}
}
