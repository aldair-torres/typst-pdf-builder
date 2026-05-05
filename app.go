package main

import (
	"context"
	"fmt"

	runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ScanResult is returned to the frontend after scanning a project folder.
type ScanResult struct {
	RootTypes   []string            `json:"rootTypes"`
	LangFolders []string            `json:"langFolders"`
	LangTypes   map[string][]string `json:"langTypes"`
}

// BuildParams holds all user inputs for a build operation.
type BuildParams struct {
	Mode       string            `json:"mode"`       // "languages" | "multi"
	Langs      []string          `json:"langs"`      // for "languages" mode
	LangTyps   map[string]string `json:"langTyps"`   // selected .typ file per lang
	RootTyp    string            `json:"rootTyp"`    // for "multi" mode
	Cols       string            `json:"cols"`       // "1" | "2"
	Media      string            `json:"media"`      // "digital" | "printed"
	Audience   string            `json:"audience"`
	Production bool              `json:"production"`
	CoverImage string            `json:"coverImage"`
	ExtraArgs  string            `json:"extraArgs"` // raw string, shell-split internally
}

// BuildResult is returned to the frontend after a build.
type BuildResult struct {
	Success bool     `json:"success"`
	Log     string   `json:"log"`
	Errors  []string `json:"errors"`
}

// OpenDirectory opens the native OS directory picker and returns the chosen path.
func (a *App) OpenDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Typst Project Folder",
	})
	if err != nil {
		return ""
	}
	return dir
}

// ScanProject sets the working directory, scans for .typ files, and returns the results.
func (a *App) ScanProject(dir string) ScanResult {
	workDir = dir
	scanProject()
	return ScanResult{
		RootTypes:   rootTypes,
		LangFolders: langFolders,
		LangTypes:   langTypes,
	}
}

// Build runs the compile operation with the given parameters.
func (a *App) Build(p BuildParams) BuildResult {
	cols = p.Cols
	media = p.Media
	audience = p.Audience
	production = p.Production
	coverImage = p.CoverImage
	extraTypstArgs = shellSplit(p.ExtraArgs)
	selectedLangTyps = p.LangTyps
	selectedRootTypFile = p.RootTyp

	buildLog.Reset()
	buildLogActive = true
	defer func() { buildLogActive = false }()

	var errs []string

	switch p.Mode {
	case "languages":
		for _, lang := range p.Langs {
			if err := buildSingle(lang); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", lang, err))
			}
		}
	case "multi":
		if err := buildMulti(); err != nil {
			errs = append(errs, err.Error())
		}
	default:
		errs = append(errs, "unknown build mode: "+p.Mode)
	}

	return BuildResult{
		Success: len(errs) == 0,
		Log:     buildLog.String(),
		Errors:  errs,
	}
}
