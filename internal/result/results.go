package result

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/luismascotto/lotofacil-checker/internal/checker"
	"github.com/luismascotto/lotofacil-checker/internal/input"
	"github.com/luismascotto/lotofacil-checker/internal/model"
)

// CheckPastResultsV1 checks the past results for the bets and returns the results sorted by hits and date
//
//	V1: read all records first and then check the bets
//	V2: read chunks of N records and send to a channel
func CheckPastResultsV1(bets []model.Bet, minHits int, resultsPath string) ([]model.CheckPastResult, error) {
	if resultsPath == "" {
		resultsPath = DefaultResultsPath
	}
	pastResults, err := loadAllPastResults(resultsPath)
	if err != nil {
		return nil, err
	}
	results := make([]model.CheckPastResult, 0)
	results = checkResults(bets, pastResults, minHits, results)

	return results, nil
}

func checkResults(bets []model.Bet, pastResults []model.PastResult, minHits int, results []model.CheckPastResult) []model.CheckPastResult {
	//check past results for the bet
	for _, pastResult := range pastResults {
		results = checkBets(bets, pastResult, minHits, results)
	}
	return results
}

func checkBets(bets []model.Bet, pastResult model.PastResult, minHits int, results []model.CheckPastResult) []model.CheckPastResult {
	for _, bet := range bets {
		hits := checker.CountHits(bet.Numbers, *pastResult.Numbers)
		if hits >= minHits {
			results = append(results, model.CheckPastResult{
				Bet:        bet,
				PastResult: pastResult,
				Hits:       hits,
			})
		}
	}
	return results
}

func loadAllPastResults(path string) ([]model.PastResult, error) {
	csvFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	results := make([]model.PastResult, 0)
	defer csvFile.Close()
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = 17
	//skip header
	csvReader.Read()
	for {
		//read record
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		numbers := make([]int, 0, 15)
		results = append(results, NewPastResultFromRecord(record, &numbers))
	}
	return results, nil
}

func LoadLastPastResult(path string) (model.PastResult, error) {
	csvFile, err := os.Open(path)
	if err != nil {
		return model.PastResult{}, err
	}
	defer csvFile.Close()
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = 17
	var lastRecord []string
	csvReader.Read()

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return model.PastResult{}, err
		}
		lastRecord = record
	}
	if len(lastRecord) == 0 {
		return model.PastResult{}, fmt.Errorf("no last record found")
	}
	numbers := make([]int, 0, 15)
	return NewPastResultFromRecord(lastRecord, &numbers), nil
}

func AppendPastResult(result model.PastResult, path string) error {

	csvFile, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.Comma = ';'
	csvWriter.UseCRLF = true

	record := make([]string, 0, 17)
	record = append(record, strconv.Itoa(result.ID))
	record = append(record, input.FormatDateUnix(result.Date))
	for _, number := range *result.Numbers {
		record = append(record, strconv.Itoa(number))
	}
	csvWriter.Write(record)
	if err := csvWriter.Error(); err != nil {
		return err
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return err
	}
	return nil
}

func LoadNPastResult(n int, path string) ([]model.PastResult, error) {
	if n <= 0 {
		return loadAllPastResults(path)
	}
	csvFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = 17
	csvReader.Read()

	ringRecords := make([][]string, n)
	count := 0

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ringRecords[count%n] = record
		count++

	}
	if count == 0 {
		return nil, fmt.Errorf("no last record found")
	}
	found := min(count, n)

	results := make([]model.PastResult, 0, found)
	for i := range found {
		//Add the records in reverse order
		record := ringRecords[(count-i-1)%n]

		numbers := make([]int, 0, 15)
		results = append(results, NewPastResultFromRecord(record, &numbers))
	}

	return results, nil
}

func NewPastResultFromRecord(record []string, numbers *[]int) model.PastResult {
	return model.PastResult{
		ID:      input.Atoi(record[0]),
		Date:    input.ParseDateUnix(record[1]),
		Numbers: strSliceToPtrInts(record[2:17], numbers),
	}
}

func strSliceToPtrInts(strNumbers []string, numbers *[]int) *[]int {

	for _, number := range strNumbers {
		*numbers = append(*numbers, input.Atoi(number))
	}
	return numbers
}
