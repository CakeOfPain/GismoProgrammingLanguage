package config

import (
	"fmt"
	"os"
)

var BeforePath string = "./toolchain/before.gsm"
var AfterPath string = "./toolchain/after.gsm"
var OutputPath string = "out.a"
var OutputFile *os.File = nil
var OutputEnabled bool = true

// For inbuilt function $IOTA start value
var IotaValue int = 0

func Init() {

	if OutputEnabled {
		file, err := os.Create(OutputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not open output file '%s'\n", OutputPath)
			os.Exit(1)
		}
		OutputFile = file
	}
}

func Deinit() {
	if OutputEnabled {
		OutputFile.Close()
	}
}