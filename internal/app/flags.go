package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/luismascotto/lotofacil-checker/internal/model"
	"github.com/luismascotto/lotofacil-checker/internal/result"
)

// Options holds CLI paths resolved from flags and environment.
type Options struct {
	ConfigPath  string
	ResultsPath string
}

// ParseFlags parses -config and -results, then applies LOTOFACIL_CONFIG,
// LOTOFACIL_RESULTS, and built-in defaults when flags are omitted.
func ParseFlags(args []string) (Options, error) {
	fs := flag.NewFlagSet("lotofacil-checker", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configFlag := fs.String("config", "", "path to bets config JSON file")
	resultsFlag := fs.String("results", "", "path to past draw results CSV file")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Options{}, flag.ErrHelp
		}
		return Options{}, err
	}

	opts := Options{}

	opts.ConfigPath = *configFlag
	if opts.ConfigPath == "" {
		if path := os.Getenv("LOTOFACIL_CONFIG"); path != "" {
			opts.ConfigPath = path
		} else {
			path, err := defaultConfigPath()
			if err != nil {
				return Options{}, err
			}
			opts.ConfigPath = path
		}
	}

	opts.ResultsPath = *resultsFlag
	if opts.ResultsPath == "" {
		if path := os.Getenv("LOTOFACIL_RESULTS"); path != "" {
			opts.ResultsPath = path
		} else {
			opts.ResultsPath = result.DefaultResultsPath
		}
	}

	return opts, nil
}

func defaultConfigPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	name := fmt.Sprintf(model.ConfigFilenameLayout, time.Now().Format("20060102"))
	return filepath.Join(filepath.Dir(exe), name), nil
}
