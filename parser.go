package main

import (
	"errors"
	"fmt"
)

var NoMoreInstructionsError = errors.New("No more instructions to process")
var MachineHaltedError = errors.New("Machine halted")
var StackExhaustedError = errors.New("Stack exhausted!")

type MissingArgumentError struct {
	opcode        VMOpcode
	ArgumentIndex MachineWord
}

func (m *MissingArgumentError) Error() string {
	return fmt.Sprintf("Argument %d is missing from %v", m.ArgumentIndex, m.opcode)
}

type MachineWord uint16

type ReadOnlyMemory interface {
	load(addr MachineWord) (MachineWord, error)
}

type Memory interface {
	ReadOnlyMemory
	store(addr MachineWord, value MachineWord) error
}

func ParseArgument(buffer MachineWord) (argument Argument, err error) {
	argument.Value = buffer
	if argument.Value >= 32768 && argument.Value <= 32775 {
		argument.Type = RegisterArgumentType
		argument.Value -= 32768
	} else if argument.Value < 32768 {
		argument.Type = LiteralArgumentType
	} else {
		err = fmt.Errorf("Invalid value %q", argument.Value)
	}
	return
}

func ParseArguments(memory ReadOnlyMemory, addr MachineWord, nargs MachineWord) (args []Argument, err error) {
	args = make([]Argument, nargs)
	for i := MachineWord(0); i < nargs; i += 1 {
		value, readErr := memory.load(addr + i)
		if readErr != nil {
			err = readErr
			break
		}

		if args[i], err = ParseArgument(value); err != nil {
			args = args[:i]
			break
		}
	}
	return
}

func ParseCommand(memory ReadOnlyMemory, addr MachineWord) (cmd VMCommand, err error) {
	rawOpcode, err := memory.load(addr)
	if err != nil {
		return cmd, nil
	}

	if rawOpcode > MachineWord(MaxOpcode) {
		return cmd, fmt.Errorf("Invalid opcode \"%d\"", rawOpcode)
	}

	cmd.Opcode = VMOpcode(rawOpcode)

	nargs := MachineWord(0)
	switch cmd.Opcode {
	case EqOp, GtOp, AddOp, MultOp, ModOp, AndOp, OrOp:
		nargs = 3
	case SetOp, JtOp, JfOp, NotOp, RMemOp, WMemOp:
		nargs = 2
	case PushOp, PopOp, JmpOp, CallOp, OutOp, InOp:
		nargs = 1
	case RetOp, NoopOp, HaltOp:
		return
	}

	cmd.Arguments, err = ParseArguments(memory, addr+1, nargs)

	return
}

func resolveArgument(arg Argument, registers []MachineWord) (MachineWord, error) {
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
		if values[i], err = resolveArgument(arg, registers); err != nil {
			return
		}
	}
	return
}

func push[T any](value T, stack *[]T) {
	*stack = append(*stack, value)
}

func pop[T any](stack *[]T) (value T, err error) {
	if len(*stack) == 0 {
		err = StackExhaustedError
	} else {
		value = (*stack)[len(*stack)-1]
		*stack = (*stack)[:len(*stack)-1]
	}
	return
}

func readChar() (char rune, err error) {
	nchars, err := fmt.Scanf("%c", &char)
	if nchars != 1 {
		err = fmt.Errorf("Failed to read input from stdin")
	}
	return
}

type VM struct {
	registers []MachineWord
	stack     []MachineWord
	memory    Memory
	pc        MachineWord
}

func NewVM(memory Memory) VM {
	return VM{
		registers: make([]MachineWord, 8),
		stack:     make([]MachineWord, 0),
		memory:    memory,
		pc:        MachineWord(0),
	}
}

func (v *VM) resolve(args []Argument) ([]MachineWord, error) {
	return resolveArguments(args, v.registers)
}

func (v *VM) fetch() (cmd VMCommand, err error) {
	// Fetch instruction at pc
	cmd, err = ParseCommand(v.memory, v.pc)
	return
}

func (v *VM) Step() (err error) {
	var cmd VMCommand
	if cmd, err = v.fetch(); err != nil {
		return err
	}

	if pc, err := v.executeOp(cmd); err != nil {
		return err
	} else {
		v.pc = pc
	}

	return nil
}

func (v *VM) Run() error {
	stepResult := v.Step()
	for ; stepResult == nil; stepResult = v.Step() {
	}
	if !errors.Is(stepResult, NoMoreInstructionsError) {
		return stepResult
	}
	return nil
}

func (v *VM) executeStorageOperation(cmd VMCommand) error {
	dest := cmd.Arguments[0].Value
	args, err := v.resolve(cmd.Arguments[1:])

	if err != nil {
		return err
	}

	switch cmd.Opcode {
	case EqOp:
		if args[0] == args[1] {
			v.registers[dest] = 1
		} else {
			v.registers[dest] = 0
		}
	case GtOp:
		if args[0] > args[1] {
			v.registers[dest] = 1
		} else {
			v.registers[dest] = 0
		}
	case SetOp:
		v.registers[dest] = args[0]
	case PopOp:
		v.registers[dest], err = pop(&v.stack)
		if err != nil {
			return err
		}
	case AddOp:
		v.registers[dest] = (args[0] + args[1]) % 32768
	case MultOp:
		v.registers[dest] = (args[0] * args[1]) % 32768
	case ModOp:
		v.registers[dest] = args[0] % args[1]
	case AndOp:
		v.registers[dest] = args[0] & args[1]
	case OrOp:
		v.registers[dest] = args[0] | args[1]
	case NotOp:
		v.registers[dest] = (^args[0]) & 32767
	case RMemOp:
		v.registers[dest], err = v.memory.load(args[0])
		if err != nil {
			return err
		}
	case InOp:
		char, err := readChar()
		if err != nil {
			return err
		}
		v.registers[dest] = MachineWord(char)
	default:
		return fmt.Errorf("Unknown storage command %q", cmd)
	}

	return nil
}

func (v *VM) executeControlFlow(cmd VMCommand) (pc MachineWord, err error) {
	pc = v.pc + MachineWord(len(cmd.Arguments)) + 1

	args, err := v.resolve(cmd.Arguments)
	if err != nil {
		return
	}

	switch cmd.Opcode {
	case HaltOp:
		err = MachineHaltedError
	case RetOp:
		pc, err = pop(&v.stack)
		if errors.Is(err, StackExhaustedError) {
			err = MachineHaltedError
		}
	case NoopOp:
	case PushOp:
		push(args[0], &v.stack)
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
	case WMemOp:
		err = v.memory.store(args[0], args[1])
	case CallOp:
		push(pc, &v.stack)
		pc = args[0]
	case OutOp:
		fmt.Printf("%c", args[0])
	default:
		pc = 0
		err = fmt.Errorf("Unknown control flow operation: %q", cmd)
	}
	return
}

func (v *VM) executeOp(cmd VMCommand) (MachineWord, error) {
	switch cmd.Opcode {
	case HaltOp, RetOp, NoopOp, PushOp, JmpOp, JtOp, JfOp, WMemOp, CallOp, OutOp:
		return v.executeControlFlow(cmd)
	case EqOp, GtOp, SetOp, PopOp, AddOp, MultOp, ModOp, AndOp, OrOp, NotOp, RMemOp, InOp:
		return v.pc + MachineWord(len(cmd.Arguments)) + 1, v.executeStorageOperation(cmd)
	default:
		return 0, fmt.Errorf("Unknown operation: %q", cmd)
	}
}
