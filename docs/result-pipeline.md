# Past-results check pipeline (V4)

`CheckPastResultsV4WithConfig` in `internal/result` checks every configured bet against historical draws loaded from a CSV file. It uses a **producer–consumer** design: one goroutine streams batched past results into a bounded channel; a fixed pool of workers compares each batch against all bets and merges hits into a shared slice.

Entry points:

- `CheckPastResultsV4` — wraps `CheckPastResultsV4WithConfig` with default `PipelineConfig` and `results.csv`
- `CheckPastResultsV4WithConfig` — tunable batch size, worker count, and paths

Configuration (`PipelineConfig`, see `internal/result/pipeline.go`):

| Field | Default | Role |
|-------|---------|------|
| `BatchSize` | 256 | Past draws per job |
| `WorkerCount` | `runtime.GOMAXPROCS(0)` | Parallel workers |
| `ResultsPath` | `results.csv` | CSV input |
| `ResultSize` | `model.RaffleSize` | Initial cap for `[]PastResult` pool slices |
| `HitSize` | 8 | Initial cap for `[]CheckPastResult` pool slices |

## High-level flow

```mermaid
flowchart TB
    subgraph entry["CheckPastResultsV4WithConfig"]
        CFG["PipelineConfig"]
        POOLS["newPipelinePools(cfg)\n3× sync.Pool"]
        JOBS["jobsCh\nchan pastResultsJob\ncap = WorkerCount"]
        CTX["context.WithCancel"]
        ERR["errCh buffer 1"]
        RES["results + sync.Mutex"]
    end

    CFG --> POOLS
    POOLS --> JOBS

    subgraph producer["Producer — 1 goroutine"]
        LOAD["loadPastResults"]
        CSV["Read CSV → PastResult rows"]
        SEND["sendPastResultsJob → jobsCh"]
        CLOSE["defer close(jobsCh)"]
    end

    subgraph workers["Workers — WorkerCount goroutines"]
        W1["Worker"]
        W2["Worker"]
        WN["Worker …"]
    end

    CTX --> LOAD
    LOAD --> CSV --> SEND --> JOBS
    LOAD --> CLOSE

    JOBS --> W1 & W2 & WN
    W1 & W2 & WN --> CHECK["checkResultsPooled"]
    CHECK --> MERGE["mu.Lock → append hits"]
    MERGE --> RES

    LOAD -.->|on error| ERR
    ERR -.->|cancel| CTX
```

## Parallel processing

| Role | Count | Responsibility |
|------|--------|----------------|
| Loader | 1 | Streams CSV, fills batches of `BatchSize`, enqueues `pastResultsJob` |
| Workers | `WorkerCount` | Per job: compare batch × all bets, merge hits under mutex |
| Channel | `jobsCh` cap = `WorkerCount` | Backpressure when workers are busy |

Workers exit when `jobsCh` is closed and drained, or when `ctx` is cancelled (loader error or shutdown).

## Pool architecture

Three `sync.Pool` instances on `pipelinePools` reuse slices and cut allocations during CSV load and per-job checking.

```mermaid
flowchart LR
    subgraph pools["pipelinePools"]
        PB["pastBatches\n[]PastResult"]
        NP["numbers\n[]int"]
        HB["hitBatches\n[]CheckPastResult"]
    end

    subgraph loader["loadPastResults"]
        L1["acquirePastResultsBatch"]
        L2["acquireNumbers per row"]
        L3["append → batch"]
        L4["full batch → jobsCh"]
    end

    subgraph worker["per job"]
        W1["checkResultsPooled"]
        W2["acquireHitsBatch"]
        W3["CountHits bet × draw"]
        W4["releasePastResultsBatch"]
        W5["releaseHitsBatch"]
    end

    PB --> L1
    NP --> L2
    L2 --> L3 --> L4
    L4 -->|pastResultsJob| W1
    HB --> W2
    W1 --> W2 --> W3 --> W4 --> W5
    W4 -->|Numbers → numbers pool\nbatch → pastBatches| PB
    W4 --> NP
    W5 --> HB
```

**Lifecycle**

1. **Loader** — `acquirePastResultsBatch` builds a batch; each CSV row uses `acquireNumbers` for the 15 draw numbers. When `len(batch) == BatchSize`, the batch is sent as a job and a fresh batch is acquired.
2. **Worker** — `acquireHitsBatch` collects matches; `releasePastResultsBatch` returns number slices to `numbers` and clears the batch back to `pastBatches`; `releaseHitsBatch` clears the hits slice back to `hitBatches` (via `defer` after merge).
3. **Copy on hit** — matching rows copy `PastResult.Numbers` so pooled slices are not aliased after release.

## Per-job sequence

```mermaid
sequenceDiagram
    participant L as Loader
    participant J as jobsCh
    participant W as Worker
    participant P as pipelinePools
    participant M as results + Mutex

    L->>P: acquirePastResultsBatch, acquireNumbers
    L->>J: pastResultsJob{batch}
    J->>W: receive job
    W->>P: acquireHitsBatch
    W->>W: ∀ draw in batch, ∀ bet: CountHits
    W->>P: releasePastResultsBatch(batch)
    alt len(hits) > 0
        W->>M: Lock, append, Unlock
    end
    W->>P: releaseHitsBatch(hits)
    L->>J: close after EOF
```

## Error and cancellation

```mermaid
flowchart TD
    E["loadPastResults error or ctx cancelled"]
    E --> C["cancel(ctx)"]
    E --> EC["errCh <- err"]
    C --> WSTOP["workers exit on ctx.Done()\nor release batch if cancelled mid-job"]
    WG["wg.Wait()"]
    WG --> SEL{err on errCh?}
    SEL -->|yes| DRAIN["drainPastResultJobs\nrelease remaining batches"]
    DRAIN --> FAIL["return nil, err"]
    SEL -->|no| OK["return results, nil"]
```

If the loader fails, `drainPastResultJobs` consumes any jobs still on `jobsCh` and calls `releasePastResultsBatch` so pooled memory is not leaked.

## Data plane (one job)

```
CSV row → acquireNumbers() → PastResult.Numbers
                ↓
     batch[0 .. BatchSize-1] → jobsCh → Worker
                ↓
  ∀ pastResult ∈ batch, ∀ bet ∈ bets:
       hits = CountHits(bet, pastResult)
       if hits >= minHits → append to hits slice (copy Numbers)
                ↓
       merge into shared results (mutex)
```

## Source map

| Concern | File |
|---------|------|
| Orchestration, workers, `checkResultsPooled` | `internal/result/results_parallel.go` |
| Pools, CSV loader, `PipelineConfig` | `internal/result/pipeline.go` |
| Hit counting | `internal/checker` |

Benchmarks: `BenchmarkCheckPastResultsV4WithConfig_*` in `internal/result/results_bench_test.go`.
