package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	bold    = "\033[1m"
	dim     = "\033[2m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	cyan    = "\033[36m"
	red     = "\033[31m"
	magenta = "\033[35m"
	reset   = "\033[0m"
)

var stdinReader = bufio.NewReader(os.Stdin)

func clearScreen() { fmt.Print("\033[H\033[2J") }

func readLine() string {
	line, _ := stdinReader.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func logInfo(msg string)  { fmt.Printf("  %sℹ%s  %s\n", cyan, reset, msg) }
func logOk(msg string)    { fmt.Printf("  %s✔%s  %s\n", green, reset, msg) }
func logWarn(msg string)  { fmt.Printf("  %s⚠%s  %s\n", yellow, reset, msg) }
func logErr(msg string)   { fmt.Fprintf(os.Stderr, "  %s✖%s  %s\n", red, reset, msg) }
func separator()          { fmt.Printf("  %s───────────────────────────────────────────%s\n", dim, reset) }

// prompt asks a free-text question with an optional default. Returns the answer.
func prompt(question, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s%s%s [%s%s%s]: ", bold, question, reset, dim, defaultVal, reset)
	} else {
		fmt.Printf("  %s%s%s: ", bold, question, reset)
	}
	ans := readLine()
	if ans == "" {
		return defaultVal
	}
	return ans
}

// yesno asks a yes/no question. Returns true for yes.
func yesno(question, defaultVal string) bool {
	hint := "Y/n"
	if defaultVal != "y" {
		hint = "y/N"
	}
	for {
		fmt.Printf("  %s%s%s [%s]: ", bold, question, reset, hint)
		ans := readLine()
		if ans == "" {
			ans = defaultVal
		}
		switch strings.ToLower(ans) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("  Please enter y or n.")
		}
	}
}

// pickOne presents a numbered list and returns the chosen item.
func pickOne(question string, opts []string) string {
	for {
		fmt.Printf("  %s%s%s\n", bold, question, reset)
		for i, o := range opts {
			fmt.Printf("    %s%d)%s %s\n", cyan, i+1, reset, o)
		}
		fmt.Printf("  Choice [1-%d]: ", len(opts))
		ans := readLine()
		n := 0
		fmt.Sscanf(ans, "%d", &n)
		if n >= 1 && n <= len(opts) {
			return opts[n-1]
		}
		fmt.Println("  Invalid choice.")
	}
}
