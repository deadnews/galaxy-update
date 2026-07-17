package main

import (
	"fmt"
	"os"

	"github.com/deadnews/galaxy-update/internal/galaxy"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

var colorEnabled bool

func setupColors() {
	_, noColor := os.LookupEnv("NO_COLOR")
	colorEnabled = !noColor && isTerminal()
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// paint wraps s in the given color code when color output is enabled.
func paint(code, s string) string {
	if !colorEnabled {
		return s
	}
	return code + s + colorReset
}

// label renders text padded to a fixed width in the given color.
func label(code, text string) string {
	return paint(code, fmt.Sprintf("%-8s", text))
}

// printResults prints visible results grouped by file, one line per entry.
func printResults(results []galaxy.Result, verbose bool) {
	lastFile := ""
	for i := range results {
		r := &results[i]
		if !visible(r, verbose) {
			continue
		}
		if r.File != lastFile {
			if lastFile != "" {
				fmt.Println()
			}
			fmt.Println(paint(colorBold, r.File))
			lastFile = r.File
		}
		printResult(r)
	}
	if lastFile != "" {
		fmt.Println()
	}
}

// visible reports whether a result is shown at the given verbosity.
func visible(r *galaxy.Result, verbose bool) bool {
	switch r.Status {
	case galaxy.StatusUpdated, galaxy.StatusError:
		return true
	case galaxy.StatusCurrent, galaxy.StatusSkipped:
		return verbose
	}
	return false
}

func printResult(r *galaxy.Result) {
	switch r.Status {
	case galaxy.StatusUpdated:
		if r.Old == "" {
			fmt.Printf("  %s %s → %s\n", label(colorYellow, "UPDATED"), r.Name, r.New)
		} else {
			fmt.Printf("  %s %s %s → %s\n", label(colorYellow, "UPDATED"), r.Name, r.Old, r.New)
		}
	case galaxy.StatusCurrent:
		fmt.Printf("  %s %s %s\n", label(colorGreen, "OK"), r.Name, r.Old)
	case galaxy.StatusSkipped:
		fmt.Printf("  %s %s\n", label(colorDim, "SKIP"), r.Name)
	case galaxy.StatusError:
		fmt.Printf("  %s %s  %s\n", label(colorRed, "ERROR"), r.Name, paint(colorRed, r.Err.Error()))
	}
}
