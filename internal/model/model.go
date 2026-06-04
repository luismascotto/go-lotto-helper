package model

const (
	MinBetSize           = 15
	MaxBetSize           = 20
	MinNumber            = 1
	MaxNumber            = 25
	RaffleSize           = 15
	ConfigFilenameLayout = "config-%s.json"
)

type Config struct {
	Bets []Bet `json:"bets"`
}

type Bet struct {
	Name    string `json:"name"`
	Numbers []int  `json:"numbers"`
}

type CheckResult struct {
	Bet    Bet
	Hits   int
	Missed []int
}

type PastResult struct {
	ID      int
	Date    int64
	Numbers *[]int
}

type Test struct {
	ID      int
	Date    int64
	Numbers [5]int
}

type CheckPastResult struct {
	Bet        Bet
	PastResult PastResult
	Hits       int
}
