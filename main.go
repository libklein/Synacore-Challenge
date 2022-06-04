package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s binary\n", os.Args[0])
		os.Exit(0)
	}

	file := os.Args[1]

	instructionFile, err := os.Open(file)
	if err != nil {
		fmt.Printf("Error opening %q for reading: %v\n", file, err.Error())
		os.Exit(1)
	}

	// TODO Get size
	memory := make([]MachineWord, 0)
	var nextWord MachineWord
	for {
		err = binary.Read(instructionFile, binary.LittleEndian, &nextWord)
		memory = append(memory, nextWord)

		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Printf("Error parsing memory image from %q: %v\n", file, err.Error())
				os.Exit(1)
			}
			// Remove last (invalid) word
			memory = memory[:len(memory)-1]
			break
		}
	}

	vm := NewVM(memory)
	vm.Execute()

	return
}
