package main

import (
	"fmt"
	"strings"
)

// showMenu renders the main action menu. Returns false on invalid input.
func showMenu() bool {
	hasLang := len(langFolders) > 0
	hasRoot := len(rootTypes) > 0

	if !hasLang && !hasRoot {
		logErr("No .typ files found anywhere. Nothing to build.")
		return false
	}

	fmt.Println()
	fmt.Printf("  %s%s▸ What would you like to build?%s\n", bold, magenta, reset)
	separator()

	var actions []string

	if hasLang {
		actions = append(actions, "languages")
		fmt.Printf("    %s%d)%s Build one or more languages\n", cyan, len(actions), reset)
	}
	if hasRoot {
		actions = append(actions, "multi")
		fmt.Printf("    %s%d)%s Build multilingual booklet\n", cyan, len(actions), reset)
	}
	actions = append(actions, "changedir")
	fmt.Printf("    %s%d)%s Change working directory\n", cyan, len(actions), reset)
	actions = append(actions, "exit")
	fmt.Printf("    %s%d)%s Exit\n", cyan, len(actions), reset)

	fmt.Printf("  Choice [1-%d]: ", len(actions))
	ans := readLine()
	n := 0
	fmt.Sscanf(ans, "%d", &n)
	if n < 1 || n > len(actions) {
		logErr("Invalid choice.")
		return false
	}
	menuAction = actions[n-1]
	clearScreen()
	return true
}

// selectLanguages prompts the user to pick language folders. Returns false on error.
func selectLanguages() bool {
	fmt.Println()
	fmt.Printf("  %s%s▸ Select Languages%s\n", bold, magenta, reset)
	separator()

	fmt.Println("  Available languages:")
	for _, lang := range langFolders {
		fmt.Printf("    %s●%s %s\n", green, reset, lang)
	}
	fmt.Println()

	input := prompt("Enter language codes (comma-separated) or 'all'", "all")
	selectedLangs = nil

	if input == "all" {
		selectedLangs = append(selectedLangs, langFolders...)
	} else {
		for _, p := range strings.Split(input, ",") {
			p = strings.TrimSpace(p)
			found := false
			for _, lf := range langFolders {
				if lf == p {
					found = true
					break
				}
			}
			if found {
				selectedLangs = append(selectedLangs, p)
			} else {
				logWarn(fmt.Sprintf("Unknown language folder: '%s' — skipping", p))
			}
		}
	}

	if len(selectedLangs) == 0 {
		logErr("No valid languages selected.")
		return false
	}

	logOk("Selected: " + strings.Join(selectedLangs, " "))
	return true
}

// pickLangTyp picks a .typ file from a language folder. Returns the filename.
func pickLangTyp(lang string) string {
	typs := langTypes[lang]
	if len(typs) == 1 {
		return typs[0]
	}
	return pickOne(fmt.Sprintf("Multiple .typ files in %s/ — pick one:", lang), typs)
}

// pickRootTyp picks a root-level .typ file. Returns the filename.
func pickRootTyp() string {
	if len(rootTypes) == 1 {
		return rootTypes[0]
	}
	return pickOne("Multiple root .typ files — pick one:", rootTypes)
}
