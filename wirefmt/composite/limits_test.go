package composite_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	bslopts "github.com/malivvan/binco/wirefmt/config/byte_slice"
	mapopts "github.com/malivvan/binco/wirefmt/config/map"
	slopts "github.com/malivvan/binco/wirefmt/config/slice"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
	"github.com/malivvan/binco/wirefmt/varnum"
)

var errOverLimit = errors.New("length over limit")

func TestValidatedLengthsRejectBeforeReading(t *testing.T) {
	const limit = 8

	validator := binco.ValidatorFn[int](func(length int) error {
		if length > limit {
			return errOverLimit
		}
		return nil
	})
	lengthPrefix := func(t *testing.T, length int) []byte {
		t.Helper()
		var buf bytes.Buffer
		if _, err := varnum.PositiveInt.Marshal(length, &buf); err != nil {
			t.Fatalf("Marshal length: %v", err)
		}
		return buf.Bytes()
	}

	t.Run("string", func(t *testing.T) {
		ser := composite.NewValidStringSerializer(
			stropts.WithLenValidator(validator))
		assertRejected(t, ser, lengthPrefix(t, limit+1))
	})

	t.Run("byte slice", func(t *testing.T) {
		ser := composite.NewValidByteSliceSerializer(
			bslopts.WithLenValidator(validator))
		assertRejected(t, ser, lengthPrefix(t, limit+1))
	})

	t.Run("slice", func(t *testing.T) {
		ser := composite.NewValidSliceSerializer(varnum.Int,
			slopts.WithLenValidator[int](validator))
		assertRejected(t, ser, lengthPrefix(t, limit+1))
	})

	t.Run("map", func(t *testing.T) {
		ser := composite.NewValidMapSerializer(varnum.Int, varnum.Int,
			mapopts.WithLenValidator[int, int](validator))
		assertRejected(t, ser, lengthPrefix(t, limit+1))
	})
}

func assertRejected[T any](t *testing.T, ser wirefmt.Serializer[T], bs []byte) {
	t.Helper()
	_, _, err := ser.Unmarshal(bytes.NewReader(bs))
	if !errors.Is(err, errOverLimit) {
		t.Fatalf("Unmarshal error = %v, want %v", err, errOverLimit)
	}
}
