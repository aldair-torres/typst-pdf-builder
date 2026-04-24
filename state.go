package main

// Global build state shared across all steps.
var (
	workDir     string
	rootTypes   []string
	langFolders []string
	langTypes   map[string][]string

	cols       string
	media      string
	audience   string
	production bool
	coverImage string

	menuAction      string
	selectedLangs   []string
	extraTypstArgs  []string
)
