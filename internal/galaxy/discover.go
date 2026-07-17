package galaxy

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

// requirementsRe matches ansible-galaxy requirements file names.
var requirementsRe = regexp.MustCompile(`(?i)^requirements\.ya?ml$`)

func isRequirementsFile(path string) bool {
	return requirementsRe.MatchString(filepath.Base(path))
}

// DiscoverFiles walks root recursively,
// skipping dot-directories, and returns matching files.
func DiscoverFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if isRequirementsFile(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}
	return files, nil
}
