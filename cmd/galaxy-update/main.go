// Package main provides the galaxy-update CLI.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/alecthomas/kong"
	"github.com/deadnews/galaxy-update/internal/galaxy"
)

var version = "dev"

var errHasErrors = errors.New("errors occurred")

// CLI updates versions to latest across the given files.
type CLI struct {
	Files   []string         `arg:"" optional:"" help:"Requirements files to update."`
	Verbose bool             `short:"v" help:"Show all entries, including current."`
	Version kong.VersionFlag `name:"version" short:"V" help:"Print version." hidden:""`
}

func main() {
	setupColors()
	var cli CLI
	kong.Parse(
		&cli,
		kong.Name("galaxy-update"),
		kong.Description("Update Ansible collection and role versions to latest."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true, FlagsLast: true}),
		kong.Vars{"version": version},
	)
	if err := execute(cli.Files, cli.Verbose); err != nil {
		if errors.Is(err, errHasErrors) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}

func execute(explicit []string, verbose bool) error {
	files, err := resolveFiles(explicit)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no files found")
		return nil
	}

	client := galaxy.NewClient()
	results, err := galaxy.Run(context.Background(), client, files)
	if err != nil {
		return fmt.Errorf("process: %w", err)
	}

	printResults(results, verbose)
	return exitStatus(results)
}

func resolveFiles(explicit []string) ([]string, error) {
	if len(explicit) > 0 {
		return explicit, nil
	}
	files, err := galaxy.DiscoverFiles(".")
	if err != nil {
		return nil, fmt.Errorf("discover files: %w", err)
	}
	return files, nil
}

// exitStatus returns errHasErrors when any entry failed to resolve.
func exitStatus(results []galaxy.Result) error {
	if slices.ContainsFunc(results, func(r galaxy.Result) bool {
		return r.Status == galaxy.StatusError
	}) {
		return errHasErrors
	}
	return nil
}
