package binco_test

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

type marshaller struct{ v int }

func (m marshaller) MarshalBinco(bs []byte) (n int) { return varnum.Int.Marshal(m.v, bs) }
func (m marshaller) SizeBinco() (size int)          { return varnum.Int.Size(m.v) }

type marshallerTyped struct{ dtm binco.DTM }

func (m marshallerTyped) MarshalTypedBinco(bs []byte) (n int) {
	bs[0] = byte(m.dtm)
	return 1
}
func (m marshallerTyped) SizeTypedBinco() (size int) { return 1 }

func TestEncodeDecode(t *testing.T) {
	v := 300
	bs := binco.Encode(varnum.Int, v)
	if got, want := len(bs), varnum.Int.Size(v); got != want {
		t.Fatalf("Encode allocated %d bytes, want %d", got, want)
	}
	got, err := binco.Decode(varnum.Int, bs)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got != v {
		t.Fatalf("Decode returned %d, want %d", got, v)
	}
	if _, err = binco.Decode(varnum.Int, nil); !errors.Is(err, binco.ErrTooSmallByteSlice) {
		t.Fatalf("Decode of empty input returned %v, want ErrTooSmallByteSlice", err)
	}
}

func TestMarshal(t *testing.T) {
	bs := binco.Marshal(marshaller{v: 5})
	want := binco.Encode(varnum.Int, 5)
	if string(bs) != string(want) {
		t.Fatalf("Marshal returned %v, want %v", bs, want)
	}
	bs = binco.MarshalTyped(marshallerTyped{dtm: 3})
	if len(bs) != 1 || bs[0] != 3 {
		t.Fatalf("MarshalTyped returned %v", bs)
	}
}

func TestValidatorFn(t *testing.T) {
	want := errors.New("invalid")
	var vl binco.Validator[int] = binco.ValidatorFn[int](func(n int) error {
		if n < 0 {
			return want
		}
		return nil
	})
	if err := vl.Validate(-1); !errors.Is(err, want) {
		t.Fatalf("Validate returned %v, want %v", err, want)
	}
	if err := vl.Validate(1); err != nil {
		t.Fatalf("Validate returned %v", err)
	}
}

func TestPtrMaps(t *testing.T) {
	a := new(int)
	b := new(int)
	m := binco.NewPtrMap()
	rev := binco.NewReversePtrMap()

	if _, ok := m.Get(unsafe.Pointer(a)); ok {
		t.Fatal("Get of an unknown pointer returned ok")
	}
	if id := m.Put(unsafe.Pointer(a)); id != 0 {
		t.Fatalf("first Put returned %d, want 0", id)
	}
	if id := m.Put(unsafe.Pointer(b)); id != 1 {
		t.Fatalf("second Put returned %d, want 1", id)
	}
	if id, ok := m.Get(unsafe.Pointer(a)); !ok || id != 0 {
		t.Fatalf("Get returned (%d, %v), want (0, true)", id, ok)
	}
	if m.Len() != 2 {
		t.Fatalf("Len returned %d, want 2", m.Len())
	}

	rev.Put(1, unsafe.Pointer(b))
	if p, ok := rev.Get(1); !ok || p != unsafe.Pointer(b) {
		t.Fatalf("reverse Get returned (%v, %v)", p, ok)
	}
	if rev.Len() != 1 {
		t.Fatalf("reverse Len returned %d, want 1", rev.Len())
	}
}

func TestVarintLimits(t *testing.T) {
	if got, want := binco.UintMaxVarintLen(), 10; got != want {
		t.Fatalf("UintMaxVarintLen returned %d, want %d", got, want)
	}
	if got, want := binco.UintMaxLastByte(), byte(1); got != want {
		t.Fatalf("UintMaxLastByte returned %d, want %d", got, want)
	}
}

func TestErrors(t *testing.T) {
	if binco.ErrorPrefix != "binco: " {
		t.Fatalf("unexpected ErrorPrefix %q", binco.ErrorPrefix)
	}
	if got, want := binco.ErrOverflow.Error(), "binco: overflow"; got != want {
		t.Fatalf("unexpected error text %q, want %q", got, want)
	}
	unexpected := binco.NewUnexpectedDTMError(7)
	if unexpected.DTM() != 7 {
		t.Fatalf("unexpected DTM %d", unexpected.DTM())
	}
	if got, want := unexpected.Error(), "binco: unexpected DTM 7"; got != want {
		t.Fatalf("unexpected error text %q, want %q", got, want)
	}
	wrong := binco.NewWrongDTMError(1, 2)
	if wrong.Expected() != 1 || wrong.Actual() != 2 {
		t.Fatalf("wrong DTM expected/actual %d/%d", wrong.Expected(), wrong.Actual())
	}
	if got, want := wrong.Error(), "binco: wrong DTM, expected 1, got 2"; got != want {
		t.Fatalf("unexpected error text %q, want %q", got, want)
	}
}
