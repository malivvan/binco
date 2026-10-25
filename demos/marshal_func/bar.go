package main

type Bar struct {
	Str string
}

func (b Bar) MarshalBinco(bs []byte) (n int) {
	return BarSer.Marshal(b, bs)
}

func (b Bar) SizeBinco() (size int) {
	return BarSer.Size(b)
}
