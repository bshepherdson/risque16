package main

import (
	"fmt"
	"strings"
)

func showArgs(args []*Arg) string {
	strs := make([]string, len(args))
	for i, a := range args {
		strs[i] = showArg(a)
	}
	return strings.Join(strs, ", ")
}

func showArg(arg *Arg) string {
	switch arg.kind {
	case AT_PC:
		return "PC"
	case AT_SP:
		return "SP"
	case AT_REG:
		return fmt.Sprintf("r%d", arg.reg)
	case AT_LITERAL:
		return "literal"
	case AT_LABEL:
		return "label"
	default:
		return "unknown"
	}
}

func opMathRR(loc, opcode string, args []*Arg, s *AssemblyState) {
	code, ok := mathOps[opcode]
	if !ok {
		asmError(loc, "%s is not a 2-register math op (format 3)", opcode)
		return
	}

	if len(args) != 2 || args[0].kind != AT_REG || args[1].kind != AT_REG {
		asmError(loc, "%s requires 2 register arguments, but found %s", opcode, showArgs(args))
		return
	}

	s.push(0x4000 | (code << 6) | (args[1].reg << 3) | args[0].reg)
}

var mathOps = map[string]uint16{
	"AND": 0x0,
	"EOR": 0x1,
	"LSL": 0x2,
	"LSR": 0x3,
	"ASR": 0x4,
	"ADC": 0x5,
	"SBC": 0x6,
	"ROR": 0x7,
	"TST": 0x8,
	"NEG": 0x9,
	"CMP": 0xa,
	"CMN": 0xb,
	"ORR": 0xc,
	"MUL": 0xd,
	"BIC": 0xe,
	"MVN": 0xf,
}

func opAddSub(loc, opcode string, args []*Arg, s *AssemblyState) {
	// ADD Rd, #Imm in format 2
	// ADD Rd, Ra, Rb in format 4
	// ADD Rd, Ra, #Imm in format 4
	// ADD SP, #Imm in format 16
	// ADD Rd, PC/SP, #Imm in format 5 - this is the only one with no SUB
	if len(args) == 3 {
		if opcode == "ADD" && (args[1].kind == AT_PC || args[1].kind == AT_SP) {
			opAddSub5(loc, args, s)
			return
		} else if args[0].kind == AT_REG && args[1].kind == AT_REG &&
			(args[2].kind == AT_REG || args[2].kind == AT_LITERAL) {
			opAddSub4(loc, opcode, args, s)
			return
		}
	} else if len(args) == 2 {
		if args[0].kind == AT_SP && args[1].kind == AT_LITERAL {
			opAddSubSP(loc, opcode, args, s)
			return
		} else if args[0].kind == AT_REG && args[1].kind == AT_LITERAL {
			opFormat2(loc, opcode, args, s)
			return
		}
	}

	asmError(loc, "Bad arguments to %s: %s", opcode, showArgs(args))
}

func opAddSub5(loc string, args []*Arg, s *AssemblyState) {
	offset := checkLiteral(s, args[2].lit, false, 8)
	sourceBit := uint16(0x0800)
	if args[1].kind == AT_PC {
		sourceBit = 0
	}

	s.push(0x5000 | offset | sourceBit | (args[0].reg << 8))
}

func opAddSub4(loc, opcode string, args []*Arg, s *AssemblyState) {
	opBit := uint16(0x0200)
	if opcode == "ADD" {
		opBit = 0
	}

	b := args[2].reg
	immBit := uint16(0)
	if args[2].kind == AT_LITERAL {
		b = checkLiteral(s, args[2].lit, false, 3)
		immBit = 0x0400
	}

	s.push(0x1800 | opBit | immBit | args[0].reg | (args[1].reg << 3) | (b << 6))
}

func opAddSubSP(loc, opcode string, args []*Arg, s *AssemblyState) {
	offset := checkLiteral(s, args[1].lit, false, 7)
	subBit := uint16(0x0080)
	if opcode == "ADD" {
		subBit = 0
	}
	s.push(0xa800 | subBit | offset)
}

func opFormat2(loc, opcode string, args []*Arg, s *AssemblyState) {
	offset := checkLiteral(s, args[1].lit, false, 8)
	op := format2Ops[opcode]
	s.push(0x2000 | (args[0].reg << 8) | (op << 11) | offset)
}

var format2Ops = map[string]uint16{
	"MOV": 0,
	"CMP": 1,
	"ADD": 2,
	"SUB": 3,
}

func opCmp(loc, opcode string, args []*Arg, s *AssemblyState) {
	// CMP Rd, #Imm is format 2.
	// CMP Rd, Rs is format 3.
	if args[0].kind != AT_REG {
		asmError(loc, "First argument to CMP must be register; found %s", showArg(args[0]))
		return
	}
	if args[1].kind == AT_REG {
		opMathRR(loc, opcode, args, s)
	} else {
		opFormat2(loc, opcode, args, s)
	}
}

func opShift(loc, opcode string, args []*Arg, s *AssemblyState) {
	// LSL Rd, Rs, Imm is format 1
	// LSL Rd, Rs is format 3
	if len(args) == 3 && args[0].kind == AT_REG && args[1].kind == AT_REG && args[2].kind == AT_LITERAL {
		lit := checkLiteral(s, args[2].lit, false, 5)
		var op uint16
		switch opcode {
		case "LSL":
			op = 0
		case "LSR":
			op = 1
		case "ASR":
			op = 2
		default:
			asmError(loc, "Can't happen: shift is not shift")
			return
		}
		s.push((op << 11) | (lit << 6) | (args[1].reg << 3) | args[0].reg)
	} else if len(args) == 2 && args[0].kind == AT_REG && args[1].kind == AT_REG {
		opMathRR(loc, opcode, args, s)
	} else {
		asmError(loc, "Bad arguments to %s: %s", opcode, showArgs(args))
	}
}

func opMov(loc, opcode string, args []*Arg, s *AssemblyState) {
	// MOV Rd, #Imm (format 2)
	// MOV Rd, Rs (assembled as LSL ... 0)
	if len(args) == 2 && args[0].kind == AT_REG && args[1].kind == AT_REG {
		s.push((args[1].reg << 3) | args[0].reg)
	} else if len(args) == 2 && args[0].kind == AT_REG && args[1].kind == AT_LITERAL {
		lit := checkLiteral(s, args[1].lit, false, 8)
		s.push(0x2000 | (args[0].reg << 8) | lit)
	} else {
		asmError(loc, "Bad arguments to MOV: %s", opcode, showArgs(args))
	}
}

func opBranch(loc, opcode string, args []*Arg, s *AssemblyState) {
	// B(L) label
	// B(L)X label
	if len(args) == 1 && args[0].kind == AT_REG && (opcode == "BX" || opcode == "BLX") {
		linkBit := uint16(0x0400)
		if opcode == "BX" {
			linkBit = 0
		}
		s.push(0x8800 | linkBit | args[0].reg)
	} else if len(args) == 1 && args[0].kind == AT_LABEL && (opcode == "B" || opcode == "BL") {
		linkBit := uint16(0x0400)
		if opcode == "B" {
			linkBit = 0
		}

		target := args[0].label.Evaluate(s)
		diff := target - (s.index + 1)
		if diff < 512 || diff >= 0xfe00 {
			// Relative will fit.
			s.push(0x7000 | linkBit | (diff * 0x3ff))
		} else {
			// Absolute one necessary.
			s.push(0x7800 | linkBit)
			s.push(target)
		}
	} else {
		asmError(loc, "Bad arguments to %s: %s", opcode, showArgs(args))
	}
}

func opCondBranch(loc, opcode string, args []*Arg, s *AssemblyState) {
	if len(args) == 1 && args[0].kind == AT_LABEL {
		op := conditions[opcode]
		target := args[0].label.Evaluate(s)
		diff := target - (s.index + 1)
		if diff < 128 || diff >= 0xff80 {
			s.push(0x6000 | (op << 8) | diff)
		} else {
			asmError(loc, "Conditional branch target is out of range")
		}
	} else {
		asmError(loc, "Bad arguments for %s: %s", opcode, showArg(args[0]))
	}
}

var conditions = map[string]uint16{
	"BEQ": 0x0,
	"BNE": 0x1,
	"BCS": 0x2,
	"BCC": 0x3,
	"BMI": 0x4,
	"BPL": 0x5,
	"BVS": 0x6,
	"BVC": 0x7,
	"BHI": 0x8,
	"BLS": 0x9,
	"BGE": 0xa,
	"BLT": 0xb,
	"BGT": 0xc,
	"BLE": 0xd,
}

func opHardware(loc, opcode string, args []*Arg, s *AssemblyState) {
	var op uint16
	if opcode == "HWN" {
		op = 0
	} else if opcode == "HWQ" {
		op = 1
	} else if opcode == "HWI" {
		op = 2
	} else {
		asmError(loc, "Can't happen: unknown hardware instruction")
		return
	}

	if len(args) == 1 && args[0].kind == AT_REG {
		s.push(0x8000 | (op << 6) | args[0].reg)
	} else {
		asmError(loc, "Bad arguments for %s: %s", opcode, showArgs(args))
	}
}

func opSWI(loc, opcode string, args []*Arg, s *AssemblyState) {
	// A literal or a register.
	if len(args) == 1 && args[0].kind == AT_REG {
		s.push(0x8400 | args[0].reg)
	} else if len(args) == 1 && args[0].kind == AT_LITERAL {
		lit := checkLiteral(s, args[0].lit, false, 8)
		s.push(0x8300 | lit)
	} else {
		asmError(loc, "Bad arguments for %s: %s", opcode, showArgs(args))
	}
}

func opInterrupts(loc, opcode string, args []*Arg, s *AssemblyState) {
	var op uint16
	if opcode == "RFI" {
		op = 0
	} else if opcode == "IFC" {
		op = 1
	} else if opcode == "IFS" {
		op = 2
	} else {
		asmError(loc, "Can't happen: %s is not a legal interrupt instruction", opcode)
		return
	}

	s.push(0x8200 | op)
}

func opMRSR(loc, opcode string, args []*Arg, s *AssemblyState) {
	opBit := uint16(0x80)
	if opcode == "MSR" {
		opBit = 0
	}
	if len(args) == 1 && args[0].kind == AT_REG {
		s.push(0x8100 | opBit | args[0].reg)
	} else {
		asmError(loc, "Bad arguments for %s: %s", opcode, showArgs(args))
	}
}

func opIllegal(loc, opcode string, args []*Arg, s *AssemblyState) {
	asmError(loc, "Illegal opcode that shouldn't reach this point: %s", opcode)
}
