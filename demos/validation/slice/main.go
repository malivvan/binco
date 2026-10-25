// Command slice validates the length and the elements of a slice while decoding.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	slopts "github.com/malivvan/binco/config/slice"
)

var (
	ErrTooMany = errors.New("too many elements")

	ErrEmptyElem = errors.New("empty element")
)

const maxElems = 2

func main() {
	lenVl := binco.ValidatorFn[int](func(length int) error {
		if length > maxElems {
			return fmt.Errorf("%w: %d > %d", ErrTooMany, length, maxElems)
		}
		return nil
	})
	elemVl := binco.ValidatorFn[string](func(elem string) error {
		if elem == "" {
			return ErrEmptyElem
		}
		return nil
	})

	ser := composite.NewValidSliceSer(
		composite.String,
		slopts.WithLenValidator[string](lenVl),
		slopts.WithElemValidator[string](elemVl),
	)

	values := [][]string{
		{"hello", "world"},
		{"hello", "world", "again"},
		{"hello", ""},
	}
	for _, value := range values {
		bs := binco.Encode(ser, value)
		got, _, err := ser.Unmarshal(bs)
		if err != nil {
			fmt.Printf("%-28s → rejected: %v\n", render(value), err)
			continue
		}
		fmt.Printf("%-28s → accepted: %s\n", render(value), render(got))
	}

	n, err := ser.Skip(binco.Encode(ser, values[1]))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Skip: %d bytes, no validators called\n", n)
}

func render(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
