package enginefacts

import "testing"

func TestLoadDefaultFactsLedger(t *testing.T) {
	ledger, err := LoadDefault()
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger.Facts) < 10 {
		t.Fatalf("facts = %d, want seeded ledger", len(ledger.Facts))
	}
	fact, ok := ledger.ByID("text.trigger_color_first_tag_only")
	if !ok {
		t.Fatal("missing text.trigger_color_first_tag_only")
	}
	if fact.Tier != EngineVerified || fact.VerifiedDate == "" || fact.FixtureRef == "" {
		t.Fatalf("fact citation incomplete: %+v", fact)
	}
	if len(ledger.Filter(false, "")) >= len(ledger.Facts) {
		t.Fatalf("engine-verified filter did not hide provisional facts")
	}
	if len(ledger.Filter(true, "xs")) == 0 {
		t.Fatalf("domain filter xs returned no facts")
	}
}
