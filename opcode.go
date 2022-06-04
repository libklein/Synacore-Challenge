package main

type VMOpcode uint16

const (
	HaltOp    VMOpcode = 0
	SetOp     VMOpcode = 1
	PushOp    VMOpcode = 2
	PopOp     VMOpcode = 3
	EqOp      VMOpcode = 4
	GtOp      VMOpcode = 5
	JmpOp     VMOpcode = 6
	JtOp      VMOpcode = 7
	JfOp      VMOpcode = 8
	AddOp     VMOpcode = 9
	MultOp    VMOpcode = 10
	ModOp     VMOpcode = 11
	AndOp     VMOpcode = 12
	OrOp      VMOpcode = 13
	NotOp     VMOpcode = 14
	RMemOp    VMOpcode = 15
	WMemOp    VMOpcode = 16
	CallOp    VMOpcode = 17
	RetOp     VMOpcode = 18
	OutOp     VMOpcode = 19
	InOp      VMOpcode = 20
	NoopOp    VMOpcode = 21
	MaxOpcode VMOpcode = 21
)
