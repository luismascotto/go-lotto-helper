package checker

import (
	"sort"

	"github.com/luismascotto/lotofacil-checker/internal/model"
)

func CountHits(orderedBetNumbers, orderedRaffleNumbers []int) int {
	hits := 0
	ixBet := 0
	ixRaffle := 0
	for ixBet < len(orderedBetNumbers) && ixRaffle < len(orderedRaffleNumbers) {
		if orderedBetNumbers[ixBet] == orderedRaffleNumbers[ixRaffle] {
			hits++
			ixBet++
			ixRaffle++
		} else if orderedBetNumbers[ixBet] < orderedRaffleNumbers[ixRaffle] {
			ixBet++
		} else {
			ixRaffle++
		}
	}
	return hits
}

func CheckBets(bets []model.Bet, raffle []int) []model.CheckResult {
	results := make([]model.CheckResult, len(bets))
	for i, bet := range bets {
		hits := CountHits(bet.Numbers, raffle)
		results[i] = model.CheckResult{
			Bet:    bet,
			Hits:   hits,
			Missed: missedNumbers(bet.Numbers, raffle),
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Hits != results[j].Hits {
			return results[i].Hits > results[j].Hits
		}
		return results[i].Bet.Name < results[j].Bet.Name
	})
	return results
}

func missedNumbers(betNumbers, raffle []int) []int {
	raffleSet := make(map[int]struct{}, len(raffle))
	for _, n := range raffle {
		raffleSet[n] = struct{}{}
	}

	var missed []int
	for _, n := range betNumbers {
		if _, ok := raffleSet[n]; !ok {
			missed = append(missed, n)
		}
	}
	sort.Ints(missed)
	return missed
}
