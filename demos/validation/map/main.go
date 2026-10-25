// Command map validates the length, the keys and the values of a map while decoding.
package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	mapopts "github.com/malivvan/binco/config/map"
	"github.com/malivvan/binco/varnum"
)

var (
	ErrTooMany = errors.New("too many entries")

	ErrNegativeKey = errors.New("negative key")

	ErrEmptyValue = errors.New("empty value")
)

const maxEntries = 1

func main() {
	lenVl := binco.ValidatorFn[int](func(length int) error {
		if length > maxEntries {
			return fmt.Errorf("%w: %d > %d", ErrTooMany, length, maxEntries)
		}
		return nil
	})
	keyVl := binco.ValidatorFn[int](func(key int) error {
		if key < 0 {
			return fmt.Errorf("%w: %d", ErrNegativeKey, key)
		}
		return nil
	})
	valueVl := binco.ValidatorFn[string](func(value string) error {
		if value == "" {
			return ErrEmptyValue
		}
		return nil
	})

	serializer := composite.NewMapSerializer(
		varnum.Int,
		composite.String,
		mapopts.WithLenValidator[int, string](lenVl),
		mapopts.WithKeyValidator[int, string](keyVl),
		mapopts.WithValueValidator[int, string](valueVl),
	)

	values := []map[int]string{
		{1: "one"},
		{1: "one", 2: "two"},
		{-1: "minus one"},
		{1: ""},
	}
	for _, value := range values {
		bs := binco.Encode(serializer, value)
		got, _, err := serializer.Unmarshal(bs)
		if err != nil {
			fmt.Printf("%-22s → rejected: %v\n", render(value), err)
			continue
		}
		fmt.Printf("%-22s → accepted: %s\n", render(value), render(got))
	}
}

func render(m map[int]string) string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%d:%q", k, m[k]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
