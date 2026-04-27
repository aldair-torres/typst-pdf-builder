package main

import (
	"fmt"
	"os"
	"strings"
)

func askColumns() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Columns%s\n", bold, magenta, reset)
	separator()
	pick := pickOne("Column count:", []string{"1 column", "2 columns"})
	cols = string(pick[0])
	logOk("Columns: " + cols)
}

func askMedia() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Media%s\n", bold, magenta, reset)
	separator()
	media = pickOne("Media type:", []string{"digital", "printed"})
	logOk("Media: " + media)
}

func askAudience() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Audience%s\n", bold, magenta, reset)
	separator()
	audience = prompt("Audience (Enter to skip)", "")
	if audience != "" {
		logOk("Audience: " + audience)
	} else {
		logInfo("No audience set")
	}
}

func askProduction() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Production Mode%s\n", bold, magenta, reset)
	separator()
	production = yesno("Enable production mode?", "n")
	if production {
		logOk("Production: true")
	} else {
		logOk("Production: false")
	}
}

func askCoverImage() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Cover Image%s\n", bold, magenta, reset)
	separator()
	raw := prompt("Path to cover image (Enter to skip)", "")
	coverImage = expandTilde(raw)
	if coverImage != "" {
		if _, err := os.Stat(coverImage); os.IsNotExist(err) {
			logWarn("File not found: " + coverImage + " (proceeding anyway)")
		}
		logOk("Cover image: " + coverImage)
	} else {
		logInfo("No cover image set")
	}
}

func askExtraTypstArgs() {
	fmt.Println()
	fmt.Printf("  %s%s▸ Extra Typst Arguments (for main Typst file only) %s\n", bold, magenta, reset)
	separator()
	raw := prompt("Extra args passed to all typst compile calls (Enter to skip)", "")
	extraTypstArgs = shellSplit(raw)
	if len(extraTypstArgs) > 0 {
		logOk("Extra args: " + strings.Join(extraTypstArgs, " "))
	} else {
		logInfo("No extra args")
	}
}

// shellSplit splits a string into tokens respecting single and double quotes.
func shellSplit(s string) []string {
	var tokens []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inSingle:
			if c == '\'' {
				inSingle = false
			} else {
				cur.WriteByte(c)
			}
		case inDouble:
			if c == '"' {
				inDouble = false
			} else {
				cur.WriteByte(c)
			}
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == ' ' || c == '\t':
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + path[1:]
		}
	}
	return path
}
