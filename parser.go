package main

import (
	"errors"
	"fmt"
)

var NoMoreInstructionsError = errors.New("No more instructions to process")
var MachineHaltedError = errors.New("Machine halted")

type MissingArgumentError struct {
	opcode        VMOpcode
	ArgumentIndex MachineWord
}

func (m *MissingArgumentError) Error() string {
	return fmt.Sprintf("Argument %d is missing from %v", m.ArgumentIndex, m.opcode)
}

func ParseArguments(buffer []MachineWord) (args []Argument, err error) {
	for i := 0; i < len(buffer); i += 1 {
		arg := Argument{}
		arg.Value = buffer[i]
		if arg.Value >= 32768 && arg.Value <= 32775 {
			arg.Type = RegisterArgumentType
			arg.Value -= 32768
		} else if arg.Value < 32768 {
			arg.Type = LiteralArgumentType
		} else {
			err = fmt.Errorf("Invalid value %q", arg.Value)
			break
		}
		args = append(args, arg)
	}
	return
}

func ParseCommand(buffer []MachineWord) (cmd VMCommand, err error) {
	cmd.Opcode = VMOpcode(buffer[0])
	if cmd.Opcode > MaxOpcode {
		err = fmt.Errorf("Invalid opcode \"%d\"", cmd.Opcode)
		return
	}

	nargs := 0
	switch cmd.Opcode {
	case HaltOp:
		return
	case SetOp:
		nargs = 2
	case PushOp:
		nargs = 1
	case PopOp:
		nargs = 1
	case EqOp:
		nargs = 3
	case GtOp:
		nargs = 3
	case JmpOp:
		nargs = 1
	case JtOp:
		nargs = 2
	case JfOp:
		nargs = 2
	case AddOp:
		nargs = 3
	case MultOp:
		nargs = 3
	case ModOp:
		nargs = 3
	case AndOp:
		nargs = 3
	case OrOp:
		nargs = 3
	case NotOp:
		nargs = 2
	case RMemOp:
		nargs = 2
	case WMemOp:
		nargs = 2
	case CallOp:
		nargs = 1
	case RetOp:
		return
	case OutOp:
		nargs = 1
	case InOp:
		nargs = 1
	case NoopOp:
		return
	}

	cmd.Arguments, err = ParseArguments(buffer[1 : 1+nargs])

	return
}

type MachineWord uint16
type MachineAddress uint16

type VM struct {
	registers []MachineWord
	stack     []MachineWord
	memory    []MachineWord
	pc        MachineWord
}

func NewVM(rom []MachineWord) VM {
	//memory := make(map[MachineAddress]MachineWord, len(rom))
	//for i, val := range rom {
	//	memory[MachineAddress(i)] = val
	//}

	return VM{
		registers: make([]MachineWord, 8),
		stack:     make([]MachineWord, 0),
		memory:    rom,
		pc:        MachineWord(0),
	}
}

func resolve(arg Argument, registers []MachineWord) (MachineWord, error) {
	// Either get value from register or return literal
	switch arg.Type {
	case LiteralArgumentType:
		if arg.Value > 32767 {
			return 0, fmt.Errorf("Invalid literal %q", arg.Value)
		}
		return arg.Value, nil
	case RegisterArgumentType:
		if arg.Value > 7 {
			return 0, fmt.Errorf("Invalid register %q", arg.Value)
		}
		return MachineWord(registers[arg.Value]), nil
	default:
		return 0, fmt.Errorf("Invalid argument type %q", arg.Type)
	}
}

func resolveArguments(args []Argument, registers []MachineWord) (values []MachineWord, err error) {
	values = make([]MachineWord, len(args))
	for i, arg := range args {
		if values[i], err = resolve(arg, registers); err != nil {
			return
		}
	}
	return
}

func (v *VM) resolve(args []Argument) ([]MachineWord, error) {
	return resolveArguments(args, v.registers)
}

func (v *VM) store(addr MachineAddress, word MachineWord) {
	if cap(v.memory) <= int(addr) {
		replacement := make([]MachineWord, addr)
		for i, val := range v.memory {
			replacement[i] = val
		}
		// TODO initialize open values with invalid word
		v.memory = replacement[:int(addr)+1]
	}

	v.memory[int(addr)] = word
}

func (v *VM) fetch() (cmd VMCommand, err error) {
	// Fetch instruction at pc
	cmd, err = ParseCommand(v.memory[v.pc:])
	// fmt.Println("Fetched", cmd, "from", v.pc)
	return
}

func (v *VM) Step() (err error) {
	var cmd VMCommand
	if cmd, err = v.fetch(); err != nil {
		return err
	}

	if err := v.executeOp(cmd); err != nil {
		return err
	}

	return nil
}

func (v *VM) Execute() error {
	stepResult := v.Step()
	for ; stepResult == nil; stepResult = v.Step() {
	}
	if !errors.Is(stepResult, NoMoreInstructionsError) {
		return stepResult
	}
	return nil
}

func (v *VM) executeOp(cmd VMCommand) error {
	// Default PC
	pc := v.pc + MachineWord(len(cmd.Arguments)) + 1

	args, err := v.resolve(cmd.Arguments)
	if err != nil {
		return err
	}

	switch cmd.Opcode {
	case HaltOp:
		return MachineHaltedError
	case SetOp:
		v.registers[cmd.Arguments[0].Value] = args[1]
	case NoopOp:
		// Idle
	case PushOp:
		v.stack = append(v.stack, args[0])
	case PopOp:
		v.registers[cmd.Arguments[0].Value] = v.stack[len(v.stack)-1]
		v.stack = v.stack[:len(v.stack)-1]
	case EqOp:
		if args[1] == args[2] {
			v.registers[cmd.Arguments[0].Value] = 1
		} else {
			v.registers[cmd.Arguments[0].Value] = 0
		}
	case GtOp:
		if args[1] > args[2] {
			v.registers[cmd.Arguments[0].Value] = 1
		} else {
			v.registers[cmd.Arguments[0].Value] = 0
		}
	case JmpOp:
		pc = args[0]
	case JtOp:
		if args[0] != 0 {
			pc = args[1]
		}
	case JfOp:
		if args[0] == 0 {
			pc = args[1]
		}
	case AddOp:
		v.registers[cmd.Arguments[0].Value] = (args[1] + args[2]) % 32768
	case MultOp:
		v.registers[cmd.Arguments[0].Value] = (args[1] * args[2]) % 32768
	case ModOp:
		v.registers[cmd.Arguments[0].Value] = args[1] % args[2]
	case AndOp:
		v.registers[cmd.Arguments[0].Value] = args[1] & args[2]
	case OrOp:
		v.registers[cmd.Arguments[0].Value] = args[1] | args[2]
	case NotOp:
		v.registers[cmd.Arguments[0].Value] = (^args[1]) & 32767
	case RMemOp:
		v.registers[cmd.Arguments[0].Value] = v.memory[args[1]]
	case WMemOp:
		v.store(MachineAddress(args[0]), args[1])
	case CallOp:
		v.stack = append(v.stack, pc)
		pc = args[0]
	case RetOp:
		if len(v.stack) == 0 {
			return MachineHaltedError
		}
		pc = v.stack[len(v.stack)-1]
		v.stack = v.stack[:len(v.stack)-1]
	case OutOp:
		fmt.Printf("%c", args[0])
	case InOp:
		// read char
		var char rune
		nchars, err := fmt.Scanf("%c", &char)
		if nchars == 0 {
			return fmt.Errorf("Failed to read input from stdin")
		}
		if err != nil {
			return err
		}

		v.registers[cmd.Arguments[0].Value] = MachineWord(char)
	default:
		fmt.Println("Command", cmd, "not implemented!")
	}

	// Commit new PC
	v.pc = pc

	return nil
}
