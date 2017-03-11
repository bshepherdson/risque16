package main

import (
	"fmt"
	"os"
)

type AST struct {
	Lines []Assembled
}

// Expressions evaluate to a number.
type Expression interface {
	Evaluate(s *AssemblyState) uint16
	Location() string
}

// LabelUse is a kind of expression.
// It might be a real label, or a define.
type LabelUse struct {
	label string
	loc   string
}

func (l *LabelUse) Evaluate(s *AssemblyState) uint16 {
	value, defined := s.lookup(l.label)
	if !defined {
		s.resolved = false
		return 0
	}
	return value
}

func (l *LabelUse) Location() string { return l.loc }

// Constants are a fixed-value Expression.
type Constant struct {
	value uint16
	loc   string
}

func (c *Constant) Evaluate(s *AssemblyState) uint16 { return c.value }
func (c *Constant) Location() string                 { return c.loc }

// Assembled describes something that can be assembled into the binary,
// such as an instruction, and some directives.
type Assembled interface {
	Assemble(s *AssemblyState)
}

type Include struct{ filename string }

func (i *Include) Assemble(s *AssemblyState) {
	panic("can't happen! Include survived to assembly time")
}

type Org struct{ loc Expression }

func (o *Org) Assemble(s *AssemblyState) {
	s.index = o.loc.Evaluate(s)
}

type SymbolDef struct {
	name  string
	value Expression
}

func (d *SymbolDef) Assemble(s *AssemblyState) {
	s.updateSymbol(d.name, d.value.Evaluate(s))
}

type DatBlock struct{ values []Expression }

func (b *DatBlock) Assemble(s *AssemblyState) {
	for _, v := range b.values {
		s.push(v.Evaluate(s))
	}
}

type FillBlock struct {
	length Expression
	value  Expression
}

func (b *FillBlock) Assemble(s *AssemblyState) {
	len := b.length.Evaluate(s)
	val := b.value.Evaluate(s)
	for i := uint16(0); i < len; i++ {
		s.push(val)
	}
}

type LabelDef struct{ label string }

func (l *LabelDef) Assemble(s *AssemblyState) {
	// Labels are collected in an earlier pass, but we need to note the current
	// index as its value.
	s.updateLabel(l.label, s.index)
}

type Instruction struct {
	opcode string // Should be upcased.
	args   []*Arg
	loc    string
}

func (op *Instruction) Assemble(s *AssemblyState) {
	if f, ok := instructions[op.opcode]; ok {
		f(op.loc, op.opcode, op.args, s)
	} else {
		asmError(op.loc, "Unrecognized opcode: %s", op.opcode)
	}
}

type LoadStore struct {
	storing bool
	dest    uint16 // Destination register. Required
	base    uint16 // Base register. Required, but -1 for SP/PC.
	preLit  Expression
	preReg  uint16
	postLit Expression
	postReg uint16
	baseSP  bool
}

func (op *LoadStore) Assemble(s *AssemblyState) {
	// Deal with the SP and PC special cases first.
	if op.base == 0xffff {
		// Always an 8-bit unsigned offset.
		off := uint16(0)
		if op.preLit != nil {
			off = checkLiteral(s, op.preLit, false, 8)
		}

		// Then check SP vs. PC, and LDR vs. STR.
		if !op.baseSP && !op.storing { // LDR from PC
			s.push(0xe800 | uint16(op.dest<<8) | off)
		} else if !op.baseSP && op.storing { // STR to PC is illegal.
			asmError("", "Can't do PC-relative STR")
		} else {
			storeBit := uint16(0x0800)
			if op.storing {
				storeBit = 0
			}
			s.push(0xf000 | uint16(op.dest<<8) | off | storeBit)
		}
		return
	}

	// Otherwise, standard format.
	if op.preReg != 0xffff || op.postReg != 0xffff {
		loadBit := uint16(0x0400)
		if op.storing {
			loadBit = 0
		}

		var offset uint16
		postBit := uint16(0x0200)
		if op.preReg != 0xffff {
			postBit = 0
			offset = op.preReg << 6
		} else {
			offset = op.postReg << 6
		}

		s.push(0xf000 | offset | loadBit | postBit | (op.base << 3) | op.dest)
	} else {
		loadBit := uint16(0x1000)
		if op.storing {
			loadBit = 0
		}

		var offset uint16
		postBit := uint16(0x0800)
		if op.preLit != nil {
			postBit = 0
			offset = checkLiteral(s, op.preLit, false, 5)
		} else if op.postLit != nil {
			offset = checkLiteral(s, op.postLit, false, 5)
		} else {
			offset = 0
		}

		s.push(0xc000 | loadBit | postBit | (offset << 6) | (op.base << 3) | (op.dest))
	}
}

func asmError(loc, msg string, args ...interface{}) {
	fmt.Printf("Assembly error at "+loc+" "+msg+"\n", args...)
	os.Exit(1)
}

// Exits with an error message if the literal won't fit.
func checkLiteral(s *AssemblyState, expr Expression, signed bool, width uint) uint16 {
	value := expr.Evaluate(s)
	loc := expr.Location()
	if !signed {
		if value < (1 << width) {
			return value
		}
		asmError(loc, "Unsigned literal %d (0x%x) is too big for %d-bit literal", value, value, width)
	} else {
		mask := uint16((1 << width) - 1)
		// No non-default bits outside the range.
		if (value|mask) == mask || (value|mask) == 0xffff {
			return value
		}
		asmError(loc, "Signed literal %d (0x%x) doesn't fit in %d-bit literal", value, value, width)
	}
	return 0 // Never actually happens.
}

type StackOp struct {
	regs    uint16
	storing bool
	lrpc    bool
	base    uint16
}

func (op *StackOp) Assemble(s *AssemblyState) {
	// If base is 0xffff then this is a PUSH/POP.
	if op.base == 0xffff {
		loadBit := uint16(0x0200)
		if op.storing {
			loadBit = 0
		}
		lrpcBit := uint16(0x0100)
		if !op.lrpc {
			lrpcBit = 0
		}

		s.push(0xa000 | loadBit | lrpcBit | op.regs)
	} else { // LDMIA/STMIA
		loadBit := uint16(0x0800)
		if op.storing {
			loadBit = 0
		}

		s.push(0xb000 | loadBit | op.regs | (op.base << 8))
	}
}

const (
	AT_REG int = iota
	AT_PC
	AT_SP
	AT_RLIST // Uses reg and lrpc
	AT_LABEL
	AT_LITERAL
)

type Arg struct {
	kind  int // One of the AT_* constants above.
	reg   uint16
	lrpc  bool // For reg lists.
	lit   Expression
	label Expression
}

var instructions = map[string]func(string, string, []*Arg, *AssemblyState){
	"ADC": opMathRR,
	"ADD": opAddSub,
	"AND": opMathRR,
	"ASR": opShift,
	"BIC": opMathRR,
	"CMN": opMathRR,
	"CMP": opCmp,
	"EOR": opMathRR,
	"LSL": opShift,
	"LSR": opShift,
	"MOV": opMov,
	"MUL": opMathRR,
	"MVN": opMathRR,
	"NEG": opMathRR,
	"ORR": opMathRR,
	"ROR": opMathRR,
	"SBC": opMathRR,
	"SUB": opAddSub,
	"TST": opMathRR,
	"B":   opBranch,
	"BL":  opBranch,
	"BX":  opBranch,
	"BLX": opBranch,
	"BEQ": opCondBranch,
	"BNE": opCondBranch,
	"BCS": opCondBranch,
	"BCC": opCondBranch,
	"BMI": opCondBranch,
	"BPL": opCondBranch,
	"BVS": opCondBranch,
	"BVC": opCondBranch,
	"BHI": opCondBranch,
	"BLS": opCondBranch,
	"BGE": opCondBranch,
	"BLT": opCondBranch,
	"BGT": opCondBranch,
	"BLE": opCondBranch,
	"HWN": opHardware,
	"HWI": opHardware,
	"HWQ": opHardware,
	"SWI": opSWI,
	"RFI": opInterrupts,
	"IFS": opInterrupts,
	"IFC": opInterrupts,
	"MRS": opMRSR,
	"MSR": opMRSR,
	// These are legal instructions, but we shouldn't try to evaluate them like
	// this. They have special cases.
	"LDR":   opIllegal,
	"STR":   opIllegal,
	"PUSH":  opIllegal,
	"POP":   opIllegal,
	"STMIA": opIllegal,
	"LDMIA": opIllegal,
}
