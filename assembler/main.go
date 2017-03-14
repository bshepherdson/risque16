package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Grab the first argument and assemble it.
	file := os.Args[1]
	f, err := os.Open(file)
	p := NewParser(file, bufio.NewReader(f))
	ast, err := p.Parse()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		s := new(AssemblyState)
		s.labels = make(map[string]*LabelRef)
		s.reset()
		// Collect the labels.
		fmt.Printf("===========================\n")
		for _, l := range ast.Lines {
			fmt.Printf("line: %#v\n", l)
			labelDef, ok := l.(*LabelDef)
			if ok {
				fmt.Printf("label added: %s\n", labelDef.label)
				s.addLabel(labelDef.label)
			}
		}

		// Now actually assemble everything.
		s.dirty = true
		for s.dirty || !s.resolved {
			s.reset()
			for _, l := range ast.Lines {
				l.Assemble(s)
			}
			fmt.Printf("resolved %t dirty %t\n", s.resolved, s.dirty)
		}

		// Now output the binary, big-endian.
		// TODO: Flexible endianness.
		// TODO: Output filename.
		// TODO: Include support.
		out, _ := os.Create("out.bin")
		defer out.Close()
		for i := uint16(0); i < s.index; i++ {
			out.Write([]byte{byte(s.rom[i] >> 8), byte(s.rom[i] & 0xff)})
		}
	}
}
