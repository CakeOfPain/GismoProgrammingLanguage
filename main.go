package main

import (
	"flag"
	"log"
	"os"

	"gismolang.org/compiler/config"
	"gismolang.org/compiler/interpreter"
	"gismolang.org/compiler/parser"
	"gismolang.org/compiler/tokenizer"
	"gismolang.org/compiler/tokenizer/tokentype"
)

func main() {
	// 1. Define the "-o" flag (we'll also accept it after the file path)
	flag.StringVar(&config.OutputPath, "o", config.OutputPath, "Output file path")

	// Pre-scan os.Args so "-o" and "-o=..." work even after the <file-path>.
	args := os.Args[1:]
	cleaned := []string{}
	for i := 0; i < len(args); i++ {
		a := args[i]

		// "-o value"
		if a == "-o" {
			if i+1 >= len(args) {
				log.Fatal("flag needs an argument: -o")
			}
			config.OutputPath = args[i+1]
			i++
			continue
		}

		// "-o=value"
		if len(a) >= 3 && a[:3] == "-o=" {
			config.OutputPath = a[3:]
			continue
		}

		cleaned = append(cleaned, a)
	}

	// Rebuild os.Args so flag.Parse() sees only the remaining args.
	os.Args = append([]string{os.Args[0]}, cleaned...)
	flag.Parse()

	code := os.Getenv("GISMO_CODE")

	// --- REPL CHECK ---
	// If no file is provided and no code is in the environment, start REPL.
	if flag.NArg() < 1 && code == "" {
		interpreter.StartREPL()
		return
	}

	file := "ENVIRONMENT"
	config.OutputEnabled = os.Getenv("NO_OUT") == ""

	if code == "" {
		// 2. Ensure a file argument is passed
		if flag.NArg() < 1 {
			log.Fatal("Usage: gismo [-o <output-path>] <file-path>")
		}

		// 3. Read file content
		file = flag.Arg(0)
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
		tokens = append(tokens, tokenizer.Tokenize(string(before), config.BeforePath)...)
		tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline})
	}

	tokens = append(tokens, tokenizer.Tokenize(code, file)...)

	if after != nil {
		tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline})
		tokens = append(tokens, tokenizer.Tokenize(string(after), config.AfterPath)...)
	}

	// Parse the tokens into an AST
	ast := parser.Parse(tokens, file)

	config.Init()
	defer config.Deinit()

	// --- SAFE EXECUTION ---
	// We wrap the interpreter call in a function to catch panics gracefully.
	// This ensures that if a file crashes (RuntimeError), we print the error cleanly
	// instead of showing a Go stack trace.
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Print the error cleanly
				log.Printf("Execution Error: %v", r)
				// Exit with an error code so scripts know it failed
				os.Exit(1)
			}
		}()

		interpreter.Interpret(ast)
	}()
}