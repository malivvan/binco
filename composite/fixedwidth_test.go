package composite_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	slopts "github.com/malivvan/binco/config/slice"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

// TestFixedWidthFastPathMatchesTheSlowPath covers P3: the bulk path for slices
// of fixed-width scalars must produce exactly the bytes of the element-by-element
// loop, and decode them back.
func TestFixedWidthFastPathMatchesTheSlowPath(t *testing.T) {
	if _, ok := interface{}(scalar.Int32).(binco.FixedWidth); !ok {
		t.Fatal("scalar.Int32 does not report a fixed width")
	}

	v := []int32{0, 1, -1, 1 << 20, -1 << 20}
	serializer := composite.NewSliceSerializer(scalar.Int32)
	bs := binco.Encode(serializer, v)

	// Reference: length prefix followed by the elements written one by one.
	want := binco.Encode(varnum.PositiveInt, len(v))
	for _, e := range v {
		want = append(want, binco.Encode(scalar.Int32, e)...)
	}
	if !bytes.Equal(bs, want) {
		t.Fatalf("bulk Marshal = %v, want %v", bs, want)
	}
	if got := serializer.Size(v); got != len(bs) {
		t.Fatalf("Size = %d, Marshal wrote %d", got, len(bs))
	}

	got, n, err := serializer.Unmarshal(bs)
	if err != nil || n != len(bs) || !reflect.DeepEqual(got, v) {
		t.Fatalf("Unmarshal = (%v, %d, %v)", got, n, err)
	}
	if n, err = serializer.Skip(bs); err != nil || n != len(bs) {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}

	if _, _, err = serializer.Unmarshal(bs[:len(bs)-1]); err == nil {
		t.Fatal("Unmarshal of a truncated fixed-width slice returned no error")
	}
}

// TestFixedWidthPathStillValidates checks that element validators are not
// bypassed by the bulk path.
func TestFixedWidthPathStillValidates(t *testing.T) {
	notNegative := binco.ValidatorFn[int32](func(v int32) error {
		if v < 0 {
			return binco.ErrOverflow
		}
		return nil
	})
	serializer := composite.NewValidSliceSerializer(scalar.Int32,
		slopts.WithElemValidator[int32](notNegative))
	bs := binco.Encode(composite.NewSliceSerializer(scalar.Int32), []int32{1, -1})
	if _, _, err := serializer.Unmarshal(bs); err == nil {
		t.Fatal("the element validator was not called")
	}
}

// TestAppendHelpers covers the P4 append helpers: they grow the caller's buffer
// and write exactly the encoded value into the tail.
func TestAppendHelpers(t *testing.T) {
	dst := []byte{0xAA}
	dst = binco.AppendEncode(varnum.Int, dst, 300)
	if want := append([]byte{0xAA}, binco.Encode(varnum.Int, 300)...); !bytes.Equal(dst, want) {
		t.Fatalf("AppendEncode = %v, want %v", dst, want)
	}
	type selfMarshalling struct{}
	_ = selfMarshalling{}
}
