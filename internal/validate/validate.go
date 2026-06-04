package validate

import (
	"fmt"
	"sort"

	"github.com/luismascotto/lotofacil-checker/internal/model"
)

func Numbers(nums []int, minSize, maxSize int) error {
	if len(nums) < minSize || len(nums) > maxSize {
		return fmt.Errorf("expected between %d and %d numbers, got %d", minSize, maxSize, len(nums))
	}

	seen := make(map[int]struct{}, len(nums))
	for _, n := range nums {
		if n < model.MinNumber || n > model.MaxNumber {
			return fmt.Errorf("number %d is out of range (%02d-%02d)", n, model.MinNumber, model.MaxNumber)
		}
		if _, ok := seen[n]; ok {
			return fmt.Errorf("duplicate number: %02d", n)
		}
		seen[n] = struct{}{}
	}
	return nil
}

func Bet(numbers []int) error {
	return Numbers(numbers, model.MinBetSize, model.MaxBetSize)
}

func Raffle(numbers []int) error {
	return Numbers(numbers, model.RaffleSize, model.RaffleSize)
}

func Normalize(nums []int) []int {
	out := append([]int(nil), nums...)
	sort.Ints(out)
	return out
}
