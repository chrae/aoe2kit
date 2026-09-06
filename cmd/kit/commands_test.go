package main

import "testing"

// The catalog is only trustworthy if it cannot silently fall behind the CLI.
// These tests are the enforcement: add a command to usage() without declaring
// its traits and the build fails here, which is the whole point of moving the
// classification out of downstream tooling.

func TestCommandCatalogCoversUsage(t *testing.T) {
	declared := map[string]bool{}
	for _, spec := range commandCatalog {
		declared[spec.Name] = true
	}
	for _, name := range usageCommandNames(usageText()) {
		if !declared[name] {
			t.Errorf("command %q appears in usage() but has no entry in commandCatalog "+
				"(add one in commands_catalog.go declaring its trait and input)", name)
		}
	}
}

func TestCommandCatalogHasNoStaleEntries(t *testing.T) {
	documented := map[string]bool{}
	for _, name := range usageCommandNames(usageText()) {
		documented[name] = true
	}
	for _, spec := range commandCatalog {
		if !documented[spec.Name] {
			t.Errorf("commandCatalog declares %q but usage() does not document it", spec.Name)
		}
	}
}

func TestCommandCatalogTraitsAreValid(t *testing.T) {
	valid := map[string]bool{TraitReadOnly: true, TraitWrites: true, TraitNetwork: true}
	for _, spec := range commandCatalog {
		if !valid[spec.Trait] {
			t.Errorf("command %q has invalid trait %q", spec.Name, spec.Trait)
		}
		if spec.Input == "" {
			t.Errorf("command %q has no declared input kind", spec.Name)
		}
	}
}
