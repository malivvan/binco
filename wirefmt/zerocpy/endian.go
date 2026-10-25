package zerocpy

import (
	"unsafe"

	"github.com/malivvan/binco"
)

// The package refuses to run on a big-endian host: the memory access below assumes little-endian byte order.
func init() {
	var probe uint16 = 1
	if *(*byte)(unsafe.Pointer(&probe)) != 1 {
		panic(binco.ErrUnsupportedArch)
	}
}
