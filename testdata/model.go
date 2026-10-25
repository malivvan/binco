// Package testdata holds the fixtures the tests and the code generator share.
package testdata

import (
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/wirefmt"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
)

type Int int

type IntSlice []int

type IntHolder struct {
	V Int
}

type Str string

type SimpleStruct struct {
	Num int
	Str string
}

type IgnoredFieldStruct struct {
	Num int
	Str string
}

type Time time.Time

type Universal interface {
	Describe() string
}

type StructImpl struct {
	Value int
}

func (i StructImpl) Describe() string { return "struct" }

type DefinedImpl string

func (i DefinedImpl) Describe() string { return string(i) }

type VerFooVer1 int

type VerFooVer2 string

type VerFoo VerFooVer2

func MigrateVerFooVer1(v VerFooVer1) VerFoo {
	return VerFoo(VerFooVer2("migrated"))
}

func ValidateInt(v Int) error { return nil }

func ValidateLength(n int) error { return nil }

func ValidateElem(v int) error { return nil }

func ValidateKey(s string) error { return nil }

type FullStruct struct {
	B   bool
	By  byte
	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64
	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64
	F32 float32
	F64 float64
	S   string
	Bs  []byte
	Sl  []int
	M   map[string]int
	A   [3]int
	P   *int
	T   time.Time
	IF  Universal
}

type SelfSerialized struct {
	V int
}

const SelfSerializedDTM binco.DTM = 42

func (SelfSerialized) Describe() string { return "self" }

func (s SelfSerialized) MarshalBinco(bs []byte) (n int) {
	return varnum.PositiveInt.Marshal(s.V, bs)
}

func (s SelfSerialized) SizeBinco() (size int) {
	return varnum.PositiveInt.Size(s.V)
}

func (s SelfSerialized) MarshalTypedBinco(bs []byte) (n int) {
	n = varnum.PositiveInt.Marshal(int(SelfSerializedDTM), bs)
	return n + s.MarshalBinco(bs[n:])
}

func (s SelfSerialized) SizeTypedBinco() (size int) {
	return varnum.PositiveInt.Size(int(SelfSerializedDTM)) + s.SizeBinco()
}

type StreamSelfSerialized struct {
	V int
}

const StreamSelfSerializedDTM binco.DTM = 43

func (StreamSelfSerialized) Describe() string { return "stream-self" }

func (s StreamSelfSerialized) MarshalTypedBinco(w wirefmt.Writer) (n int, err error) {
	n, err = wvarnum.PositiveInt.Marshal(int(StreamSelfSerializedDTM), w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = wvarnum.PositiveInt.Marshal(s.V, w)
	return n + n1, err
}

func (s StreamSelfSerialized) SizeTypedBinco() (size int) {
	return wvarnum.PositiveInt.Size(int(StreamSelfSerializedDTM)) +
		wvarnum.PositiveInt.Size(s.V)
}

type Namer interface {
	Name() string
}

type NamedA struct{}

func (NamedA) Name() string { return "a" }

type NamedB int

func (NamedB) Name() string { return "b" }

type UnsupportedFieldStruct struct {
	C chan int
}
