package composite_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	mapopts "github.com/malivvan/binco/config/map"
	slopts "github.com/malivvan/binco/config/slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

type failingSerializer struct{ err error }

func (s failingSerializer) Marshal(v int, bs []byte) (n int)      { return 0 }
func (s failingSerializer) Unmarshal(bs []byte) (int, int, error) { return 0, 0, s.err }
func (s failingSerializer) Size(v int) (size int)                 { return 1 }
func (s failingSerializer) Skip(bs []byte) (n int, err error)     { return 0, s.err }

type negativeLenSerializer struct{}

func (s negativeLenSerializer) Marshal(v int, bs []byte) (n int) { return 1 }
func (s negativeLenSerializer) Unmarshal(bs []byte) (int, int, error) {
	return -1, 1, nil
}
func (s negativeLenSerializer) Size(v int) (size int)             { return 1 }
func (s negativeLenSerializer) Skip(bs []byte) (n int, err error) { return 1, nil }

func checkPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("no panic")
		}
	}()
	fn()
}

func TestBool(t *testing.T) {
	for _, v := range []bool{true, false} {
		bs := binco.Encode(composite.Bool, v)
		want := byte(0)
		if v {
			want = 1
		}
		if len(bs) != 1 || bs[0] != want {
			t.Fatalf("Marshal(%v) = %v, want [%d]", v, bs, want)
		}
		got, n, err := composite.Bool.Unmarshal(bs)
		if err != nil || n != 1 || got != v {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}
		if n, err = composite.Bool.Skip(bs); err != nil || n != 1 {
			t.Fatalf("Skip(%v) = (%d, %v)", bs, n, err)
		}
	}
	if _, _, err := composite.Bool.Unmarshal([]byte{2}); err != binco.ErrWrongFormat {
		t.Fatalf("Unmarshal(2) returned %v, want ErrWrongFormat", err)
	}
	if _, _, err := composite.Bool.Unmarshal(nil); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal(nil) returned %v, want ErrTruncated", err)
	}
	if _, err := composite.Bool.Skip([]byte{2}); err != binco.ErrWrongFormat {
		t.Fatalf("Skip(2) returned %v, want ErrWrongFormat", err)
	}
	if _, err := composite.Bool.Skip(nil); err != binco.ErrTruncated {
		t.Fatalf("Skip(nil) returned %v, want ErrTruncated", err)
	}
}

func TestString(t *testing.T) {
	for _, v := range []string{"", "a", "hello world", "héllo"} {
		bs := binco.Encode(composite.String, v)
		if len(bs) != composite.String.Size(v) {
			t.Fatalf("Size(%q) = %d, Marshal wrote %d bytes", v, composite.String.Size(v), len(bs))
		}
		got, n, err := composite.String.Unmarshal(bs)
		if err != nil || n != len(bs) || got != v {
			t.Fatalf("Unmarshal(%v) = (%q, %d, %v)", bs, got, n, err)
		}
		if n, err = composite.String.Skip(bs); err != nil || n != len(bs) {
			t.Fatalf("Skip(%v) = (%d, %v)", bs, n, err)
		}
	}

	if bs := binco.Encode(composite.String, "hello world"); bs[0] != 11 {
		t.Fatalf("length prefix = %d, want 11", bs[0])
	}

	if _, _, err := composite.String.Unmarshal([]byte{5, 'a'}); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal of a truncated string returned %v", err)
	}
	if _, err := composite.String.Skip([]byte{5, 'a'}); err != binco.ErrTruncated {
		t.Fatalf("Skip of a truncated string returned %v", err)
	}
	if _, _, err := composite.NewStringSerializer().Unmarshal(nil); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal(nil) returned %v", err)
	}
}

func TestStringOptions(t *testing.T) {

	serializer := composite.NewStringSerializer(stropts.WithLenSerializer(scalar.Int))
	bs := binco.Encode(serializer, "abc")
	if want := append([]byte{3, 0, 0, 0, 0, 0, 0, 0}, 'a', 'b', 'c'); !reflect.DeepEqual(bs, want) {
		t.Fatalf("Marshal = %v, want %v", bs, want)
	}
	if got, err := binco.Decode(serializer, bs); err != nil || got != "abc" {
		t.Fatalf("Unmarshal = (%q, %v)", got, err)
	}

	wantErr := errors.New("too long")
	valid := composite.NewStringSerializer(stropts.WithLenValidator(binco.ValidatorFn[int](
		func(n int) error {
			if n > 2 {
				return wantErr
			}
			return nil
		})))
	if _, _, err := valid.Unmarshal(binco.Encode(valid, "abc")); !errors.Is(err, wantErr) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, wantErr)
	}
	if got, err := binco.Decode(valid, binco.Encode(valid, "ab")); err != nil || got != "ab" {
		t.Fatalf("validated Unmarshal = (%q, %v)", got, err)
	}
}

func TestByteSlice(t *testing.T) {
	for _, v := range [][]byte{nil, {}, {1}, {1, 2, 3}} {
		bs := binco.Encode(composite.ByteSlice, v)
		if len(bs) != composite.ByteSlice.Size(v) {
			t.Fatalf("Size = %d, Marshal wrote %d", composite.ByteSlice.Size(v), len(bs))
		}
		got, n, err := composite.ByteSlice.Unmarshal(bs)
		if err != nil || n != len(bs) || !reflect.DeepEqual(got, append([]byte{}, v...)) {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}
		if n, err = composite.ByteSlice.Skip(bs); err != nil || n != len(bs) {
			t.Fatalf("Skip(%v) = (%d, %v)", bs, n, err)
		}
	}
	if _, _, err := composite.ByteSlice.Unmarshal([]byte{3, 1}); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal of a truncated slice returned %v", err)
	}
	if _, err := composite.ByteSlice.Skip([]byte{3, 1}); err != binco.ErrTruncated {
		t.Fatalf("Skip of a truncated slice returned %v", err)
	}

	bs := binco.Encode(composite.ByteSlice, []byte{1, 2, 3})
	got, _, _ := composite.ByteSlice.Unmarshal(bs)
	bs[1] = 9
	if got[0] != 1 {
		t.Fatal("Unmarshal returned a byte slice that shares memory with the payload")
	}
	checkPanics(t, func() { composite.MarshalByteSlice([]byte{1, 2, 3}, varnum.PositiveInt, make([]byte, 1)) })
}

func TestPtr(t *testing.T) {
	serializer := composite.NewPtrSerializer(varnum.Int)

	var nilPtr *int
	if bs := binco.Encode(serializer, nilPtr); !reflect.DeepEqual(bs, []byte{byte(binco.PtrNil)}) {
		t.Fatalf("Marshal(nil) = %v", bs)
	}
	if got, err := binco.Decode(serializer, []byte{byte(binco.PtrNil)}); err != nil || got != nil {
		t.Fatalf("Unmarshal(nil) = (%v, %v)", got, err)
	}

	v := 5
	bs := binco.Encode(serializer, &v)
	if !reflect.DeepEqual(bs, []byte{byte(binco.PtrNotNil), 10}) {
		t.Fatalf("Marshal(&5) = %v", bs)
	}
	got, n, err := serializer.Unmarshal(bs)
	if err != nil || n != 2 || got == nil || *got != v {
		t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
	}
	if got == &v {
		t.Fatal("Unmarshal returned the original pointer")
	}
	if n, err = serializer.Skip(bs); err != nil || n != 2 {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}
	if _, _, err = serializer.Unmarshal([]byte{byte(binco.PtrMapping), 0}); err != binco.ErrWrongFormat {
		t.Fatalf("Unmarshal of a mapping flag returned %v, want ErrWrongFormat", err)
	}
	if _, _, err = serializer.Unmarshal(nil); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal(nil) returned %v", err)
	}
	if _, err = serializer.Skip(nil); err != binco.ErrTruncated {
		t.Fatalf("Skip(nil) returned %v", err)
	}
	if _, err = serializer.Skip([]byte{byte(binco.PtrMapping)}); err != binco.ErrWrongFormat {
		t.Fatalf("Skip of a mapping flag returned %v", err)
	}
}

func TestSlice(t *testing.T) {
	serializer := composite.NewSliceSerializer(varnum.Int)
	for _, v := range [][]int{nil, {}, {0}, {1, -1, 300}} {
		bs := binco.Encode(serializer, v)
		if len(bs) != serializer.Size(v) {
			t.Fatalf("Size = %d, Marshal wrote %d", serializer.Size(v), len(bs))
		}
		got, n, err := serializer.Unmarshal(bs)
		if err != nil || n != len(bs) || !reflect.DeepEqual(got, append([]int{}, v...)) {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}
		if n, err = serializer.Skip(bs); err != nil || n != len(bs) {
			t.Fatalf("Skip = (%d, %v)", n, err)
		}
	}

	wantErr := errors.New("elem")
	bad := composite.NewSliceSerializer(varnum.Int)
	if _, _, err := bad.Unmarshal([]byte{2, 0xFF}); err == nil {
		t.Fatal("Unmarshal of a malformed element returned no error")
	}
	failing := composite.NewSliceSerializer(failingSerializer{err: wantErr})
	if _, _, err := failing.Unmarshal([]byte{1, 0}); !errors.Is(err, wantErr) {
		t.Fatalf("Unmarshal returned %v, want %v", err, wantErr)
	}
	if _, err := failing.Skip([]byte{1, 0}); !errors.Is(err, wantErr) {
		t.Fatalf("Skip returned %v, want %v", err, wantErr)
	}

	rawSerializer := composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](scalar.Int))
	bs := binco.Encode(rawSerializer, []int{7})
	if len(bs) != 9 || bs[0] != 1 || bs[8] != 14 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}

	lenErr, elemErr := errors.New("len"), errors.New("elem")
	valid := composite.NewSliceSerializer(varnum.Int,
		slopts.WithLenValidator[int](binco.ValidatorFn[int](func(n int) error {
			if n > 1 {
				return lenErr
			}
			return nil
		})),
		slopts.WithElemValidator[int](binco.ValidatorFn[int](func(v int) error {
			if v < 0 {
				return elemErr
			}
			return nil
		})))
	if _, _, err := valid.Unmarshal(binco.Encode(serializer, []int{1, 2})); !errors.Is(err, lenErr) {
		t.Fatalf("length validator returned %v", err)
	}
	if _, _, err := valid.Unmarshal(binco.Encode(serializer, []int{-1})); !errors.Is(err, elemErr) {
		t.Fatalf("element validator returned %v", err)
	}
	if _, err := valid.Skip(binco.Encode(serializer, []int{1, 2})); err != nil {
		t.Fatalf("Skip returned %v", err)
	}
	checkPanics(t, func() { serializer.Marshal([]int{1, 2, 3}, make([]byte, 1)) })
}

func TestMap(t *testing.T) {
	serializer := composite.NewMapSerializer(composite.String, varnum.Int)
	for _, v := range []map[string]int{nil, {}, {"a": 1}, {"a": 1, "b": -2}} {
		bs := binco.Encode(serializer, v)
		if len(bs) != serializer.Size(v) {
			t.Fatalf("Size = %d, Marshal wrote %d for %v", serializer.Size(v), len(bs), v)
		}
		got, n, err := serializer.Unmarshal(bs)
		if err != nil || n != len(bs) {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}

		if len(v) == 0 {
			if len(got) != 0 {
				t.Fatalf("Unmarshal(%v) = %v, want an empty map", bs, got)
			}
		} else if !reflect.DeepEqual(got, v) {
			t.Fatalf("Unmarshal(%v) = %v, want %v", bs, got, v)
		}
		if n, err = serializer.Skip(bs); err != nil || n != len(bs) {
			t.Fatalf("Skip = (%d, %v)", n, err)
		}
	}

	lenErr, keyErr, valErr := errors.New("len"), errors.New("key"), errors.New("value")
	valid := composite.NewMapSerializer(composite.String, varnum.Int,
		mapopts.WithLenValidator[string, int](binco.ValidatorFn[int](func(n int) error {
			if n > 1 {
				return lenErr
			}
			return nil
		})),
		mapopts.WithKeyValidator[string, int](binco.ValidatorFn[string](func(k string) error {
			if k == "bad" {
				return keyErr
			}
			return nil
		})),
		mapopts.WithValueValidator[string, int](binco.ValidatorFn[int](func(v int) error {
			if v < 0 {
				return valErr
			}
			return nil
		})))

	nonEmpty := binco.Encode(serializer, map[string]int{"a": 1, "b": 2})
	if _, _, err := valid.Unmarshal(nonEmpty); !errors.Is(err, lenErr) {
		t.Fatalf("length validator returned %v", err)
	}
	if _, _, err := valid.Unmarshal(binco.Encode(serializer, map[string]int{"bad": 1})); !errors.Is(err, keyErr) {
		t.Fatalf("key validator returned %v", err)
	}
	if _, _, err := valid.Unmarshal(binco.Encode(serializer, map[string]int{"a": -1})); !errors.Is(err, valErr) {
		t.Fatalf("value validator returned %v", err)
	}

	failingKey := composite.NewMapSerializer(failingSerializer{err: keyErr}, varnum.Int)
	if _, _, err := failingKey.Unmarshal([]byte{1, 0}); !errors.Is(err, keyErr) {
		t.Fatalf("Unmarshal returned %v, want %v", err, keyErr)
	}
	if _, err := failingKey.Skip([]byte{1, 0}); !errors.Is(err, keyErr) {
		t.Fatalf("Skip returned %v, want %v", err, keyErr)
	}

	rawMap := composite.NewMapSerializer(varnum.Int, varnum.Int, mapopts.WithLenSerializer[int, int](scalar.Int))
	bs := binco.Encode(rawMap, map[int]int{1: 2})
	if len(bs) < 9 || bs[0] != 1 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}
	if _, err := rawMap.Skip(bs); err != nil {
		t.Fatalf("Skip = %v", err)
	}
}

func TestNegativeLength(t *testing.T) {
	if _, _, err := composite.NewStringSerializer(stropts.WithLenSerializer(negativeLenSerializer{})).Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("string returned %v, want ErrNegativeLength", err)
	}
	if _, _, err := composite.NewByteSliceSerializer(bslopts.WithLenSerializer(negativeLenSerializer{})).Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("byte slice returned %v, want ErrNegativeLength", err)
	}
	if _, _, err := composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](negativeLenSerializer{})).Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("slice returned %v, want ErrNegativeLength", err)
	}
	if _, _, err := composite.NewMapSerializer(varnum.Int, varnum.Int, mapopts.WithLenSerializer[int, int](negativeLenSerializer{})).Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("map returned %v, want ErrNegativeLength", err)
	}
}

func TestValidByteSlice(t *testing.T) {
	wantErr := errors.New("too long")
	serializer := composite.NewByteSliceSerializer(bslopts.WithLenValidator(
		binco.ValidatorFn[int](func(n int) error {
			if n > 2 {
				return wantErr
			}
			return nil
		})))

	if got, err := binco.Decode(serializer, binco.Encode(serializer, []byte{1, 2})); err != nil ||
		!reflect.DeepEqual(got, []byte{1, 2}) {
		t.Fatalf("Unmarshal = (%v, %v)", got, err)
	}

	if _, _, err := serializer.Unmarshal(binco.Encode(serializer, []byte{1, 2, 3})); !errors.Is(err, wantErr) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, wantErr)
	}

	if got, n, err := serializer.Unmarshal([]byte{0}); err != nil || n != 1 || len(got) != 0 {
		t.Fatalf("Unmarshal of an empty slice = (%v, %d, %v)", got, n, err)
	}

	if _, _, err := serializer.Unmarshal([]byte{3, 1}); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal of a truncated slice returned %v", err)
	}

	neg := composite.NewByteSliceSerializer(bslopts.WithLenSerializer(negativeLenSerializer{}))
	if _, _, err := neg.Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v, want ErrNegativeLength", err)
	}
}

func TestValidStringErrorPaths(t *testing.T) {
	valid := composite.NewStringSerializer()

	if _, _, err := valid.Unmarshal(nil); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal(nil) returned %v", err)
	}

	if _, _, err := valid.Unmarshal([]byte{5, 'a'}); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal of a truncated string returned %v", err)
	}

	if got, n, err := valid.Unmarshal([]byte{0}); err != nil || n != 1 || got != "" {
		t.Fatalf("Unmarshal of an empty string = (%q, %d, %v)", got, n, err)
	}

	neg := composite.NewStringSerializer(stropts.WithLenSerializer(negativeLenSerializer{}))
	if _, _, err := neg.Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v, want ErrNegativeLength", err)
	}
	if got, err := binco.Decode(valid, binco.Encode(valid, "abc")); err != nil || got != "abc" {
		t.Fatalf("Unmarshal = (%q, %v)", got, err)
	}
}

func TestSliceErrorPaths(t *testing.T) {
	serializer := composite.NewSliceSerializer(varnum.Int)
	valid := composite.NewSliceSerializer(varnum.Int)

	if _, _, err := serializer.Unmarshal(nil); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal(nil) returned %v", err)
	}
	if _, err := serializer.Skip(nil); err != binco.ErrTruncated {
		t.Fatalf("Skip(nil) returned %v", err)
	}
	if _, _, err := valid.Unmarshal(nil); err != binco.ErrTruncated {
		t.Fatalf("valid Unmarshal(nil) returned %v", err)
	}

	neg := composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](negativeLenSerializer{}))
	if _, _, err := neg.Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if _, err := neg.Skip([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Skip returned %v", err)
	}
	negValid := composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](negativeLenSerializer{}))
	if _, _, err := negValid.Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("valid Unmarshal returned %v", err)
	}

	wantErr := errors.New("elem")
	failing := composite.NewSliceSerializer(failingSerializer{err: wantErr})
	if _, err := failing.Skip([]byte{1, 0}); !errors.Is(err, wantErr) {
		t.Fatalf("Skip returned %v, want %v", err, wantErr)
	}
	if _, _, err := failing.Unmarshal([]byte{1, 0}); !errors.Is(err, wantErr) {
		t.Fatalf("Unmarshal returned %v, want %v", err, wantErr)
	}

	if got, n, err := valid.Unmarshal([]byte{0}); err != nil || n != 1 || len(got) != 0 {
		t.Fatalf("Unmarshal of an empty slice = (%v, %d, %v)", got, n, err)
	}
}

func TestMapErrorPaths(t *testing.T) {
	serializer := composite.NewMapSerializer(composite.String, varnum.Int)
	valid := composite.NewMapSerializer(composite.String, varnum.Int)

	if _, _, err := serializer.Unmarshal(nil); err != binco.ErrTruncated {
		t.Fatalf("Unmarshal(nil) returned %v", err)
	}
	if _, err := serializer.Skip(nil); err != binco.ErrTruncated {
		t.Fatalf("Skip(nil) returned %v", err)
	}
	if _, _, err := valid.Unmarshal(nil); err != binco.ErrTruncated {
		t.Fatalf("valid Unmarshal(nil) returned %v", err)
	}

	neg := composite.NewMapSerializer(composite.String, varnum.Int,
		mapopts.WithLenSerializer[string, int](negativeLenSerializer{}))
	if _, _, err := neg.Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v", err)
	}
	if _, err := neg.Skip([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Skip returned %v", err)
	}
	negValid := composite.NewMapSerializer(composite.String, varnum.Int,
		mapopts.WithLenSerializer[string, int](negativeLenSerializer{}))
	if _, _, err := negValid.Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("valid Unmarshal returned %v", err)
	}

	wantErr := errors.New("value")
	failingValue := composite.NewMapSerializer(composite.String, failingSerializer{err: wantErr})
	encoded := binco.Encode(composite.NewMapSerializer(composite.String, varnum.Int), map[string]int{"a": 1})
	if _, _, err := failingValue.Unmarshal(encoded); !errors.Is(err, wantErr) {
		t.Fatalf("Unmarshal returned %v, want %v", err, wantErr)
	}
	if _, err := failingValue.Skip(encoded); !errors.Is(err, wantErr) {
		t.Fatalf("Skip returned %v, want %v", err, wantErr)
	}

	if got, err := binco.Decode(valid, binco.Encode(valid, map[string]int{"a": 1})); err != nil ||
		got["a"] != 1 {
		t.Fatalf("valid Unmarshal = (%v, %v)", got, err)
	}

	failingKeyValid := composite.NewMapSerializer(failingSerializer{err: wantErr}, varnum.Int)
	if _, _, err := failingKeyValid.Unmarshal(encoded); !errors.Is(err, wantErr) {
		t.Fatalf("valid Unmarshal returned %v, want %v", err, wantErr)
	}
}

func TestPtrBaseErrors(t *testing.T) {
	wantErr := errors.New("base")
	serializer := composite.NewPtrSerializer(failingSerializer{err: wantErr})
	v := 5
	bs := binco.Encode(serializer, &v)

	if _, _, err := serializer.Unmarshal(bs); !errors.Is(err, wantErr) {
		t.Fatalf("Unmarshal returned %v, want %v", err, wantErr)
	}
	if _, err := serializer.Skip(bs); !errors.Is(err, wantErr) {
		t.Fatalf("Skip returned %v, want %v", err, wantErr)
	}
}
