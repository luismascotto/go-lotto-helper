package result

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/luismascotto/lotofacil-checker/internal/model"
)

const (
	DefaultResultsPath = "results.csv"
	defaultBatchSize   = 64
)

// PipelineConfig tunes batching, worker count, and input path for CheckPastResultsV4.
type PipelineConfig struct {
	BatchSize   int
	WorkerCount int
	ResultsPath string
	ResultSize  int
	HitSize     int
}

func (c *PipelineConfig) withDefaults() *PipelineConfig {
	if c.BatchSize <= 0 {
		c.BatchSize = defaultBatchSize
	}
	if c.WorkerCount <= 0 {
		c.WorkerCount = runtime.GOMAXPROCS(0)
	}
	if c.ResultsPath == "" {
		c.ResultsPath = DefaultResultsPath
	}
	if c.ResultSize <= 0 {
		c.ResultSize = model.RaffleSize
	}
	if c.HitSize <= 0 {
		c.HitSize = 8
	}
	return c
}

type pastResultsJob struct {
	batch []model.PastResult
}

type pipelinePools struct {
	pastBatches sync.Pool
	numbers     sync.Pool
	hitBatches  sync.Pool
}

func newPipelinePools(cfg *PipelineConfig) *pipelinePools {
	cfg = cfg.withDefaults()
	p := &pipelinePools{}
	p.pastBatches.New = func() any {
		s := make([]model.PastResult, 0, cfg.ResultSize)
		return &s
	}
	p.numbers.New = func() any {
		s := make([]int, 0, cfg.ResultSize)
		return &s
	}
	p.hitBatches.New = func() any {
		s := make([]model.CheckPastResult, 0, cfg.HitSize)
		return &s
	}
	return p
}

func (p *pipelinePools) acquirePastResultsBatch() []model.PastResult {
	s := *p.pastBatches.Get().(*[]model.PastResult)
	return s[:0]
}

func (p *pipelinePools) releasePastResultsBatch(batch []model.PastResult) {
	for i := range batch {
		if batch[i].Numbers == nil {
			continue
		}
		*batch[i].Numbers = (*batch[i].Numbers)[:0]
		p.numbers.Put(batch[i].Numbers)
		batch[i].Numbers = nil
	}
	batch = batch[:0]
	p.pastBatches.Put(&batch)
}

func (p *pipelinePools) acquireNumbers() *[]int {
	nums := p.numbers.Get().(*[]int)
	*nums = (*nums)[:0]
	return nums
}

func (p *pipelinePools) acquireHitsBatch() []model.CheckPastResult {
	s := *p.hitBatches.Get().(*[]model.CheckPastResult)
	return s[:0]
}

func (p *pipelinePools) releaseHitsBatch(batch []model.CheckPastResult) {
	batch = batch[:0]
	p.hitBatches.Put(&batch)
}

func (p *pipelinePools) loadPastResults(ctx context.Context, jobs chan<- pastResultsJob, path string, batchSize int) (retError error) {
	var batch []model.PastResult
	defer func() {
		//Always close the jobs channel
		close(jobs)

		//Only release the current batch manually if there was an error
		if retError != nil && len(batch) > 0 {
			p.releasePastResultsBatch(batch)
		}
	}()

	csvFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open results %q: %w", path, err)
	}
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	r.Comma = ';'
	r.FieldsPerRecord = 17
	if _, err := r.Read(); err != nil {
		return fmt.Errorf("read results header %q: %w", path, err)
	}

	batch = p.acquirePastResultsBatch()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read results %q: %w", path, err)
		}
		numbers := p.acquireNumbers()
		batch = append(batch, NewPastResultFromRecord(record, numbers))

		if len(batch) == batchSize {
			if err := p.sendPastResultsJob(ctx, jobs, batch); err != nil {
				return err
			}
			batch = p.acquirePastResultsBatch()
		}
	}

	if len(batch) > 0 {
		if err := p.sendPastResultsJob(ctx, jobs, batch); err != nil {
			return err
		}
	}

	return nil
}

func (p *pipelinePools) sendPastResultsJob(ctx context.Context, jobs chan<- pastResultsJob, batch []model.PastResult) error {
	select {
	case jobs <- pastResultsJob{batch: batch}:
		return nil
	case <-ctx.Done():
		p.releasePastResultsBatch(batch)
		return ctx.Err()
	}
}
