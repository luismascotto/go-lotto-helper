package result

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/luismascotto/lotofacil-checker/internal/model"
)

func TestMain(m *testing.M) {
	for _, dir := range []string{".", "..", "../.."} {
		if _, err := os.Stat(filepath.Join(dir, "results.csv")); err == nil {
			_ = os.Chdir(dir)
			break
		}
	}
	os.Exit(m.Run())
}

func benchmarkBets() []model.Bet {
	return []model.Bet{
		{
			Name:    "low",
			Numbers: []int{11, 2, 3, 4, 5, 6, 7, 8, 9, 12, 13, 14, 17, 19, 21},
		},
		{
			Name:    "high",
			Numbers: []int{5, 6, 7, 9, 10, 12, 13, 15, 17, 20, 21, 22, 23, 24, 25},
		},
	}
}

func requireResultsCSV(b *testing.B) {
	b.Helper()
	if _, err := os.Stat("results.csv"); err != nil {
		b.Skip("results.csv not found (run from repo root or ensure file exists)")
	}
}

// func BenchmarkCheckPastResultsV1(b *testing.B) {
// 	requireResultsCSV(b)
// 	bets := benchmarkBets()
// 	minHits := 12
// 	b.ReportAllocs()

// 	for b.Loop() {
// 		_, _ = CheckPastResultsV1(bets, minHits, DefaultResultsPath)
// 	}
// }

func benchmarkCheckPastResultsV4WithConfig_Parameters(b *testing.B, batchSize int, workerCount int) {
	requireResultsCSV(b)
	bets := benchmarkBets()
	minHits := 11
	b.ReportAllocs()
	cfg := PipelineConfig{BatchSize: batchSize, WorkerCount: workerCount}
	for i := 0; i < b.N; i++ {
		_, _ = CheckPastResultsV4WithConfig(bets, minHits, &cfg)
	}
}

func BenchmarkCheckPastResultsV4WithConfig_200(b *testing.B) {
	benchmarkCheckPastResultsV4WithConfig_Parameters(b, 200, runtime.GOMAXPROCS(0))
}

func BenchmarkCheckPastResultsV4WithConfig_100(b *testing.B) {
	benchmarkCheckPastResultsV4WithConfig_Parameters(b, 100, runtime.GOMAXPROCS(0))
}

func BenchmarkCheckPastResultsV4WithConfig_60(b *testing.B) {
	benchmarkCheckPastResultsV4WithConfig_Parameters(b, 60, runtime.GOMAXPROCS(0))
}

func BenchmarkCheckPastResultsV4WithConfig_32(b *testing.B) {
	benchmarkCheckPastResultsV4WithConfig_Parameters(b, 32, runtime.GOMAXPROCS(0))
}
