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

    file := "ENVIRONMENT"
    code := os.Getenv("GISMO_CODE")
    config.OutputEnabled = os.Getenv("NO_OUT") == ""

    if code == "" {
        // 2. Ensure a file argument is passed
        // flag.NArg() returns the number of arguments remaining after flags are parsed.
        if flag.NArg() < 1 {
            log.Fatal("Usage: gismo [-o <output-path>] <file-path>")
        }

        // 3. Read file content
        // flag.Arg(0) is the first non-flag argument (the file path).
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
        tokens = append(tokens, tokenizer.Tokenize(string(before), file)...)
        tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline})
    }

    tokens = append(tokens, tokenizer.Tokenize(code, file)...)

    if after != nil {
        tokens = append(tokens, &tokenizer.Token{TokenType: tokentype.Newline})
        tokens = append(tokens, tokenizer.Tokenize(string(after), file)...)
    }

    // Parse the tokens into an AST
    ast := parser.Parse(tokens, file)

    config.Init()
    defer config.Deinit()
    interpreter.Interpret(ast)
}