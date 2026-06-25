package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// runCmd runs a command in dir. In GUI mode output is captured to buildLog.
func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if buildLogActive {
		cmd.Stdout = &buildLog
		cmd.Stderr = &buildLog
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// runQpdf assembles the final PDF using qpdf.
func runQpdf(inputPDF, outputPDF, front, back, empty string) error {
	// Build page sources: front [empty] inputPDF back
	pages := []string{front}
	if empty != "" {
		pages = append(pages, empty)
	}
	pages = append(pages, inputPDF, back)

	buildPassword := os.Getenv("TYPST_PDF_BUILD_PASSWORD")

	var args []string
	args = append(args, inputPDF, "--pages")
	args = append(args, pages...)
	args = append(args, "--")

	if production && buildPassword != "" {
		logInfo("Encryption enabled")
		args = append(args,
			"--encrypt", "", buildPassword, "256",
			"--modify=none", "--accessibility=y", "--extract=n", "--annotate=n",
			"--",
		)
	} else if production {
		logWarn("TYPST_PDF_BUILD_PASSWORD not set — building unprotected PDF")
	}

	args = append(args, outputPDF)
	return runCmd(workDir, "qpdf", args...)
}

var (
	reDocName   = regexp.MustCompile(`document-name\s*=\s*\[([^\]]*)\]`)
	reShortName = regexp.MustCompile(`full-product-name\s*=\s*"([^"]*)"`)
)

// extractVars reads document-name and short-product-name from a vars file.
func extractVars(varsFile string) (docName, shortName string, err error) {
	data, err := os.ReadFile(varsFile)
	if err != nil {
		return "", "", err
	}
	content := string(data)

	m := reDocName.FindStringSubmatch(content)
	if m != nil {
		docName = m[1]
	}
	m = reShortName.FindStringSubmatch(content)
	if m != nil {
		shortName = m[1]
	}
	if docName == "" || shortName == "" {
		return "", "", fmt.Errorf("failed to extract variables from %s", varsFile)
	}
	return docName, shortName, nil
}

// buildSingle compiles a single language document.
func buildSingle(lang string) error {
	fmt.Println()
	separator()
	fmt.Printf("  %sBuilding: %s%s\n", bold, lang, reset)
	separator()

	typ := pickLangTyp(lang)
	relPath := lang + "/" + typ
	baseName := strings.TrimSuffix(typ, ".typ")

	logInfo("File: " + relPath)

	varsFile := filepath.Join(workDir, lang, "snippets-vars", "document-info-vars.typ")
	docName, shortName, err := extractVars(varsFile)
	if err != nil {
		logErr("Variables file missing or unreadable: " + lang + "/snippets-vars/document-info-vars.typ")
		return err
	}

	if product != "" {
		shortName = product
	}
	if publication != "" {
		docName = publication
	}

	logInfo(fmt.Sprintf("Product: %s | Document: %s", shortName, docName))

	// Cover args
	coverArgs := []string{}
	if coverImage != "" {
		coverArgs = append(coverArgs, "--input", "cover-image="+coverImage)
	}
	if productLine2 != "" {
		coverArgs = append(coverArgs, "--input", "product-line2="+productLine2)
	}
	if publicationLine2 != "" {
		coverArgs = append(coverArgs, "--input", "publication-line2="+publicationLine2)
	}

	// Front cover
	logInfo("Compiling front cover...")
	frontTyp := "./" + lang + "/sharedResources/pdf-cover/" + media + "-front-cover.typ"
	frontArgs := []string{"compile", frontTyp,
		"--input", "product=" + shortName,
		"--input", "publication=" + docName,
	}
	frontArgs = append(frontArgs, coverArgs...)
	frontArgs = append(frontArgs, "--root", ".")
	if err := runCmd(workDir, "typst", frontArgs...); err != nil {
		return err
	}

	// Back cover
	logInfo("Compiling back cover...")
	backTyp := "./" + lang + "/sharedResources/pdf-cover/" + media + "-back-cover.typ"
	backArgs := []string{"compile", backTyp, "--root", "."}
	if err := runCmd(workDir, "typst", backArgs...); err != nil {
		return err
	}

	// Move covers to workDir
	coverDir := filepath.Join(workDir, lang, "sharedResources", "pdf-cover")
	if err := os.Rename(
		filepath.Join(coverDir, media+"-front-cover.pdf"),
		filepath.Join(workDir, media+"-front-cover.pdf"),
	); err != nil {
		return err
	}
	if err := os.Rename(
		filepath.Join(coverDir, media+"-back-cover.pdf"),
		filepath.Join(workDir, media+"-back-cover.pdf"),
	); err != nil {
		return err
	}

	// Main document
	logInfo("Compiling main document...")
	mainArgs := []string{"compile",
		"--input", "columns=" + cols,
		"--input", "media=" + media,
		"--input", "audience=" + audience,
		"./" + relPath,
	}
	mainArgs = append(mainArgs, extraTypstArgs...)
	if err := runCmd(workDir, "typst", mainArgs...); err != nil {
		return err
	}

	// Assemble
	date := time.Now().Format("02-01-2006")
	outFile := baseName + "_" + lang + "-" + date + ".pdf"
	inputPDF := "./" + strings.TrimSuffix(relPath, ".typ") + ".pdf"

	logInfo("Assembling PDF...")
	if err := runQpdf(
		inputPDF, outFile,
		"./"+media+"-front-cover.pdf",
		"./"+media+"-back-cover.pdf",
		"",
	); err != nil {
		return err
	}

	// Cleanup
	os.Remove(filepath.Join(workDir, inputPDF))
	os.Remove(filepath.Join(workDir, media+"-front-cover.pdf"))
	os.Remove(filepath.Join(workDir, media+"-back-cover.pdf"))

	logOk("Done: " + outFile)
	return nil
}

// buildMulti compiles the multilingual booklet.
func buildMulti() error {
	fmt.Println()
	separator()
	fmt.Printf("  %sBuilding: Multilingual booklet%s\n", bold, reset)
	separator()

	typ := pickRootTyp()
	base := strings.TrimSuffix(typ, ".typ")

	logInfo("File: " + typ)

	varsFile := filepath.Join(workDir, "en", "snippets-vars", "document-info-vars.typ")
	docName, shortName, err := extractVars(varsFile)
	if err != nil {
		logErr("Variables file missing or unreadable: en/snippets-vars/document-info-vars.typ")
		return err
	}

	if product != "" {
		shortName = product
	}
	if publication != "" {
		docName = publication
	}

	coverArgs := []string{}
	if coverImage != "" {
		coverArgs = append(coverArgs, "--input", "cover-image="+coverImage)
	}
	if productLine2 != "" {
		coverArgs = append(coverArgs, "--input", "product-line2="+productLine2)
	}
	if publicationLine2 != "" {
		coverArgs = append(coverArgs, "--input", "publication-line2="+publicationLine2)
	}

	// Front cover (en)
	logInfo("Compiling front cover (en)...")
	frontTyp := "./en/sharedResources/pdf-cover/" + media + "-front-cover.typ"
	frontArgs := []string{"compile", frontTyp,
		"--input", "product=" + shortName,
		"--input", "publication=" + docName,
	}
	frontArgs = append(frontArgs, coverArgs...)
	frontArgs = append(frontArgs, "--root", ".")
	if err := runCmd(workDir, "typst", frontArgs...); err != nil {
		return err
	}

	// Back cover (en)
	logInfo("Compiling back cover (en)...")
	backTyp := "./en/sharedResources/pdf-cover/" + media + "-back-cover.typ"
	backArgs := []string{"compile", backTyp, "--root", "."}
	if err := runCmd(workDir, "typst", backArgs...); err != nil {
		return err
	}

	coverDir := filepath.Join(workDir, "en", "sharedResources", "pdf-cover")
	if err := os.Rename(
		filepath.Join(coverDir, media+"-front-cover.pdf"),
		filepath.Join(workDir, media+"-front-cover.pdf"),
	); err != nil {
		return err
	}
	if err := os.Rename(
		filepath.Join(coverDir, media+"-back-cover.pdf"),
		filepath.Join(workDir, media+"-back-cover.pdf"),
	); err != nil {
		return err
	}

	// Main document
	logInfo("Compiling multilingual manual...")
	mainArgs := []string{"compile",
		"--input", "media=" + media,
		"--input", "audience=" + audience,
		"./" + typ,
	}
	mainArgs = append(mainArgs, extraTypstArgs...)
	if err := runCmd(workDir, "typst", mainArgs...); err != nil {
		return err
	}

	// Empty page
	empty := "./.resources/a4-empty.pdf"
	if media != "digital" {
		empty = "./.resources/a5-empty.pdf"
	}

	// Assemble
	date := time.Now().Format("02-01-2006")
	outFile := base + "_all-" + date + ".pdf"
	inputPDF := "./" + base + ".pdf"

	logInfo("Assembling PDF...")
	if err := runQpdf(
		inputPDF, outFile,
		"./"+media+"-front-cover.pdf",
		"./"+media+"-back-cover.pdf",
		empty,
	); err != nil {
		return err
	}

	// Cleanup
	os.Remove(filepath.Join(workDir, inputPDF))
	os.Remove(filepath.Join(workDir, media+"-front-cover.pdf"))
	os.Remove(filepath.Join(workDir, media+"-back-cover.pdf"))

	logOk("Done: " + outFile)
	return nil
}
