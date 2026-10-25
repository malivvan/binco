// Command string validates a string length while decoding, before the characters are read.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	stropts "github.com/malivvan/binco/config/string"
)

const maxLen = 10

var ErrTooLong = errors.New("string too long")

type lengthLimit struct{ max int }

func (l lengthLimit) Validate(n int) error {
	if n > l.max {
		return fmt.Errorf("%w: %d > %d", ErrTooLong, n, l.max)
	}
	return nil
}

func main() {

	serializer := composite.NewStringSerializer(stropts.WithLenValidator(lengthLimit{max: maxLen}))

	for _, value := range []string{"hello", strings.Repeat("a", maxLen+1)} {
		bs := binco.Encode(serializer, value)
		got, _, err := serializer.Unmarshal(bs)
		switch {
		case err != nil:
			fmt.Printf("%-14q → rejected: %v (errors.Is: %v)\n", value, err,
				errors.Is(err, ErrTooLong))
		default:
			fmt.Printf("%-14q → accepted: %q\n", value, got)
		}
	}

	big := strings.Repeat("a", maxLen+1)
	got, _, err := composite.String.Unmarshal(binco.Encode(composite.String, big))
	if err != nil {
		panic(err)
	}
	fmt.Printf("plain reader accepts %d characters\n", len(got))

	fn := binco.ValidatorFn[int](func(n int) error {
		if n%2 != 0 {
			return errors.New("odd length")
		}
		return nil
	})
	fmt.Printf("ValidatorFn(3) = %v\n", fn.Validate(3))
}
