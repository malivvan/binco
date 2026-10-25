package main

type Foo FooV2

func (f Foo) MarshalBinco(bs []byte) (n int) {
	return FooSer.Marshal(f, bs)
}

func (f Foo) SizeBinco() (size int) {
	return FooSer.Size(f)
}

func (f Foo) String() string {
	return FooV2(f).Str
}
