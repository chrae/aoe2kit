package scenario

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DiffSeriesOptions struct {
	AllFields    bool
	RuntimeNames *RuntimeNames `json:"-"`
}

type DiffSeriesReport struct {
	Files []string         `json:"files"`
	Steps []DiffSeriesStep `json:"steps,omitempty"`
}

type DiffSeriesStep struct {
	Index  int        `json:"index"`
	Before string     `json:"before"`
	After  string     `json:"after"`
	Diff   DiffReport `json:"diff"`
}

// DiffSeries expands one local directory or uses the supplied paths in order.
// It performs no remote discovery and never mutates the inputs.
func DiffSeries(input []string, opts DiffSeriesOptions) (DiffSeriesReport, error) {
	paths, err := scenarioSeriesPaths(input)
	if err != nil {
		return DiffSeriesReport{}, err
	}
	report := DiffSeriesReport{Files: paths}
	for i := 1; i < len(paths); i++ {
		diff, err := DiffFilesWithOptions(paths[i-1], paths[i], DiffOptions{AllFields: opts.AllFields, RuntimeNames: opts.RuntimeNames})
		if err != nil {
			return DiffSeriesReport{}, fmt.Errorf("diff step %d (%s -> %s): %w", i, paths[i-1], paths[i], err)
		}
		report.Steps = append(report.Steps, DiffSeriesStep{Index: i, Before: paths[i-1], After: paths[i], Diff: diff})
	}
	return report, nil
}

func scenarioSeriesPaths(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("at least one scenario path is required")
	}
	if len(input) == 1 {
		info, err := os.Stat(input[0])
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			paths := make([]string, 0)
			err := filepath.WalkDir(input[0], func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".aoe2scenario") {
					paths = append(paths, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			sort.Strings(paths)
			if len(paths) < 2 {
				return nil, fmt.Errorf("directory %s contains fewer than two scenario files", input[0])
			}
			return paths, nil
		}
	}
	paths := append([]string(nil), input...)
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			return nil, fmt.Errorf("mixed directory/file series input is not supported: %s", path)
		}
	}
	return paths, nil
}
