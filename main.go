package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func checkDeps() {
	missing := []string{}
	for _, dep := range []string{"typst", "qpdf"} {
		if _, err := exec.LookPath(dep); err != nil {
			missing = append(missing, dep)
		}
	}
	if len(missing) > 0 {
		for _, m := range missing {
			logErr("Missing dependency: " + m)
		}
		logErr("Please install them before running this tool.")
		os.Exit(1)
	}
}

func askDirectory() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Working Directory%s\n", bold, magenta, reset)
	separator()

	cwd, _ := os.Getwd()
	raw := prompt("Path to Typst project", cwd)
	raw = expandTilde(raw)

	if !filepath.IsAbs(raw) {
		raw = filepath.Join(cwd, raw)
	}

	info, err := os.Stat(raw)
	if err != nil || !info.IsDir() {
		logErr("Directory does not exist: " + raw)
		os.Exit(1)
	}

	workDir = raw
	logOk("Using: " + workDir)
}

func main() {
	fmt.Println()
	fmt.Printf("  %s%s╔══════════════════════════════════════════╗%s\n", bold, magenta, reset)
	fmt.Printf("  %s%s║   Interactive Typst Document Builder     ║%s\n", bold, magenta, reset)
	fmt.Printf("  %s%s╚══════════════════════════════════════════╝%s\n", bold, magenta, reset)

	checkDeps()
	askDirectory()
	scanProject()
	showScan()

	for {
		if !showMenu() {
			continue
		}

		switch menuAction {
		case "exit":
			fmt.Println()
			logInfo("Goodbye!")
			os.Exit(0)
		case "changedir":
			askDirectory()
			scanProject()
			showScan()
			continue
		case "languages":
			if !selectLanguages() {
				continue
			}
		case "multi":
			// no language selection needed
		}

		if menuAction == "languages" {
			askColumns()
		}
		askMedia()
		askAudience()
		askProduction()
		askCoverImage()
		askExtraTypstArgs()

		switch menuAction {
		case "languages":
			for _, lang := range selectedLangs {
				if err := buildSingle(lang); err != nil {
					logWarn(fmt.Sprintf("Build failed for '%s': %v", lang, err))
				}
			}
		case "multi":
			if err := buildMulti(); err != nil {
				logWarn("Multilingual build failed: " + err.Error())
			}
		}

		fmt.Println()
		separator()
		if yesno("Build another document?", "y") {
			clearScreen()
			scanProject()
			showScan()
			fmt.Println()
		} else {
			fmt.Println()
			logOk("All done. Goodbye!")
			os.Exit(0)
		}
	}
}
