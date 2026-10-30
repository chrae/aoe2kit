package aifile

import (
	"os"
	"path/filepath"
	"strings"
)

type LoaderPair struct {
	Marker        string   `json:"marker"`
	Scripts       []string `json:"scripts,omitempty"`
	Status        string   `json:"status"`
	Confidence    string   `json:"confidence"`
	RuntimeStatus string   `json:"runtime_status"`
}

func findLoaderPair(path, rel string) LoaderPair {
	pair := LoaderPair{Marker: rel, Status: "unresolved", Confidence: "static_analysis", RuntimeStatus: "unverified"}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		return pair
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if (ext == ".per" || ext == ".per2") && strings.EqualFold(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())), base) {
			pair.Scripts = append(pair.Scripts, filepath.ToSlash(filepath.Join(filepath.Dir(rel), entry.Name())))
		}
	}
	if len(pair.Scripts) > 0 {
		pair.Status = "present"
	}
	return pair
}
