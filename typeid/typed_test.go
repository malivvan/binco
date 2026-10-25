package typeid_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const dtm binco.DTM = 7

func TestSer(t *testing.T) {
	ser := typeid.NewSer(dtm, varnum.Int)
	if ser.DTM() != dtm {
		t.Fatalf("DTM returned %d", ser.DTM())
	}
	bs := binco.Encode(ser, 300)
	if want := []byte{7, 216, 4}; len(bs) != len(want) || bs[0] != 7 {
		t.Fatalf("Marshal = %v, want %v", bs, want)
	}
	if len(bs) != ser.Size(300) {
		t.Fatalf("Size = %d, Marshal wrote %d", ser.Size(300), len(bs))
	}
	got, n, err := ser.Unmarshal(bs)
	if err != nil || n != len(bs) || got != 300 {
		t.Fatalf("Unmarshal = (%d, %d, %v)", got, n, err)
	}
	if n, err = ser.Skip(bs); err != nil || n != len(bs) {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}

	data, n, err := ser.UnmarshalData(bs[1:])
	if err != nil || n != 2 || data != 300 {
		t.Fatalf("UnmarshalData = (%d, %d, %v)", data, n, err)
	}
	if n, err = ser.SkipData(bs[1:]); err != nil || n != 2 {
		t.Fatalf("SkipData = (%d, %v)", n, err)
	}
}

func TestWrongDTM(t *testing.T) {
	ser := typeid.NewSer(dtm, varnum.Int)
	bs := binco.Encode(typeid.NewSer(dtm+1, varnum.Int), 1)

	_, n, err := ser.Unmarshal(bs)
	if err == nil {
		t.Fatal("Unmarshal of a foreign DTM returned no error")
	}
	wrong, ok := errors.AsType[binco.WrongDTMError](err)
	if !ok {
		t.Fatalf("Unmarshal returned %v, want WrongDTMError", err)
	}
	if wrong.Expected() != dtm || wrong.Actual() != dtm+1 {
		t.Fatalf("WrongDTMError carries (%d, %d)", wrong.Expected(), wrong.Actual())
	}
	if n != 1 {
		t.Fatalf("Unmarshal read %d bytes, want 1", n)
	}
	if _, err = ser.Skip(bs); err == nil {
		t.Fatal("Skip of a foreign DTM returned no error")
	}

	if _, _, err = ser.Unmarshal(nil); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal(nil) returned %v", err)
	}
}

func TestDTMSer(t *testing.T) {
	for _, v := range []binco.DTM{1, 300, binco.DTM(1 << 20)} {
		bs := binco.Encode(typeid.DTMSer, v)
		if len(bs) != typeid.DTMSer.Size(v) {
			t.Fatalf("Size = %d, Marshal wrote %d", typeid.DTMSer.Size(v), len(bs))
		}
		got, n, err := typeid.DTMSer.Unmarshal(bs)
		if err != nil || n != len(bs) || got != v {
			t.Fatalf("Unmarshal(%v) = (%d, %d, %v)", bs, got, n, err)
		}
		if n, err = typeid.DTMSer.Skip(bs); err != nil || n != len(bs) {
			t.Fatalf("Skip = (%d, %v)", n, err)
		}
	}
}
