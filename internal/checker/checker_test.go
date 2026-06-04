package checker

import (
	"testing"

	"github.com/luismascotto/lotofacil-checker/internal/model"
)

func TestCountHits(t *testing.T) {
	bet := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	raffle := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 25}

	if got := CountHits(bet, raffle); got != 14 {
		t.Fatalf("CountHits() = %d, want 14", got)
	}
}

func TestCheckBetsSortedByHits(t *testing.T) {
	bets := []model.Bet{
		{Name: "low", Numbers: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}},
		{Name: "high", Numbers: []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}},
	}
	raffle := []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}

	results := CheckBets(bets, raffle)
	if results[0].Bet.Name != "high" || results[0].Hits != 15 {
		t.Fatalf("expected high first with 15 hits, got %+v", results[0])
	}
}
