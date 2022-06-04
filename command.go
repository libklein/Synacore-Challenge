package main

type ArgumentType int

const (
	InvalidArgumentType  ArgumentType = 0
	RegisterArgumentType ArgumentType = 1
	LiteralArgumentType  ArgumentType = 2
)

type Argument struct {
	Type  ArgumentType
	Value MachineWord
}

type VMCommand struct {
	Opcode    VMOpcode
	Arguments []Argument
}

// Execute should take looked up args and return PC? What about writing to memory? Maybe return PC, Address, Value
