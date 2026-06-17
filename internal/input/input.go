package input

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/luismascotto/lotofacil-checker/internal/validate"
)

var reader = bufio.NewReader(os.Stdin)

func Prompt(label string) (string, error) {
	fmt.Print(label)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func PromptInt(label string) (int, error) {
	for {
		text, err := Prompt(label)
		if err != nil {
			return 0, err
		}
		n, err := strconv.Atoi(text)
		if err != nil {
			fmt.Println("Enter a valid number.")
			continue
		}
		return n, nil
	}
}

func ReadNumbers(label string, expected int, strict bool) ([]int, error) {
	hint := ""
	if expected > 0 {
		if !strict {
			hint = fmt.Sprintf(" (min %d numbers)", expected)
		} else {
			hint = fmt.Sprintf(" (%d numbers)", expected)
		}
	}
	for {
		text, err := Prompt(label + hint + ": ")
		if err != nil {
			return nil, err
		}
		if text == "" {
			return nil, fmt.Errorf("no numbers provided")
		}
		nums, err := parseNumbers(text)
		if err != nil {
			fmt.Printf("Invalid input: %v\n", err)
			continue
		}
		if expected > 0 && strict && len(nums) != expected {
			fmt.Printf("Enter exactly %d numbers. Got %d.\n", expected, len(nums))
			continue
		}
		if expected > 0 && !strict && len(nums) < expected {
			fmt.Printf("Enter at least %d numbers. Got %d.\n", expected, len(nums))
			continue
		}
		return validate.Normalize(nums), nil
	}
}

func parseNumbers(text string) ([]int, error) {
	text = strings.ReplaceAll(text, ",", " ")
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return nil, fmt.Errorf("no numbers provided")
	}

	nums := make([]int, 0, len(fields))
	for _, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q", field)
		}
		nums = append(nums, n)
	}
	return nums, nil
}

func FormatNumbers(nums []int) string {
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = fmt.Sprintf("%02d", n)
	}
	return strings.Join(parts, " ")
}

func FormatResultWithBetNumbers(resultNums []int, betNums []int) (result string, bet string) {
	var strbResult strings.Builder
	var strbBet strings.Builder
	b := 0
	r := 0
	for b < len(betNums) || r < len(resultNums) {
		if b < len(betNums) && r < len(resultNums) && betNums[b] == resultNums[r] {
			fmt.Fprintf(&strbResult, "%02d ", resultNums[r])
			fmt.Fprintf(&strbBet, "%02d ", betNums[b])
			b++
			r++
		} else if r == len(resultNums) || betNums[b] < resultNums[r] {
			fmt.Fprintf(&strbResult, "-- ")
			fmt.Fprintf(&strbBet, "%02d ", betNums[b])
			b++
		} else if b == len(betNums) || resultNums[r] < betNums[b] {
			fmt.Fprintf(&strbResult, "%02d ", resultNums[r])
			fmt.Fprintf(&strbBet, "[] ")
			r++
		}
	}
	return strbResult.String(), strbBet.String()
}

// ParseDateUnix parses DD/MM/YYYY from CSV into Unix seconds at UTC midnight.
func ParseDateUnix(date_DD_MM_YYYY string) int64 {
	t, err := time.Parse("02/01/2006", date_DD_MM_YYYY)
	if err != nil {
		return 0
	}
	return t.UTC().Unix()
}

// FormatDateUnix formats Unix seconds as DD/MM/YYYY.
func FormatDateUnix(unix int64) string {
	return time.Unix(unix, 0).UTC().Format("02/01/2006")
}

func Atoi(s string) int {
	n, _ := strconv.Atoi(s)

	return n
}
