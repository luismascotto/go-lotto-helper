package input

import (
	"strings"
	"testing"
)

func TestFormatResultWithBetNumbers(t *testing.T) {
	tests := []struct {
		name       string
		resultNums []int
		betNums    []int
		wantResult string
		wantBet    string
	}{
		{name: "equal", resultNums: []int{1, 2, 3, 4, 5}, betNums: []int{1, 2, 3, 4, 5}, wantResult: "01 02 03 04 05", wantBet: "01 02 03 04 05"},
		{name: "different", resultNums: []int{1, 2, 3, 4, 5}, betNums: []int{6, 7, 8, 9, 10},
			wantResult: "01 02 03 04 05 -- -- -- -- --", wantBet: "[] [] [] [] [] 06 07 08 09 10"},
		{name: "missing", resultNums: []int{1, 2, 3, 4, 5}, betNums: []int{2, 4, 6, 7, 8},
			wantResult: "01 02 03 04 05 -- -- --", wantBet: "[] 02 [] 04 [] 06 07 08"},
		{name: "extra", resultNums: []int{1, 2, 3, 4, 5}, betNums: []int{4, 5, 6, 7, 8, 9, 10},
			wantResult: "01 02 03 04 05 -- -- -- -- --", wantBet: "[] [] [] 04 05 06 07 08 09 10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, bet := FormatResultWithBetNumbers(tt.resultNums, tt.betNums)
			if strings.TrimSpace(result) != tt.wantResult {
				t.Errorf("FormatResultWithBetNumbers() result = %v, want %v", result, tt.wantResult)
			}
			if strings.TrimSpace(bet) != tt.wantBet {
				t.Errorf("FormatResultWithBetNumbers() bet = %v, want %v", bet, tt.wantBet)
			}
		})
	}
}
