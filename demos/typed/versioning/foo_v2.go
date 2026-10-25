package main

type FooV2 struct {
	Str string
}

func (v FooV2) MarshalTypedBinco(bs []byte) (n int) {
	return FooV2TypedSerializer.Marshal(v, bs)
}

func (v FooV2) SizeTypedBinco() (size int) {
	return FooV2TypedSerializer.Size(v)
}
