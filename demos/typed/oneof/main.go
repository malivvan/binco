// Command oneof encodes a Go interface: a type identifier followed by the concrete value.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {

	values := []Instruction{
		Copy{Start: 10, End: 20},
		Insert{Str: "hello world"},
	}
	for _, v := range values {
		bs := binco.Encode(InstructionSerializer, v)
		fmt.Printf("marshaled %T %+v → %v (%d bytes)\n", v, v, bs, len(bs))

		back, n, err := InstructionSerializer.Unmarshal(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  unmarshaled %T (read %d bytes): ", back, n)
		back.Do()
	}

	var copyIns Instruction = Copy{Start: 1, End: 2}
	bs := binco.Encode(InstructionSerializer, copyIns)
	n, err := InstructionSerializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nskipped %d of %d bytes\n", n, len(bs))

	unknown := append([]byte{99}, bs[1:]...)
	if _, _, err = InstructionSerializer.Unmarshal(unknown); err != nil {
		fmt.Printf("unknown instruction: %v\n", err)
	}
}
