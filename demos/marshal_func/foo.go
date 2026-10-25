// Foo marshals itself through binco.Marshaller and binco.MarshallerTyped.
package main

type Foo struct {
	Num int
}

func (f Foo) MarshalBinco(bs []byte) (n int) {
	return FooSerializer.Marshal(f, bs)
}

func (f Foo) SizeBinco() (size int) {
	return FooSerializer.Size(f)
}

func (f Foo) MarshalTypedBinco(bs []byte) (n int) {
	return FooTypedSerializer.Marshal(f, bs)
}

func (f Foo) SizeTypedBinco() (size int) {
	return FooTypedSerializer.Size(f)
}
