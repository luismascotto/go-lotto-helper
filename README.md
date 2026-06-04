# lotofacil-checker

CLI to manage recurring Lotofácil bets and check how many numbers matched after a draw.

## Rules

- Each bet has **15 to 20** numbers in the range **01–25**
- A draw has exactly **15** raffled numbers
- Hits = how many of your bet numbers appear in the draw

## Quick start

```bash
cd lotofacil-checker
go build -o lotofacil-checker.exe ./cmd/lotofacil-checker
./lotofacil-checker.exe
```

On first run (without `-config` or `LOTOFACIL_CONFIG`), an empty config file is created next to the executable as `config-yyyyMMdd.json` (for example `config-20260527.json`).

Copy the example config if you prefer to edit bets by hand:

```bash
copy config.example.json config-20260527.json
```

## Menu

1. **List bets** — show all registered bets
2. **Add bet** — name (optional) + 15–20 numbers
3. **Remove bet** — delete by index
4. **Check results** — enter 15 draw numbers and see hits per bet
5. **Exit**

## Config file

Default path: `config-yyyyMMdd.json` in the same folder as the executable (created automatically if missing). Multiple runs on the same day reuse the same file.

Any explicit path (`-config` or `LOTOFACIL_CONFIG`) is also created automatically when missing.

```json
{
  "bets": [
    {
      "name": "Aposta principal",
      "numbers": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15]
    }
  ]
}
```

Custom paths:

```bash
./lotofacil-checker.exe -config C:\path\to\my-bets.json -results C:\path\to\results.csv
```

Or set `LOTOFACIL_CONFIG` and `LOTOFACIL_RESULTS`. Past draws default to `results.csv` in the working directory.

## Input format

Numbers can be typed with spaces or commas:

```
01 02 03 04 05 06 07 08 09 10 11 12 13 14 15
```

Single-digit numbers (`1`) and zero-padded (`01`) are both accepted.

## Architecture

Checking bets against historical draws (`results.csv`) uses a parallel pipeline with bounded workers and `sync.Pool` batch reuse. See [docs/result-pipeline.md](docs/result-pipeline.md) for flow diagrams, pool lifecycle, and tuning (`BatchSize`, `WorkerCount`).
