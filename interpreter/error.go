package interpreter

import (
	"fmt"
	"os"
	"strings"

	"gismolang.org/compiler/tokenizer"
)

// RuntimeError prints a formatted error message with source context and exits.
func RuntimeError(token *tokenizer.Token, format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("Error: %s\n", message)

	if token != nil {
		printErrorContext(token)
	} else {
		fmt.Println("  (No source location available)")
	}

	os.Exit(1)
}

func printErrorContext(token *tokenizer.Token) {
	fmt.Printf("%s:\n", token.Source)

	content, err := os.ReadFile(token.Source)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		lineIdx := token.Line - 1

		if lineIdx >= 0 && lineIdx < len(lines) {
			codeLine := lines[lineIdx]
			// Replace tabs with spaces for alignment calculation
			cleanLine := strings.ReplaceAll(codeLine, "\t", "    ")
			
			fmt.Printf("%d: %s\n", token.Line, cleanLine)

			// Calculate padding
			prefix := fmt.Sprintf("%d: ", token.Line)
			paddingLen := len(prefix) + token.Column - 1
			padding := strings.Repeat(" ", paddingLen)

			// Determine underline length
			tokenLen := len(token.Value)
			if tokenLen == 0 {
				tokenLen = 1
			}
			underlines := strings.Repeat("^", tokenLen)

			fmt.Printf("%s%s\n", padding, underlines)
		}
	} else {
		// Fallback if file cannot be read
		fmt.Printf("  at Line %d, Column %d\n", token.Line, token.Column)
	}
}