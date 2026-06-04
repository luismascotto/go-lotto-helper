package validate

import "testing"

func TestBet(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		wantErr bool
	}{
		{"valid 15", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, false},
		{"valid 20", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, false},
		{"too few", []int{1, 2, 3}, true},
		{"too many", make([]int, 21), true},
		{"out of range", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, true},
		{"duplicate", []int{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "too many" {
				for i := range tt.numbers {
					tt.numbers[i] = i + 1
				}
			}
			err := Bet(tt.numbers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Bet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
