package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// scanProject populates rootTypes, langFolders, langTypes.
func scanProject() {
	rootTypes = nil
	langFolders = nil
	langTypes = make(map[string][]string)

	entries, err := os.ReadDir(workDir)
	if err != nil {
		return
	}

	// Root .typ files
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".typ") {
			rootTypes = append(rootTypes, e.Name())
		}
	}
	sort.Strings(rootTypes)

	// Language subdirectories
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)

	for _, name := range dirs {
		if strings.HasPrefix(name, ".") || name == "sharedResources" || name == ".resources" {
			continue
		}
		subDir := filepath.Join(workDir, name)
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}
		var typs []string
		for _, se := range subEntries {
			if !se.IsDir() && strings.HasSuffix(se.Name(), ".typ") {
				typs = append(typs, se.Name())
			}
		}
		if len(typs) > 0 {
			sort.Strings(typs)
			langFolders = append(langFolders, name)
			langTypes[name] = typs
		}
	}
}

func showScan() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Project Scan Results%s\n", bold, magenta, reset)
	separator()

	if len(rootTypes) > 0 {
		fmt.Printf("  %sRoot .typ files (multilingual booklets):%s\n", bold, reset)
		for _, f := range rootTypes {
			fmt.Printf("    %s●%s %s\n", green, reset, f)
		}
	} else {
		fmt.Printf("  %sNo .typ files in project root (no multilingual booklet available)%s\n", dim, reset)
	}

	fmt.Println()

	if len(langFolders) > 0 {
		fmt.Printf("  %sLanguage folders:%s\n", bold, reset)
		for _, lang := range langFolders {
			typs := langTypes[lang]
			if len(typs) == 1 {
				fmt.Printf("    %s●%s %s/  →  %s\n", green, reset, lang, typs[0])
			} else {
				fmt.Printf("    %s●%s %s/  →  %d files\n", green, reset, lang, len(typs))
			}
		}
	} else {
		fmt.Printf("  %sNo language folders with .typ files found%s\n", dim, reset)
	}

	separator()
}
