// Command different puts unrelated types behind type identifiers and dispatches on them.
package main

import (
	"errors"
	"fmt"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/typeid"
)

func main() {

	fooBS := binco.Encode(FooTypedSer, Foo{Num: 5})
	barBS := binco.Encode(BarTypedSer, Bar{Str: "hello world"})
	fmt.Printf("Foo: %v (DTM %d)\n", fooBS, FooDTM)
	fmt.Printf("Bar: %v (DTM %d)\n\n", barBS, BarDTM)

	for _, bs := range [][]byte{fooBS, barBS} {
		if err := receive(bs); err != nil {
			panic(err)
		}
	}

	_, _, err := BarTypedSer.Unmarshal(fooBS)
	if wrong, ok := errors.AsType[binco.WrongDTMError](err); ok {
		fmt.Printf("Bar reader got a Foo payload: expected %d, actual %d\n",
			wrong.Expected(), wrong.Actual())
	}

	dtm, n, err := typeid.DTMSer.Unmarshal(fooBS)
	if err != nil {
		panic(err)
	}
	fmt.Printf("DTM prefix: %d (%d byte), data: %v\n", dtm, n, fooBS[n:])
}

func receive(bs []byte) (err error) {

	dtm, n, err := typeid.DTMSer.Unmarshal(bs)
	if err != nil {
		return
	}

	switch dtm {
	case FooDTM:
		foo, _, err := FooTypedSer.UnmarshalData(bs[n:])
		if err != nil {
			return err
		}
		fmt.Printf("received Foo %+v\n", foo)
	case BarDTM:
		bar, _, err := BarTypedSer.UnmarshalData(bs[n:])
		if err != nil {
			return err
		}
		fmt.Printf("received Bar %+v\n", bar)
	default:
		return fmt.Errorf("unexpected message: %w", binco.NewUnexpectedDTMError(dtm))
	}
	return nil
}
