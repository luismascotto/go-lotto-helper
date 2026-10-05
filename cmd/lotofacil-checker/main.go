package main

import (
	"fmt"
	"os"

	"github.com/luismascotto/lotofacil-checker/internal/app"
)

func main() {
	opts, err := app.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	a := app.New(opts)

	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
