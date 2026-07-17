// Package galaxy updates Ansible collection and role versions to the latest release.
package galaxy

import (
	"context"
	"fmt"
)

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

	collNames, roleNames := namesToResolve(parsed)
	resolved := map[kind]resolution{
		kindCollection: c.resolveCollections(ctx, collNames),
		kindRole:       c.resolveRoles(ctx, roleNames),
	}

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

type resolution struct {
	versions map[string]string
	errs     map[string]error
}

// namesToResolve collects the galaxy names to look up, split by kind.
func namesToResolve(parsed []fileData) (collections, roles []string) {
	for _, fd := range parsed {
		for _, e := range fd.entries {
			if e.skip {
				continue
			}
			if e.kind == kindRole {
				roles = append(roles, e.name)
			} else {
				collections = append(collections, e.name)
			}
		}
	}
	return collections, roles
}

func classifyFile(fd *fileData, resolved map[kind]resolution) (results []Result, changed bool) {
	for i := range fd.entries {
		e := &fd.entries[i]
		result := classifyEntry(fd.path, e, resolved[e.kind])
		if result.Status == StatusUpdated {
			e.setVersion(result.New)
			changed = true
		}
		results = append(results, result)
	}
	return results, changed
}

func classifyEntry(path string, e *entry, res resolution) Result {
	result := Result{File: path, Name: e.name, Old: e.version}
	if e.skip {
		result.Status = StatusSkipped
		return result
	}

	latest, ok := res.versions[e.name]
	if !ok {
		result.Status = StatusError
		result.Err = res.errs[e.name]
		if result.Err == nil {
			result.Err = fmt.Errorf("no result for %s", e.name)
		}
		return result
	}

	result.New = latest
	if e.version == latest {
		result.Status = StatusCurrent
	} else {
		result.Status = StatusUpdated
	}
	return result
}
