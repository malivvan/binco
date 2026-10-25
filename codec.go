package binco

// Serializer is the contract every codec in this module implements: encode into a caller-owned buffer, decode from one, report the encoded size, and skip a value.
type Serializer[T any] interface {
	Marshal(t T, bs []byte) (n int)
	Unmarshal(bs []byte) (t T, n int, err error)
	Size(t T) (size int)
	Skip(bs []byte) (n int, err error)
}

// Marshaller lets a value encode itself, which is what binco.Marshal and generated code rely on.
type Marshaller interface {
	MarshalBinco(bs []byte) (n int)
	SizeBinco() (size int)
}

// MarshallerTyped is Marshaller with a type identifier in front of the payload.
type MarshallerTyped interface {
	MarshalTypedBinco(bs []byte) (n int)
	SizeTypedBinco() (size int)
}

// Marshal allocates a buffer of SizeBinco bytes and lets v fill it.
func Marshal(v Marshaller) (bs []byte) {
	bs = make([]byte, v.SizeBinco())
	v.MarshalBinco(bs)
	return
}

// MarshalTyped is Marshal with the type identifier included.
func MarshalTyped(v MarshallerTyped) (bs []byte) {
	bs = make([]byte, v.SizeTypedBinco())
	v.MarshalTypedBinco(bs)
	return
}

// Encode returns a fresh buffer holding the encoded v.
func Encode[T any](s Serializer[T], v T) (bs []byte) {
	bs = make([]byte, s.Size(v))
	s.Marshal(v, bs)
	return
}

// Decode unmarshals v and returns only the value and the error.
func Decode[T any](s Serializer[T], bs []byte) (v T, err error) {
	v, _, err = s.Unmarshal(bs)
	return
}
