package config

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// ------------------------- Auto-Discovery (build-time friendly) -------------------------

// AutoDiscover checks for presence of files in a directory. Because Go cannot import
// packages at runtime, consumers should blank-import the package to trigger init()s.
// AutoDiscover primarily acts as verification and helpful debug output.
func AutoDiscover(dir string) {
	root, _ := os.Getwd()
	target := filepath.Join(root, dir)
	var files []string
	_ = filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" && !isTestFile(path) {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		log.Printf("[CONFIG] Warning: No config files found in %s", dir)
		return
	}

	if debugMode {
		log.Printf("[CONFIG] Auto-discovered %d config files in '%s'", len(files), dir)
	}
}

// PrintRegistrySummary shows a short summary of all registered config prefixes.
func PrintRegistrySummary() {
	prefixes := map[string]bool{}
	for key := range registry {
		if parts := strings.Split(key, "."); len(parts) > 0 {
			prefixes[parts[0]] = true
		}
	}
	log.Print("[CONFIG] ==== Registered Configs ====")
	count := 1
	for p := range prefixes {
		log.Printf("[CONFIG] %d. %s", count, p)
	}
	//log.Print("[CONFIG] ================================")
	log.Print("[CONFIG]")
}

// Helper to get type name (optional, used in debug)
func typeOf(v any) string {
	if v == nil {
		return "<nil>"
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
