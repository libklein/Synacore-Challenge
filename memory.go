package main

import (
	"encoding/binary"
	"errors"
	"io"
)

var OutOfMemoryError = errors.New("Out of memory")

type ContinuousMemory struct {
	memory []MachineWord
}

func (c *ContinuousMemory) store(addr MachineWord, word MachineWord) error {
	if cap(c.memory) <= int(addr) {
		replacement := make([]MachineWord, addr+1)
		copy(replacement, c.memory)
		c.memory = replacement[:int(addr)+1]
	}

	c.memory[int(addr)] = word
	return nil
}

func (c *ContinuousMemory) load(addr MachineWord) (MachineWord, error) {
	if int(addr) > len(c.memory) {
		return 0, OutOfMemoryError
	}
	return c.memory[addr], nil
}

func readMachineWord(reader io.Reader) (value MachineWord, err error) {
	err = binary.Read(reader, binary.LittleEndian, &value)
	return
}

func ContiniousMemoryFromReader(reader io.Reader) (memory ContinuousMemory, err error) {
	for addr := MachineWord(0); ; addr += 1 {
		nextWord, readErr := readMachineWord(reader)

		if readErr != nil {
			// nextWord can be ignored. binary.Read returns an error
			// only if 0 or not enough bytes were read
			if !errors.Is(readErr, io.EOF) {
				err = readErr
			}
			break
		}

		memory.store(addr, nextWord)
	}
	return
}
