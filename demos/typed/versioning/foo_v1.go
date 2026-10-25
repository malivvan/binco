package main

type FooV1 struct {
	Num int
}

func (v FooV1) MarshalTypedBinco(bs []byte) (n int) {
	return FooV1TypedSer.Marshal(v, bs)
}

func (v FooV1) SizeTypedBinco() (size int) {
	return FooV1TypedSer.Size(v)
}
