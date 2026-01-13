package interpreter

import (
	"fmt"
	"io"
	"strings"

	"gismolang.org/compiler/parser"
	"gismolang.org/compiler/tokenizer"
	"gismolang.org/compiler/tokenizer/tokentype"
	"github.com/chzyer/readline"
)

// StartREPL starts the interactive Read-Eval-Print Loop.
func StartREPL() {
	fmt.Println("Gismo Language REPL v0.1")
	fmt.Println("Type 'exit' to quit.")

	scope := NewEmptyScope()
	scope.allowExports = true

	// 1. Configure the readline instance
	// This library handles Ctrl+L, Arrow Keys, and History automatically.
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      "gismo> ",
		HistoryFile: "/tmp/gismo_history.tmp", // Saves history to a temp file
		// You can disable history by removing the line above
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	var buffer strings.Builder

	for {
		// 2. Set the prompt dynamically
		if buffer.Len() == 0 {
			rl.SetPrompt("gismo> ")
		} else {
			rl.SetPrompt("...... ")
		}

		// 3. Read input
		// readline.Readline() blocks until Enter is pressed, BUT it handles
		// Ctrl+L (Clear) and Ctrl+C (Interrupt) immediately while you type.
		line, err := rl.Readline()
		
		if err == readline.ErrInterrupt {
			// Ctrl+C pressed
			if len(line) == 0 {
				break // Exit if line is empty
			} else {
				continue // Otherwise just clear the current line
			}
		} else if err == io.EOF {
			break
		}

		// 4. Handle Exit
		if buffer.Len() == 0 && strings.TrimSpace(line) == "exit" {
			break
		}

		// Skip empty lines if not inside a block
		if buffer.Len() == 0 && strings.TrimSpace(line) == "" {
			continue
		}

		// 5. Accumulate Input
		buffer.WriteString(line)
		buffer.WriteString("\n") // Add newline since Readline strips it

		// 6. Check Balance
		// If braces/parens are unbalanced, loop back to get more input (prompt becomes "......")
		if !isBalanced(buffer.String()) {
			continue
		}

		// 7. Execute
		executeChunk(buffer.String(), scope)

		// 8. Reset Buffer
		buffer.Reset()
	}
}

// isBalanced checks if the code has matching () [] {}.
func isBalanced(code string) bool {
	defer func() { recover() }()

	// Use a distinct filename for tokenizer error reporting if needed
	tokens := tokenizer.Tokenize(code, "REPL_CHECK")
	balance := 0

	for _, t := range tokens {
		switch t.Value {
		case "(", "{", "[":
			balance++
		case ")", "}", "]":
			balance--
		}
	}
	return balance <= 0
}

// executeChunk parses and interprets a complete block of code.
func executeChunk(code string, scope *Scope) {
	// Catch runtime errors so the REPL doesn't crash
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error: %v\n", r)
		}
	}()

	tokens := tokenizer.Tokenize(code, "REPL")
	tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline})

	ast := parser.Parse(tokens, "REPL")
	if ast == nil {
		return
	}

	sexpressions := syntaxNode2Value(ast)

	if consCell, ok := sexpressions.(*ConsCell); ok {
		length := consCell.Length()
		for i := 1; i < length; i++ {
			arg := consCell.Get(i)
			result := interpretExpression(arg, scope)
			if _, isNil := result.(*Nil); !isNil {
				fmt.Printf("=> %s\n", result.String())
			}
		}
	}
}