package main

import (
	"log"
	"os"

	"gismolang.org/compiler/config"
	"gismolang.org/compiler/interpreter"
	"gismolang.org/compiler/parser"
	"gismolang.org/compiler/tokenizer"
	"gismolang.org/compiler/tokenizer/tokentype"
)

func main() {
	file := "ENVIRONMENT"
	code := os.Getenv("GISMO_CODE")
	config.OutputEnabled = os.Getenv("NO_OUT") == ""

	if code == "" {
		// Ensure a file argument is passed
		if len(os.Args) < 2 {
			log.Fatal("Usage: gismo <file-path>")
		}

		// Read file content
		file = os.Args[1]
		text, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read file '%s': %v", file, err)
		}

		code = string(text)
	}

	before, _ := os.ReadFile(config.BeforePath)
	after, _ := os.ReadFile(config.AfterPath)

	// Tokenize the file content
	tokens := []*tokenizer.Token{}
	if before != nil {
		tokens = append(tokens, tokenizer.Tokenize(string(before), file)...)
		tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline,})
	}

	tokens = append(tokens, tokenizer.Tokenize(code, file)...)

	if after != nil {
		tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline,})
		tokens = append(tokens, tokenizer.Tokenize(string(after), file)...)
	}

	// Parse the tokens into an AST
	ast := parser.Parse(tokens, file)

	config.Init()
	defer config.Deinit()
	interpreter.Interpret(ast)
}