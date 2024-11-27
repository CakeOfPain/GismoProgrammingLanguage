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
	// Ensure a file argument is passed
	if len(os.Args) < 2 {
		log.Fatal("Usage: gismo <file-path>")
	}

	// Read file content
	file := os.Args[1]
	text, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Failed to read file '%s': %v", file, err)
	}

	before, err := os.ReadFile(config.BeforePath)
	if err != nil {
		log.Fatalf("Failed to read file '%s': %v", config.BeforePath, err)
	}

	after, err := os.ReadFile(config.AfterPath)
	if err != nil {
		log.Fatalf("Failed to read file '%s': %v", config.AfterPath, err)
	}

	// Tokenize the file content
	tokens :=  tokenizer.Tokenize(string(before), file)
	tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline,})
	tokens = append(tokens, tokenizer.Tokenize(string(text), file)...)
	tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline,})
	tokens = append(tokens, tokenizer.Tokenize(string(after), file)...)

	// Parse the tokens into an AST
	ast := parser.Parse(tokens, file)

	config.Init()
	defer config.Deinit()
	interpreter.Interpret(ast)
}