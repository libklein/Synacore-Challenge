package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <vm ROM>\n", os.Args[0])
		os.Exit(0)
	}

	file := os.Args[1]

	instructionFile, err := os.Open(file)
	if err != nil {
		fmt.Printf("Error opening %q for reading: %v\n", file, err.Error())
		os.Exit(1)
	}

	memory, err := ContiniousMemoryFromReader(instructionFile)
	if err != nil {
		fmt.Printf("Error reading ROM from %q: %v\n", file, err.Error())
		os.Exit(1)
	}

	vm := NewVM(&memory)
	vm.Run()

	return
}
