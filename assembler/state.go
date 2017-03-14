package main

import "fmt"

type LabelRef struct {
	value   uint16
	defined bool
}

// AssemblyState tracks the state of the assembly so far.
type AssemblyState struct {
	// Fixed labels in the code, defined with :label.
	// These must be unique, and cannot be redefined.
	// These are collected early and added with addLabel(), but their values are
	// set to null initially.
	labels map[string]*LabelRef

	// Updateable defines.
	symbols map[string]*LabelRef

	// True when all labels are resolved, false otherwise.
	resolved bool
	// True when something has changed this pass (eg. a label's value).
	dirty bool

	rom   [65536]uint16
	index uint16
	used  map[uint16]bool
}

func (s *AssemblyState) lookup(key string) (uint16, bool, bool) {
	if lr, ok := s.labels[key]; ok {
		return lr.value, lr.defined, true
	}
	if lr, ok := s.symbols[key]; ok {
		return lr.value, lr.defined, true
	}
	return 0, false, false
}

func (s *AssemblyState) addLabel(l string) {
	s.labels[l] = &LabelRef{0, false}
}

func (s *AssemblyState) updateLabel(l string, loc uint16) {
	if lr, ok := s.labels[l]; ok {
		if !lr.defined || lr.value != loc {
			s.dirty = true
		}
		lr.value = loc
		lr.defined = true
	} else {
		panic(fmt.Sprintf("unknown label: '%s'", l))
	}
}

func (s *AssemblyState) updateSymbol(l string, val uint16) {
	s.symbols[l] = &LabelRef{val, true}
}

func (s *AssemblyState) reset() {
	s.symbols = make(map[string]*LabelRef)
	s.resolved = true
	s.dirty = false
	s.index = 0
	s.used = make(map[uint16]bool)
}

func (s *AssemblyState) push(x uint16) {
	if s.used[s.index] {
		panic(fmt.Sprintf("overlapping regions at $%04x", s.index))
	}
	s.used[s.index] = true
	s.rom[s.index] = x
	s.index++
}
