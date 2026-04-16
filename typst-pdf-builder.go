package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
)

// ExtractionRule defines a regex pattern and how to apply it.
type ExtractionRule struct {
	Pattern   string `json:"pattern"`
	MatchMode string `json:"matchMode"` // "first" or "all"
}

// ProjectRules defines custom variable extraction rules loaded from .typst-rules.json.
// If ProductArg is present, the cos/rfs-style lookup branch is used.
type ProjectRules struct {
	ShortName  ExtractionRule  `json:"shortName"`
	DocName    ExtractionRule  `json:"docName"`
	ProductArg *ExtractionRule `json:"productArg,omitempty"`
}

// Global state
type State struct {
	WorkDir       string
	RootTypes     []string
	LangFolders   []string
	LangTypes     map[string][]string
	Rules         *ProjectRules
	Cols          string
	Media         string
	Audience      string
	Production    bool
	CoverImage    string
	MenuAction    string
	SelectedLangs []string
	Pick          string // Temporary storage for pick_one results
}

var state State
var reader = bufio.NewReader(os.Stdin)
var (
    // Pre-compile static patterns
    docNameRegex   = regexp.MustCompile(`document-name\s*=\s*\[([^\]]*)\]`)
    shortNameRegex = regexp.MustCompile(`short-product-name\s*=\s*"([^"]*)"`)
)
func main() {
	state.LangTypes = make(map[string][]string)

	showBanner()
	checkDeps()
	askDirectory()
	loadRules()
	scanProject()
	showScan()

	for {
		if !showMenu() {
			continue
		}

		switch state.MenuAction {
		case "exit":
			fmt.Println()
			logInfo("Goodbye!")
			return
		case "languages":
			if !selectLanguages() {
				continue
			}
		case "multi":
			// No language selection needed
		}

		// Build parameters
		if state.MenuAction == "languages" {
			askColumns()
		}
		askMedia()
		askAudience()
		askProduction()
		askCoverImage()

		// Build
		switch state.MenuAction {
		case "languages":
			for _, lang := range state.SelectedLangs {
				if err := buildSingle(lang); err != nil {
					logWarn(fmt.Sprintf("Build failed for '%s': %v", lang, err))
				}
			}
		case "multi":
			if err := buildMulti(); err != nil {
				logWarn(fmt.Sprintf("Multilingual build failed: %v", err))
			}
		}

		// Continue?
		fmt.Println()
		separator()
		if !yesno("Build another document?", true) {
			fmt.Println()
			logOk("All done. Goodbye!")
			return
		}

		// Re-scan in case files changed
		scanProject()
		showScan()
		fmt.Println()
	}
}

// ── Logging Utilities ─────────────────────────────────────────

func logInfo(msg string) {
	fmt.Printf("  %sℹ%s  %s\n", Cyan, Reset, msg)
}

func logOk(msg string) {
	fmt.Printf("  %s✔%s  %s\n", Green, Reset, msg)
}

func logWarn(msg string) {
	fmt.Printf("  %s⚠%s  %s\n", Yellow, Reset, msg)
}

func logErr(msg string) {
	fmt.Fprintf(os.Stderr, "  %s✖%s  %s\n", Red, Reset, msg)
}

func separator() {
	fmt.Printf("  %s───────────────────────────────────────────%s\n", Dim, Reset)
}

func showBanner() {
	fmt.Println()
	fmt.Printf("  %s%s╔══════════════════════════════════════════╗%s\n", Bold, Magenta, Reset)
	fmt.Printf("  %s%s║   Interactive Typst Document Builder     ║%s\n", Bold, Magenta, Reset)
	fmt.Printf("  %s%s╚══════════════════════════════════════════╝%s\n", Bold, Magenta, Reset)
}

// ── Dependency Check ─────────────────────────────────────────

func checkDeps() {
	missing := []string{}

	if _, err := exec.LookPath("typst"); err != nil {
		missing = append(missing, "typst")
	}
	if _, err := exec.LookPath("qpdf"); err != nil {
		missing = append(missing, "qpdf")
	}

	if len(missing) > 0 {
		logErr(fmt.Sprintf("Missing dependencies: %s", strings.Join(missing, ", ")))
		logErr("Please install them before running this script.")
		os.Exit(1)
	}
}

// ── Input Utilities ─────────────────────────────────────────

func prompt(question, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s%s%s [%s%s%s]: ", Bold, question, Reset, Dim, defaultVal, Reset)
	} else {
		fmt.Printf("  %s%s%s: ", Bold, question, Reset)
	}

	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultVal
	}
	return text
}

func yesno(question string, defaultYes bool) bool {
	hint := "Y/n"
	if !defaultYes {
		hint = "y/N"
	}

	for {
		fmt.Printf("  %s%s%s [%s]: ", Bold, question, Reset, hint)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))

		if text == "" {
			return defaultYes
		}

		switch text {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("  Please enter y or n.")
		}
	}
}

func pickOne(question string, options []string) error {
	for {
		fmt.Printf("  %s%s%s\n", Bold, question, Reset)
		for i, opt := range options {
			fmt.Printf("    %s%d)%s %s\n", Cyan, i+1, Reset, opt)
		}
		fmt.Printf("  Choice [1-%d]: ", len(options))

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		num, err := strconv.Atoi(text)
		if err != nil || num < 1 || num > len(options) {
			fmt.Println("  Invalid choice.")
			continue
		}

		state.Pick = options[num-1]
		return nil
	}
}

// ── Setup Steps ─────────────────────────────────────────

func askDirectory() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Working Directory%s\n", Bold, Magenta, Reset)
	separator()

	defaultDir, _ := os.Getwd()
	dir := prompt("Path to Typst project", defaultDir)

	// Expand ~
	if strings.HasPrefix(dir, "~") {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, dir[1:])
	}

	// Make absolute
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(defaultDir, dir)
	}
	dir = filepath.Clean(dir)

	if _, err := os.Stat(dir); err != nil {
		logErr(fmt.Sprintf("Directory does not exist: %s", dir))
		os.Exit(1)
	}

	state.WorkDir = dir
	logOk(fmt.Sprintf("Using: %s", state.WorkDir))
}

func loadRules() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Extraction Rules%s\n", Bold, Magenta, Reset)
	separator()

	rulesPath := "./.typst-rules.json"
	fmt.Println(rulesPath)
	data, err := os.ReadFile(rulesPath)
	if err != nil {
		logInfo("No .typst-rules.json found — using generic extraction")
		return
	}

	var rules ProjectRules
	if err := json.Unmarshal(data, &rules); err != nil {
		logWarn(fmt.Sprintf(".typst-rules.json parse error: %v — using generic extraction", err))
		return
	}

	if yesno("Rules file detected — use it?", true) {
		state.Rules = &rules
		logOk("Using project rules file")
	} else {
		logInfo("Using generic extraction")
	}
}


func scanProject() {
	state.RootTypes = []string{}
	state.LangFolders = []string{}
	state.LangTypes = make(map[string][]string)

	entries, err := os.ReadDir(state.WorkDir)
	if err != nil {
		logErr(fmt.Sprintf("Cannot read directory: %v", err))
		return
	}

	// Find .typ files in root
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".typ") {
			state.RootTypes = append(state.RootTypes, entry.Name())
		}
	}
	sort.Strings(state.RootTypes)

	// Find language folders
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "sharedResources" || name == ".resources" {
			continue
		}

		subEntries, err := os.ReadDir(filepath.Join(state.WorkDir, name))
		if err != nil {
			continue
		}

		typs := []string{}
		for _, sub := range subEntries {
			if sub.IsDir() {
				continue
			}
			if strings.HasSuffix(sub.Name(), ".typ") {
				typs = append(typs, sub.Name())
			}
		}

		if len(typs) > 0 {
			sort.Strings(typs)
			state.LangFolders = append(state.LangFolders, name)
			state.LangTypes[name] = typs
		}
	}
	sort.Strings(state.LangFolders)
}

func showScan() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Project Scan Results%s\n", Bold, Magenta, Reset)
	separator()

	if len(state.RootTypes) > 0 {
		fmt.Printf("  %sRoot .typ files (multilingual booklets):%s\n", Bold, Reset)
		for _, f := range state.RootTypes {
			fmt.Printf("    %s●%s %s\n", Green, Reset, f)
		}
	} else {
		fmt.Printf("  %sNo .typ files in project root (no multilingual booklet available)%s\n", Dim, Reset)
	}

	fmt.Println()

	if len(state.LangFolders) > 0 {
		fmt.Printf("  %sLanguage folders:%s\n", Bold, Reset)
		for _, lang := range state.LangFolders {
			typs := state.LangTypes[lang]
			if len(typs) == 1 {
				fmt.Printf("    %s●%s %s/  →  %s\n", Green, Reset, lang, typs[0])
			} else {
				fmt.Printf("    %s●%s %s/  →  %d files\n", Green, Reset, lang, len(typs))
			}
		}
	} else {
		fmt.Printf("  %sNo language folders with .typ files found%s\n", Dim, Reset)
	}

	separator()
}

func showMenu() bool {
	hasLang := len(state.LangFolders) > 0
	hasRoot := len(state.RootTypes) > 0

	if !hasLang && !hasRoot {
		logErr("No .typ files found anywhere. Nothing to build.")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  %s%s▸ What would you like to build?%s\n", Bold, Magenta, Reset)
	separator()

	actions := []string{}
	idx := 1

	if hasLang {
		actions = append(actions, "languages")
		fmt.Printf("    %s%d)%s Build one or more languages\n", Cyan, idx, Reset)
		idx++
	}

	if hasRoot {
		actions = append(actions, "multi")
		fmt.Printf("    %s%d)%s Build multilingual booklet\n", Cyan, idx, Reset)
		idx++
	}

	actions = append(actions, "exit")
	fmt.Printf("    %s%d)%s Exit\n", Cyan, idx, Reset)

	fmt.Printf("  Choice [1-%d]: ", idx)

	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	num, err := strconv.Atoi(text)
	if err != nil || num < 1 || num > len(actions) {
		logErr("Invalid choice.")
		return false
	}

	state.MenuAction = actions[num-1]
	return true
}

func selectLanguages() bool {
	fmt.Println()
	fmt.Printf("  %s%s▸ Select Languages%s\n", Bold, Magenta, Reset)
	separator()

	fmt.Println("  Available languages:")
	for _, lang := range state.LangFolders {
		fmt.Printf("    %s●%s %s\n", Green, Reset, lang)
	}
	fmt.Println()

	input := prompt("Enter language codes (comma-separated) or 'all'", "all")

	state.SelectedLangs = []string{}

	if input == "all" {
		state.SelectedLangs = append(state.SelectedLangs, state.LangFolders...)
	} else {
		parts := strings.Split(input, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			found := false
			for _, lang := range state.LangFolders {
				if p == lang {
					found = true
					break
				}
			}
			if found {
				state.SelectedLangs = append(state.SelectedLangs, p)
			} else {
				logWarn(fmt.Sprintf("Unknown language folder: '%s' — skipping", p))
			}
		}
	}

	if len(state.SelectedLangs) == 0 {
		logErr("No valid languages selected.")
		return false
	}

	logOk(fmt.Sprintf("Selected: %s", strings.Join(state.SelectedLangs, ", ")))
	return true
}

func askColumns() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Columns%s\n", Bold, Magenta, Reset)
	separator()

	opts := []string{"1 column", "2 columns"}
	pickOne("Column count:", opts)
	state.Cols = string(state.Pick[0]) // Extract "1" or "2"
	logOk(fmt.Sprintf("Columns: %s", state.Cols))
}

func askMedia() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Media%s\n", Bold, Magenta, Reset)
	separator()

	opts := []string{"digital", "printed"}
	pickOne("Media type:", opts)
	state.Media = state.Pick
	logOk(fmt.Sprintf("Media: %s", state.Media))
}

func askAudience() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Audience%s\n", Bold, Magenta, Reset)
	separator()

	state.Audience = prompt("Audience (Enter to skip)", "")
	if state.Audience != "" {
		logOk(fmt.Sprintf("Audience: %s", state.Audience))
	} else {
		logInfo("No audience set")
	}
}

func askProduction() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Production Mode%s\n", Bold, Magenta, Reset)
	separator()

	state.Production = yesno("Enable production mode?", false)
	logOk(fmt.Sprintf("Production: %t", state.Production))
}

func askCoverImage() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Cover Image%s\n", Bold, Magenta, Reset)
	separator()

	img := prompt("Path to cover image (Enter to skip)", "")
	if strings.HasPrefix(img, "~") {
		home, _ := os.UserHomeDir()
		img = filepath.Join(home, img[1:])
	}

	state.CoverImage = img
	if state.CoverImage != "" {
		if _, err := os.Stat(state.CoverImage); err != nil {
			logWarn(fmt.Sprintf("File not found: %s (proceeding anyway)", state.CoverImage))
		}
		logOk(fmt.Sprintf("Cover image: %s", state.CoverImage))
	} else {
		logInfo("No cover image set")
	}
}

// ── Build Logic ─────────────────────────────────────────

// applyRule extracts a value from text using an ExtractionRule.
func applyRule(rule ExtractionRule, text string) (string, error) {
	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return "", fmt.Errorf("invalid pattern %q: %w", rule.Pattern, err)
	}
	if rule.MatchMode == "first" {
		m := re.FindStringSubmatch(text)
		if len(m) > 1 {
			return m[1], nil
		}
		return "", nil
	}
	// default: "all"
	all := re.FindAllStringSubmatch(text, -1)
	return joinMatches(all), nil
}

func extractVariables(varsFile string, productArg string) (shortName, docName string, err error) {
    content, err := os.ReadFile(varsFile)
    if err != nil {
        return "", "", err
    }
    text := string(content)

    if state.Rules != nil {
        if state.Rules.ProductArg != nil {
            // productArg lookup: == "productArg" {"shortName"}
            escapedProduct := regexp.QuoteMeta(productArg)
            pattern := `==\s*"` + escapedProduct + `"\s*\{"([^"]*)"\}`
            re, compErr := regexp.Compile(pattern)
            if compErr != nil {
                return "", "", fmt.Errorf("invalid productArg: %w", compErr)
            }
            m := re.FindStringSubmatch(text)
            if len(m) > 1 {
                shortName = m[1]
            }
        } else {
            shortName, err = applyRule(state.Rules.ShortName, text)
            if err != nil {
                return "", "", err
            }
        }
        docName, err = applyRule(state.Rules.DocName, text)
        if err != nil {
            return "", "", err
        }
    } else {
        // Generic fallback
        allDocs := docNameRegex.FindAllStringSubmatch(text, -1)
        docName = joinMatches(allDocs)
        allShorts := shortNameRegex.FindAllStringSubmatch(text, -1)
        shortName = joinMatches(allShorts)
    }

    if shortName == "" || docName == "" {
        return "", "", fmt.Errorf("could not extract required variables")
    }
    return shortName, docName, nil
}

func joinMatches(matches [][]string) string {
    if len(matches) == 0 {
        return ""
    }
    var parts []string
    for _, m := range matches {
        if len(m) > 1 {
            parts = append(parts, m[1])
        }
    }
    return strings.Join(parts, "\n")
}

func runQpdf(inputPdf, outputPdf, front, back, empty string) error {
	pages := []string{front}
	if empty != "" {
		pages = append(pages, empty)
	}
	pages = append(pages, inputPdf, back)

	args := []string{inputPdf, "--pages"}
	args = append(args, pages...)
	args = append(args, "--")

	if state.Production {
		password := os.Getenv("BUILD_PASSWORD")
		if password != "" {
			logInfo("Encryption enabled")
			args = append(args,
				"--encrypt", "", password, "256",
				"--modify=none", "--accessibility=y", "--extract=n", "--annotate=n",
				"--")
		} else {
			logWarn("BUILD_PASSWORD not set — building unprotected PDF")
		}
	}

	args = append(args, outputPdf)

	cmd := exec.Command("qpdf", args...)
	cmd.Dir = state.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("qpdf failed: %v\n%s", err, output)
	}
	return nil
}

func buildSingle(lang string) error {
	separator()
	fmt.Printf("  %sBuilding: %s%s\n", Bold, lang, Reset)
	separator()

	// Pick the .typ file
	typs := state.LangTypes[lang]
	var typ string
	if len(typs) == 1 {
		typ = typs[0]
	} else {
		if err := pickOne(fmt.Sprintf("Multiple .typ files in %s/ — pick one:", lang), typs); err != nil {
			return err
		}
		typ = state.Pick
	}

	relPath := filepath.Join(lang, typ)
	baseName := strings.TrimSuffix(typ, ".typ")

	logInfo(fmt.Sprintf("File: %s", relPath))

	// Extract variables
	varsFile := filepath.Join(state.WorkDir, lang, "snippets-vars", "document-info-vars.typ")
	productArg := ""
	if state.Rules != nil && state.Rules.ProductArg != nil {
		re, _ := regexp.Compile(state.Rules.ProductArg.Pattern)
		if m := re.FindString(typ); m != "" {
			productArg = strings.ToLower(m)
		}
	}
	shortName, docName, err := extractVariables(varsFile, productArg)
	if err != nil {
		return fmt.Errorf("variables file missing or invalid: %v", err)
	}

	logInfo(fmt.Sprintf("Product: %s | Document: %s", shortName, docName))

	// Compile front cover
	fmt.Println()
	logInfo("Compiling front cover...")

	frontTyp := filepath.Join(lang, "sharedResources", "pdf-cover", state.Media+"-front-cover.typ")
	args := []string{"compile", frontTyp,
		"--input", "product=" + shortName,
		"--input", "publication=" + docName,
		"--root", "."}
	if state.CoverImage != "" {
		args = append(args, "--input", "cover-image="+state.CoverImage)
	}

	cmd := exec.Command("typst", args...)
	cmd.Dir = state.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("front cover compile failed: %v\n%s", err, output)
	}

	// Move front cover to root
	src := filepath.Join(state.WorkDir, lang, "sharedResources", "pdf-cover", state.Media+"-front-cover.pdf")
	dst := filepath.Join(state.WorkDir, state.Media+"-front-cover.pdf")
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move front cover: %v", err)
	}

	// Compile back cover
	logInfo("Compiling back cover...")
	backTyp := filepath.Join(lang, "sharedResources", "pdf-cover", state.Media+"-back-cover.typ")
	cmd = exec.Command("typst", "compile", backTyp, "--root", ".")
	cmd.Dir = state.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("back cover compile failed: %v\n%s", err, output)
	}

	// Move back cover to root
	src = filepath.Join(state.WorkDir, lang, "sharedResources", "pdf-cover", state.Media+"-back-cover.pdf")
	dst = filepath.Join(state.WorkDir, state.Media+"-back-cover.pdf")
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move back cover: %v", err)
	}

	// Compile main document
	logInfo("Compiling main document...")
	args = []string{"compile",
		"--input", "columns=" + state.Cols,
		"--input", "media=" + state.Media,
		"--input", "audience=" + state.Audience,
		relPath}
	cmd = exec.Command("typst", args...)
	cmd.Dir = state.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("main document compile failed: %v\n%s", err, output)
	}

	// Assemble
	date := time.Now().Format("02-01-2006")
	out := fmt.Sprintf("%s_%s-%s.pdf", baseName, lang, date)
	inputPdf := filepath.Join(state.WorkDir, lang, baseName+".pdf")

	logInfo("Assembling PDF...")
	if err := runQpdf(inputPdf, out,
		filepath.Join(state.WorkDir, state.Media+"-front-cover.pdf"),
		filepath.Join(state.WorkDir, state.Media+"-back-cover.pdf"),
		""); err != nil {
		return err
	}

	// Cleanup
	os.Remove(inputPdf)
	os.Remove(filepath.Join(state.WorkDir, state.Media+"-front-cover.pdf"))
	os.Remove(filepath.Join(state.WorkDir, state.Media+"-back-cover.pdf"))

	logOk(fmt.Sprintf("Done: %s", out))
	return nil
}

func buildMulti() error {
	separator()
	fmt.Printf("  %sBuilding: Multilingual booklet%s\n", Bold, Reset)
	separator()

	// Check required empty-page PDFs exist
	a4Empty := filepath.Join(state.WorkDir, ".resources", "a4-empty.pdf")
	a5Empty := filepath.Join(state.WorkDir, ".resources", "a5-empty.pdf")

	if _, err := os.Stat(a4Empty); err != nil {
		return fmt.Errorf("missing required empty-page PDF: %s", a4Empty)
	}
	if _, err := os.Stat(a5Empty); err != nil {
		return fmt.Errorf("missing required empty-page PDF: %s", a5Empty)
	}

	// Pick root typ file
	var typ string
	if len(state.RootTypes) == 1 {
		typ = state.RootTypes[0]
	} else {
		if err := pickOne("Multiple root .typ files — pick one:", state.RootTypes); err != nil {
			return err
		}
		typ = state.Pick
	}
	base := strings.TrimSuffix(typ, ".typ")

	logInfo(fmt.Sprintf("File: %s", typ))

	// Extract variables from 'en'
	varsFile := filepath.Join(state.WorkDir, "en", "snippets-vars", "document-info-vars.typ")
	productArg := ""
	if state.Rules != nil && state.Rules.ProductArg != nil {
		re, _ := regexp.Compile(state.Rules.ProductArg.Pattern)
		if m := re.FindString(typ); m != "" {
			productArg = strings.ToLower(m)
		}
	}
	shortName, docName, err := extractVariables(varsFile, productArg)
	if err != nil {
		return fmt.Errorf("variables file missing or invalid: %v", err)
	}

	// Compile covers using 'en'
	fmt.Println()
	logInfo("Compiling front cover (en)...")

	frontTyp := filepath.Join("en", "sharedResources", "pdf-cover", state.Media+"-front-cover.typ")
	args := []string{"compile", frontTyp,
		"--input", "product=" + shortName,
		"--input", "publication=" + docName,
		"--root", "."}
	if state.CoverImage != "" {
		args = append(args, "--input", "cover-image="+state.CoverImage)
	}

	cmd := exec.Command("typst", args...)
	cmd.Dir = state.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("front cover compile failed: %v\n%s", err, output)
	}

	// Move front cover
	src := filepath.Join(state.WorkDir, "en", "sharedResources", "pdf-cover", state.Media+"-front-cover.pdf")
	dst := filepath.Join(state.WorkDir, state.Media+"-front-cover.pdf")
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move front cover: %v", err)
	}

	logInfo("Compiling back cover (en)...")
	backTyp := filepath.Join("en", "sharedResources", "pdf-cover", state.Media+"-back-cover.typ")
	cmd = exec.Command("typst", "compile", backTyp, "--root", ".")
	cmd.Dir = state.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("back cover compile failed: %v\n%s", err, output)
	}

	// Move back cover
	src = filepath.Join(state.WorkDir, "en", "sharedResources", "pdf-cover", state.Media+"-back-cover.pdf")
	dst = filepath.Join(state.WorkDir, state.Media+"-back-cover.pdf")
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move back cover: %v", err)
	}

	// Compile main document
	logInfo("Compiling multilingual manual...")
	args = []string{"compile",
		"--input", "media=" + state.Media,
		"--input", "audience=" + state.Audience,
		typ}
	cmd = exec.Command("typst", args...)
	cmd.Dir = state.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("main document compile failed: %v\n%s", err, output)
	}

	// Select empty page
	var empty string
	if state.Media == "digital" {
		empty = a4Empty
	} else {
		empty = a5Empty
	}

	// Assemble
	date := time.Now().Format("02-01-2006")
	out := fmt.Sprintf("%s_all-%s.pdf", base, date)
	inputPdf := filepath.Join(state.WorkDir, base+".pdf")

	logInfo("Assembling PDF...")
	if err := runQpdf(inputPdf, out,
		filepath.Join(state.WorkDir, state.Media+"-front-cover.pdf"),
		filepath.Join(state.WorkDir, state.Media+"-back-cover.pdf"),
		empty); err != nil {
		return err
	}

	// Cleanup
	os.Remove(inputPdf)
	os.Remove(filepath.Join(state.WorkDir, state.Media+"-front-cover.pdf"))
	os.Remove(filepath.Join(state.WorkDir, state.Media+"-back-cover.pdf"))

	logOk(fmt.Sprintf("Done: %s", out))
	return nil
}
