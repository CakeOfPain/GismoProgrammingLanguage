package config

import (
	"fmt"
	"os"
)

var BeforePath string = "./toolchain/before.gsm"
var AfterPath string = "./toolchain/after.gsm"
var OutputPath string = "out.a"
var OutputFile *os.File = nil

func Init() {
	file, err := os.Create(OutputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not open output file '%s'\n", OutputPath)
		os.Exit(1)
	}
	OutputFile = file
}

func Deinit() {
	OutputFile.Close()
}