package aifile

import "testing"

func TestReferenceConstantAvailability(t *testing.T) {
	r, err := LintPath("../../testdata/ai/undefined-duc.per")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, issue := range r.Issues {
		if issue.Code == "constant_availability_unverified" {
			count++
			if issue.Severity != "warning" || issue.Confidence != "unverified" {
				t.Fatal(issue)
			}
		}
	}
	if count != 5 || !r.OK {
		t.Fatalf("count=%d issues=%+v", count, r.Issues)
	}
	for _, text := range []string{
		"(defconst search-remote 2) (defconst object-data-distance 44) (defconst search-order-asc 1) (defconst formation-line 2) (defconst stance-aggressive 0) (defrule (true) => (up-clean-search search-remote object-data-distance search-order-asc) (up-target-objects 1 2 formation-line stance-aggressive))",
		"(defrule (true) => (up-clean-search 2 44 1) (up-target-objects 1 2 2 0) (cc-add-resource food 1))",
	} {
		r := lintPERText(t, text)
		if !r.OK || len(r.Issues) != 0 {
			t.Fatalf("issues=%+v", r.Issues)
		}
	}
	r = lintPERText(t, "(defconst alias search-remote) (defrule (true) => (up-clean-search alias -1 0))")
	issueWithCode(t, r, "constant_availability_unverified")
}
