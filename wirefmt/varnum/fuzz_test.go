//go:build fuzz

package varnum_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/varnum"
)

func FuzzUnmarshal(f *testing.F) {
	addVarintSeeds(f)
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkUnmarshal(t, "Byte", varnum.Byte, bs)
		checkUnmarshal(t, "Uint8", varnum.Uint8, bs)
		checkUnmarshal(t, "Uint16", varnum.Uint16, bs)
		checkUnmarshal(t, "Uint32", varnum.Uint32, bs)
		checkUnmarshal(t, "Uint64", varnum.Uint64, bs)
		checkUnmarshal(t, "Uint", varnum.Uint, bs)
		checkUnmarshal(t, "Int8", varnum.Int8, bs)
		checkUnmarshal(t, "Int16", varnum.Int16, bs)
		checkUnmarshal(t, "Int32", varnum.Int32, bs)
		checkUnmarshal(t, "Int64", varnum.Int64, bs)
		checkUnmarshal(t, "Int", varnum.Int, bs)
		checkUnmarshal(t, "PositiveInt8", varnum.PositiveInt8, bs)
		checkUnmarshal(t, "PositiveInt16", varnum.PositiveInt16, bs)
		checkUnmarshal(t, "PositiveInt32", varnum.PositiveInt32, bs)
		checkUnmarshal(t, "PositiveInt64", varnum.PositiveInt64, bs)
		checkUnmarshal(t, "PositiveInt", varnum.PositiveInt, bs)
	})
}

func FuzzSkip(f *testing.F) {
	addVarintSeeds(f)
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkSkip(t, "Byte", varnum.Byte, bs)
		checkSkip(t, "Uint8", varnum.Uint8, bs)
		checkSkip(t, "Uint16", varnum.Uint16, bs)
		checkSkip(t, "Uint32", varnum.Uint32, bs)
		checkSkip(t, "Uint64", varnum.Uint64, bs)
		checkSkip(t, "Uint", varnum.Uint, bs)
		checkSkip(t, "Int8", varnum.Int8, bs)
		checkSkip(t, "Int16", varnum.Int16, bs)
		checkSkip(t, "Int32", varnum.Int32, bs)
		checkSkip(t, "Int64", varnum.Int64, bs)
		checkSkip(t, "Int", varnum.Int, bs)
		checkSkip(t, "PositiveInt8", varnum.PositiveInt8, bs)
		checkSkip(t, "PositiveInt16", varnum.PositiveInt16, bs)
		checkSkip(t, "PositiveInt32", varnum.PositiveInt32, bs)
		checkSkip(t, "PositiveInt64", varnum.PositiveInt64, bs)
		checkSkip(t, "PositiveInt", varnum.PositiveInt, bs)
	})
}

func addVarintSeeds(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		{0x00},
		{0x01},
		{0x7f},
		{0x80, 0x00},
		{0x80, 0x01},
		{0xff, 0x01},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
		{0x80},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x02},
		{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01},
	} {
		f.Add(seed)
	}
}

func checkUnmarshal[T any](t *testing.T, name string, ser wirefmt.Serializer[T],
	bs []byte,
) {
	t.Helper()
	v, n, err := ser.Unmarshal(bytes.NewReader(bs))
	if n < 0 || n > len(bs) {
		t.Fatalf("%s: Unmarshal consumed %d of %d bytes", name, n, len(bs))
	}
	if err != nil {
		return
	}
	size := ser.Size(v)
	if size <= 0 || size > n {
		t.Fatalf("%s: Size(%v) = %d, Unmarshal consumed %d", name, v, size, n)
	}
	skipN, skipErr := ser.Skip(bytes.NewReader(bs))
	if skipErr != nil {
		t.Fatalf("%s: Skip failed after Unmarshal succeeded: %v", name, skipErr)
	}
	if skipN != n {
		t.Fatalf("%s: Skip consumed %d, Unmarshal consumed %d", name, skipN, n)
	}
	var buf bytes.Buffer
	written, marshalErr := ser.Marshal(v, &buf)
	if marshalErr != nil {
		t.Fatalf("%s: Marshal after Unmarshal failed: %v", name, marshalErr)
	}
	if written != size {
		t.Fatalf("%s: Marshal wrote %d, Size reported %d", name, written, size)
	}
	v2, n2, err := ser.Unmarshal(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("%s: round-trip Unmarshal failed: %v", name, err)
	}
	if n2 != written {
		t.Fatalf("%s: round-trip consumed %d, Marshal wrote %d", name, n2, written)
	}
	if !reflect.DeepEqual(v, v2) {
		t.Fatalf("%s: round-trip value %v != %v", name, v, v2)
	}
}

func checkSkip[T any](t *testing.T, name string, ser wirefmt.Serializer[T],
	bs []byte,
) {
	t.Helper()
	n, err := ser.Skip(bytes.NewReader(bs))
	if n < 0 || n > len(bs) {
		t.Fatalf("%s: Skip consumed %d of %d bytes", name, n, len(bs))
	}
	if err != nil {
		return
	}
	_, unmarshalN, unmarshalErr := ser.Unmarshal(bytes.NewReader(bs))
	if unmarshalErr == nil && unmarshalN != n {
		t.Fatalf("%s: Unmarshal consumed %d, Skip consumed %d", name, unmarshalN, n)
	}
}
