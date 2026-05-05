package main

import "strings"

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

	menuAction     string
	selectedLangs  []string
	extraTypstArgs []string

	// GUI-only: log capture
	buildLogActive bool
	buildLog       strings.Builder

	// GUI-only: pre-selected .typ files (bypass interactive pickers)
	selectedLangTyps    map[string]string
	selectedRootTypFile string
)
