package result

import (
	"context"
	"sync"

	"github.com/luismascotto/lotofacil-checker/internal/checker"
	"github.com/luismascotto/lotofacil-checker/internal/model"
)

// CheckPastResultsV4 checks past results using a bounded worker pool and sync.Pool
// for batch reuse. Use CheckPastResultsV4WithConfig to tune batch size and workers.
func CheckPastResultsV4(bets []model.Bet, minHits int, resultsPath string) ([]model.CheckPastResult, error) {
	if resultsPath == "" {
		resultsPath = DefaultResultsPath
	}
	cfg := PipelineConfig{ResultsPath: resultsPath}
	return CheckPastResultsV4WithConfig(bets, minHits, &cfg)
}

func CheckPastResultsV4WithConfig(bets []model.Bet, minHits int, cfg *PipelineConfig) ([]model.CheckPastResult, error) {
	pools := newPipelinePools(cfg)
	jobsCh := make(chan pastResultsJob)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		if err := pools.loadPastResults(ctx, jobsCh, cfg.ResultsPath, cfg.BatchSize); err != nil {
			cancel()
			errCh <- err
		}
	}()

	var (
		mu      sync.Mutex
		results []model.CheckPastResult
		wg      sync.WaitGroup
	)

	for range cfg.WorkerCount {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobsCh:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						pools.releasePastResultsBatch(job.batch)
						return
					default:
					}

					hits := checkResultsPooled(bets, job.batch, minHits, pools)
					pools.releasePastResultsBatch(job.batch)

					if len(hits) > 0 {
						mu.Lock()
						results = append(results, hits...)
						mu.Unlock()
					}
					pools.releaseHitsBatch(hits)
				}
			}
		})
	}

	wg.Wait()

	select {
	case err := <-errCh:
		drainPastResultJobs(pools, jobsCh)
		return nil, err
	default:
		return results, nil
	}
}

func drainPastResultJobs(p *pipelinePools, jobs <-chan pastResultsJob) {
	for job := range jobs {
		p.releasePastResultsBatch(job.batch)
	}
}

func checkResultsPooled(
	bets []model.Bet,
	pastResults []model.PastResult,
	minHits int,
	pool *pipelinePools,
) []model.CheckPastResult {
	out := pool.acquireHitsBatch()

	for _, pastResult := range pastResults {
		for _, bet := range bets {
			hits := checker.CountHits(bet.Numbers, *pastResult.Numbers)
			if hits < minHits {
				continue
			}
			//Clone pastResult.Numbers to avoid aliasing
			pastResultNumbersClone := make([]int, len(*pastResult.Numbers))
			copy(pastResultNumbersClone, *pastResult.Numbers)
			pastResultClone := pastResult
			pastResultClone.Numbers = &pastResultNumbersClone
			out = append(out, model.CheckPastResult{
				Bet:        bet,
				PastResult: pastResultClone,
				Hits:       hits,
			})
		}
	}

	return out
}
