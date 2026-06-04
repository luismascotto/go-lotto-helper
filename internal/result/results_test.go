package result

import (
	"os"
	"sort"
	"testing"

	"github.com/luismascotto/lotofacil-checker/internal/model"
)

func TestCheckPastResultsV4ReturnsErrorWhenResultsMissing(t *testing.T) {
	//t.Parallel()

	results, err := CheckPastResultsV4(nil, 12, "nonexistent-results.csv")
	if err == nil {
		t.Fatal("expected error for missing results file")
	}
	if results != nil {
		t.Fatalf("expected nil results on error, got len=%d", len(results))
	}
}

func TestCheckPastResultsV4MatchesV1(t *testing.T) {
	if _, err := os.Stat("results.csv"); err != nil {
		t.Skip("results.csv not found")
	}

	bets := []model.Bet{
		{
			Name:    "low",
			Numbers: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			Name:    "high",
			Numbers: []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
		},
	}
	minHits := 12

	gotRaw, err := CheckPastResultsV4(bets, minHits, DefaultResultsPath)
	if err != nil {
		t.Fatal(err)
	}
	got := normalizeCheckPastResults(gotRaw)
	wantRaw, err := CheckPastResultsV1(bets, minHits, DefaultResultsPath)
	if err != nil {
		t.Fatal(err)
	}
	want := normalizeCheckPastResults(wantRaw)

	if len(got) != len(want) {
		t.Fatalf("result count mismatch: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if !checkPastResultEqual(got[i], want[i]) {
			t.Fatalf("result[%d] mismatch:\ngot  %+v\nwant %+v", i, got[i], want[i])
		}
	}
}

func checkPastResultEqual(a, b model.CheckPastResult) bool {
	if a.Hits != b.Hits {
		return false
	}
	if a.Bet.Name != b.Bet.Name {
		return false
	}
	if len(a.Bet.Numbers) != len(b.Bet.Numbers) {
		return false
	}
	for i := range a.Bet.Numbers {
		if a.Bet.Numbers[i] != b.Bet.Numbers[i] {
			return false
		}
	}
	if a.PastResult.ID != b.PastResult.ID || a.PastResult.Date != b.PastResult.Date {
		return false
	}
	if len(*a.PastResult.Numbers) != len(*b.PastResult.Numbers) {
		return false
	}
	for i := range *a.PastResult.Numbers {
		if (*a.PastResult.Numbers)[i] != (*b.PastResult.Numbers)[i] {
			return false
		}
	}
	return true
}

func normalizeCheckPastResults(results []model.CheckPastResult) []model.CheckPastResult {
	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.PastResult.ID != b.PastResult.ID {
			return a.PastResult.ID < b.PastResult.ID
		}
		if a.Bet.Name != b.Bet.Name {
			return a.Bet.Name < b.Bet.Name
		}
		return a.Hits > b.Hits
	})
	return results
}
