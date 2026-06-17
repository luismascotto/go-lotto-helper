package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/luismascotto/lotofacil-checker/internal/checker"
	"github.com/luismascotto/lotofacil-checker/internal/configstore"
	"github.com/luismascotto/lotofacil-checker/internal/input"
	"github.com/luismascotto/lotofacil-checker/internal/model"
	"github.com/luismascotto/lotofacil-checker/internal/result"
	"github.com/luismascotto/lotofacil-checker/internal/validate"
)

type App struct {
	store       *configstore.Store
	cfg         model.Config
	resultsPath string
}

func New(opts Options) *App {
	return &App{
		store:       configstore.New(opts.ConfigPath),
		resultsPath: opts.ResultsPath,
	}
}

func (a *App) Run() error {
	cfg, err := a.store.LoadOrCreate()
	if err != nil {
		return err
	}
	a.cfg = cfg

	for {
		choice, err := input.Prompt(a.menu())
		if err != nil {
			return err
		}

		//clear screen
		fmt.Print("\033[H\033[2J")

		switch strings.TrimSpace(choice) {
		case "1":
			a.listBets()
		case "2":
			if err := a.addBet(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "3":
			if err := a.removeBet(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "4":
			if err := a.checkResults(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "5":
			if err := a.checkPastResults(false); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "6":
			if err := a.addPastResult(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "7":
			if err := a.checkPastResults(true); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "0", "q", "Q":
			fmt.Println("Goodbye!")
			return nil
		default:
			fmt.Println("Invalid option. Choose 1-5.")
		}
		fmt.Println()
		//Enter to continue
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
	}
}

func (a *App) menu() string {
	return fmt.Sprintf(`
Lotofacil Checker
Config: %s
Registered bets: %d
Results path: %s

1) List bets
2) Add bet
3) Remove bet
4) Check your bets (enter raffle numbers)
5) Check your bets against past results (enter minimum number of hits)
6) Add a past result (enter raffle numbers)
7) Check a bet against past results (enter bet numbers andminimum number of hits)

0) Exit

Choose an option: `, a.store.Path(), len(a.cfg.Bets), a.resultsPath)
}

func (a *App) listBets() {
	if len(a.cfg.Bets) == 0 {
		fmt.Println("\nNo bets registered yet.")
		return
	}

	fmt.Println("\n--- Registered bets ---")
	for i, bet := range a.cfg.Bets {
		name := bet.Name
		if name == "" {
			name = fmt.Sprintf("Bet #%d", i+1)
		}
		fmt.Printf("[%d] %s (%d numbers): %s\n",
			i+1, name, len(bet.Numbers), input.FormatNumbers(bet.Numbers))
	}
	fmt.Println()
}

func (a *App) addBet() error {
	name, err := input.Prompt("\nBet name (optional): ")
	if err != nil {
		return err
	}

	fmt.Printf("Enter %d-%d numbers (01-25), separated by spaces or commas:\n",
		model.MinBetSize, model.MaxBetSize)

	numbers, err := input.ReadNumbers("Numbers", model.MinBetSize, false)
	if err != nil {
		return err
	}

	if err := validate.Bet(numbers); err != nil {
		return err
	}

	a.cfg.Bets = append(a.cfg.Bets, model.Bet{
		Name:    name,
		Numbers: numbers,
	})

	if err := a.store.Save(a.cfg); err != nil {
		return err
	}

	fmt.Println("Bet saved successfully.")
	return nil
}

func (a *App) removeBet() error {
	if len(a.cfg.Bets) == 0 {
		fmt.Println("\nNo bets to remove.")
		return nil
	}

	a.listBets()
	index, err := input.PromptInt("Index to remove (0 to cancel): ")
	if err != nil {
		return err
	}
	if index == 0 {
		fmt.Println("Cancelled.")
		return nil
	}
	if index < 1 || index > len(a.cfg.Bets) {
		return fmt.Errorf("invalid index %d", index)
	}

	removed := a.cfg.Bets[index-1]
	a.cfg.Bets = append(a.cfg.Bets[:index-1], a.cfg.Bets[index:]...)

	if err := a.store.Save(a.cfg); err != nil {
		return err
	}

	name := removed.Name
	if name == "" {
		name = fmt.Sprintf("Bet #%d", index)
	}
	fmt.Printf("Removed: %s\n", name)
	return nil
}

func (a *App) checkResults() error {
	if len(a.cfg.Bets) == 0 {
		fmt.Println("\nRegister at least one bet before checking results.")
		return nil
	}

	fmt.Printf("\nEnter the %d raffled numbers (01-25):\n", model.RaffleSize)
	raffle, err := input.ReadNumbers("Raffle", model.RaffleSize, true)
	if err != nil {
		return err
	}
	if err := validate.Raffle(raffle); err != nil {
		return err
	}

	results := checker.CheckBets(a.cfg.Bets, raffle)

	fmt.Println("\n========== Results ==========")
	fmt.Printf("Raffle: %s\n\n", input.FormatNumbers(raffle))

	for i, r := range results {
		name := r.Bet.Name
		if name == "" {
			name = fmt.Sprintf("Bet #%d", i+1)
		}
		fmt.Printf("[%s] %d hit(s) of %d\n",
			name, r.Hits, len(r.Bet.Numbers))
		fmt.Printf("  Numbers: %s\n", input.FormatNumbers(r.Bet.Numbers))
		if len(r.Missed) > 0 {
			fmt.Printf("  Missed:  %s\n", input.FormatNumbers(r.Missed))
		}
		fmt.Println()
	}

	return nil
}

func (a *App) checkPastResults(enterBetNumbers bool) error {
	if !enterBetNumbers && len(a.cfg.Bets) == 0 {
		fmt.Println("\nRegister at least one bet before checking past results.")
		return nil
	}
	var bets []model.Bet
	if !enterBetNumbers {
		bets = a.cfg.Bets
	}
	if enterBetNumbers {
		fmt.Printf("Enter %d-%d numbers (01-25), separated by spaces or commas:\n",
			model.MinBetSize, model.MaxBetSize)

		betNumbers, err := input.ReadNumbers("Bet numbers", model.RaffleSize, false)
		if err != nil {
			return err
		}
		bets = append(bets, model.Bet{
			Name:    "Custom bet",
			Numbers: betNumbers,
		})
	}

	minHits, err := input.PromptInt("Enter the minimum number of hits to check past results: ")
	if err != nil {
		return err
	}

	results, err := result.CheckPastResultsV4(bets, minHits, a.resultsPath)
	if err != nil {
		return err
	}

	fmt.Println("\n========== Past Results ==========")

	if len(results) == 0 {
		fmt.Println("No past results found.")
		return nil
	}
	//sort by hits descending, then by date descending
	sort.Slice(results, func(i, j int) bool {
		if results[i].Hits > results[j].Hits {
			return true
		}
		if results[i].Hits == results[j].Hits {
			return results[i].PastResult.Date > results[j].PastResult.Date
		}
		return false
	})

	for _, r := range results {
		fmt.Printf("[%s] %d hit(s) - %s\n",
			input.FormatDateUnix(r.PastResult.Date), r.Hits, r.Bet.Name)
		lineResult, lineBet := input.FormatResultWithBetNumbers(*r.PastResult.Numbers, r.Bet.Numbers)
		fmt.Printf("  Numbers: %s\n", lineResult)
		fmt.Printf("      Bet: %s\n", lineBet)
		fmt.Println()
	}

	return nil
}

func (a *App) addPastResult() error {
	//Read last result from results.csv
	lastResult, err := result.LoadLastPastResult(a.resultsPath)
	if err != nil {
		return err
	}
	fmt.Printf("\nLast result stored: ID %d - %s\n", lastResult.ID, input.FormatDateUnix(lastResult.Date))

	//Increment ID and ask for the date
	lastResult.ID++
	date, err := input.Prompt(fmt.Sprintf("Enter the date of the result ID %d (DD/MM/YYYY): ", lastResult.ID))
	if err != nil {
		return err
	}
	dateUnix := input.ParseDateUnix(date)
	if dateUnix == 0 || dateUnix <= lastResult.Date {
		return fmt.Errorf("invalid date")
	}
	lastResult.Date = dateUnix
	//Ask for the numbers
	fmt.Printf("\nEnter the %d raffled numbers (01-25)  separated by spaces or commas:\n", model.RaffleSize)
	numbers, err := input.ReadNumbers("Numbers", model.RaffleSize, true)
	if err != nil {
		return err
	}
	if err := validate.Raffle(numbers); err != nil {
		return err
	}
	lastResult.Numbers = &numbers
	//Save the result
	err = result.AppendPastResult(lastResult, a.resultsPath)
	if err != nil {
		return err
	}
	fmt.Printf("Result saved successfully. ID: %d - %s\n", lastResult.ID, input.FormatDateUnix(lastResult.Date))
	fmt.Printf("Numbers: %s\n", input.FormatNumbers(*lastResult.Numbers))
	return nil
}
