// Package galaxy updates Ansible collection and role versions to the latest release.
package galaxy

import (
	"cmp"
	"context"
	"errors"
)

var errNotResolved = errors.New("not resolved")

// Status classifies the outcome of processing one entry.
type Status int

// Processing outcomes.
const (
	StatusUpdated Status = iota
	StatusCurrent
	StatusSkipped
	StatusError
)

// Result is the outcome of processing one entry.
type Result struct {
	File   string
	Name   string
	Old    string
	New    string
	Status Status
	Err    error
}

// Run updates every galaxy entry to its latest version in-place.
func Run(ctx context.Context, c *Client, files []string) ([]Result, error) {
	parsed, err := parseAllFiles(files)
	if err != nil {
		return nil, err
	}

	resolved := c.resolve(ctx, refsToResolve(parsed))

	var results []Result
	for i := range parsed {
		fd := &parsed[i]
		fileResults, changed := classifyFile(fd, resolved)
		results = append(results, fileResults...)
		if changed {
			if err := writeFile(fd); err != nil {
				return nil, err
			}
		}
	}
	return results, nil
}

// refsToResolve collects the distinct galaxy references to look up.
func refsToResolve(parsed []fileData) []ref {
	var refs []ref
	seen := make(map[ref]bool)
	for _, fd := range parsed {
		for _, e := range fd.entries {
			r := ref{kind: e.kind, name: e.name}
			if e.skip || seen[r] {
				continue
			}
			seen[r] = true
			refs = append(refs, r)
		}
	}
	return refs
}

func classifyFile(fd *fileData, resolved map[ref]lookup) (results []Result, changed bool) {
	for i := range fd.entries {
		e := &fd.entries[i]
		result := classifyEntry(fd.path, e, resolved)
		if result.Status == StatusUpdated {
			e.setVersion(result.New)
			changed = true
		}
		results = append(results, result)
	}
	return results, changed
}

func classifyEntry(path string, e *entry, resolved map[ref]lookup) Result {
	result := Result{File: path, Name: e.name, Old: e.version}
	if e.skip {
		result.Status = StatusSkipped
		return result
	}

	l := resolved[ref{kind: e.kind, name: e.name}]
	if l.err != nil || l.version == "" {
		result.Status = StatusError
		result.Err = cmp.Or(l.err, errNotResolved)
		return result
	}

	result.New = l.version
	if e.version == l.version {
		result.Status = StatusCurrent
	} else {
		result.Status = StatusUpdated
	}
	return result
}
