package main

type Bar struct {
	Str string
}

func (b Bar) MarshalBinco(bs []byte) (n int) {
	return BarSerializer.Marshal(b, bs)
}

func (b Bar) SizeBinco() (size int) {
	return BarSerializer.Size(b)
}
