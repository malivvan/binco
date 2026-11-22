package builder

import (
	"crypto/sha256"
	"errors"
	"strings"

	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/scanner"
	"github.com/malivvan/binco/codegen/spec"
	"github.com/malivvan/binco/codegen/typename"
)

// CollectAnonOp walks a type name and records every anonymous serializer the generator must declare.
type CollectAnonOp struct {
	m             map[spec.AnonSerializerName]spec.AnonType
	firstAnonType *spec.AnonType
	converter     TypeNameConverter
}

func (o *CollectAnonOp) validatorExpr(name typename.FullName, vl string) (string, error) {
	relName, err := o.converter.ConvertToRelativeName(name)
	if err != nil {
		return "", err
	}
	return "binco.ValidatorFn[" + string(relName) + "](" + vl + ")", nil
}

func (o *CollectAnonOp) putToMap(anonType spec.AnonType) {
	if o.firstAnonType == nil {
		o.firstAnonType = &anonType
	}
	o.m[anonType.SerializerName] = anonType
}

func (o *CollectAnonOp) makeMapType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options,
) (d spec.AnonType, ok bool, err error) {
	if usesAnyName(t.KeyType) {
		return d, false, NewAnyMapKeyError()
	}
	var (
		lenSerializer = "nil"
		lenVl         = "nil"
		keyVl         = "nil"
		elemVl        = "nil"
	)
	if tops.LenEnc != tpopts.NumEncodingUndefined {
		lenSerializer = tops.LenEnc.LenSerializer()
	}
	if tops.LenValidator != "" {
		if lenVl, err = o.validatorExpr("int", tops.LenValidator); err != nil {
			return d, false, err
		}
	}
	if tops.KeyValidator != "" {
		if keyVl, err = o.validatorExpr(t.KeyType, tops.KeyValidator); err != nil {
			return d, false, err
		}
	}
	if tops.ElemValidator != "" {
		if elemVl, err = o.validatorExpr(t.ElemType, tops.ElemValidator); err != nil {
			return d, false, err
		}
	}
	return spec.AnonType{
		SerializerName: hashedName(spec.AnonKindMap, t, tops),
		Kind:           spec.AnonKindMap,
		LenSerializer:  lenSerializer,
		LenVl:          lenVl,
		KeyType:        t.KeyType,
		KeyVl:          keyVl,
		ElemType:       t.ElemType,
		ElemVl:         elemVl,
	}, true, nil
}

func (o *CollectAnonOp) makeSliceType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options) (d spec.AnonType, ok bool, err error) {
	var (
		lenSerializer = "nil"
		lenVl         = "nil"
		elemVl        = "nil"
	)
	if tops.LenEnc != tpopts.NumEncodingUndefined {
		lenSerializer = tops.LenEnc.LenSerializer()
	}
	if tops.LenValidator != "" {
		if lenVl, err = o.validatorExpr("int", tops.LenValidator); err != nil {
			return d, false, err
		}
	}
	if tops.ElemValidator != "" {
		if elemVl, err = o.validatorExpr(t.ElemType, tops.ElemValidator); err != nil {
			return d, false, err
		}
	}
	return spec.AnonType{
		SerializerName: hashedName(spec.AnonKindSlice, t, tops),
		Kind:           spec.AnonKindSlice,
		LenSerializer:  lenSerializer,
		LenVl:          lenVl,
		ElemType:       t.ElemType,
		ElemVl:         elemVl,
	}, true, nil
}

func (o *CollectAnonOp) makeByteSliceType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options,
) (d spec.AnonType, ok bool, err error) {
	if tops.LenEnc != tpopts.NumEncodingUndefined ||
		tops.LenValidator != "" {
		var (
			lenSerializer = "nil"
			lenVl         = "nil"
		)
		if tops.LenEnc != tpopts.NumEncodingUndefined {
			lenSerializer = tops.LenEnc.LenSerializer()
		}
		if tops.LenValidator != "" {
			if lenVl, err = o.validatorExpr("int", tops.LenValidator); err != nil {
				return d, false, err
			}
		}
		return spec.AnonType{
			SerializerName: hashedName(spec.AnonKindByteSlice, t, tops),
			Kind:           spec.AnonKindByteSlice,
			LenSerializer:  lenSerializer,
			LenVl:          lenVl,
		}, true, nil
	}
	return
}

func (o *CollectAnonOp) makeArrayType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options,
) (d spec.AnonType, ok bool, err error) {
	var (
		lenSerializer = "nil"
		elemVl        = "nil"
	)
	if tops.LenEnc != tpopts.NumEncodingUndefined {
		lenSerializer = tops.LenEnc.LenSerializer()
	}
	if tops.ElemValidator != "" {
		if elemVl, err = o.validatorExpr(t.ElemType, tops.ElemValidator); err != nil {
			return d, false, err
		}
	}
	return spec.AnonType{
		SerializerName: hashedName(spec.AnonKindArray, t, tops),
		Kind:           spec.AnonKindArray,
		ArrType:        typename.MakeFullName(t.Package, t.Name),
		ArrLength:      t.ArrLength,
		LenSerializer:  lenSerializer,
		ElemType:       t.ElemType,
		ElemVl:         elemVl,
	}, true, nil
}

func (o *CollectAnonOp) makePrimType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options,
) (d spec.AnonType, ok bool, err error) {
	if string(t.Name) == "string" {
		lenSerializer := "nil"
		if tops.LenEnc != tpopts.NumEncodingUndefined &&
			tops.LenEnc != tpopts.NumEncodingVarintPositive {
			lenSerializer = tops.LenEnc.LenSerializer()
			ok = true
		}
		lenVl := "nil"
		if tops.LenValidator != "" {
			if lenVl, err = o.validatorExpr("int", tops.LenValidator); err != nil {
				return d, false, err
			}
			ok = true
		}
		if ok {
			return spec.AnonType{
				SerializerName: hashedName(spec.AnonKindString, t, tops),
				Kind:           spec.AnonKindString,
				LenSerializer:  lenSerializer,
				LenVl:          lenVl,
			}, ok, nil
		}
	}
	return
}

func (o *CollectAnonOp) makePtrType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options,
) (d spec.AnonType, ok bool, err error) {
	str := t.Stars + string(typename.MakeFullName(t.Package, t.Name))
	return spec.AnonType{
		SerializerName: hashedName(spec.AnonKindPtr, t, tops),
		ElemType:       typename.FullName(stripStar(str)),
		Kind:           spec.AnonKindPtr,
	}, true, nil
}

func (o *CollectAnonOp) processPtrType(typeInfo scanner.TypeInfo[typename.FullName],
	tops tpopts.Options) (err error) {
	d, _, _ := o.makePtrType(typeInfo, tops)
	o.putToMap(d)
	typeInfo.Stars = stripStar(typeInfo.Stars)
	return o.ProcessType(typeInfo, tpopts.Options{})
}

// FirstAnonType returns the first anonymous serializer the collector met.
func (o *CollectAnonOp) FirstAnonType() *spec.AnonType {
	return o.firstAnonType
}

// ProcessRightSquare closes a container in the scanned name.
func (o *CollectAnonOp) ProcessRightSquare() {}

// ProcessComma separates the key and element types of a map.
func (o *CollectAnonOp) ProcessComma() {}

// ProcessLeftSquare opens a container in the scanned name.
func (o *CollectAnonOp) ProcessLeftSquare() {}

// ProcessType records the anonymous serializer a type needs.
func (o *CollectAnonOp) ProcessType(typeInfo scanner.TypeInfo[typename.FullName],
	tops tpopts.Options) (err error) {
	var (
		anonType spec.AnonType
		ok       bool
	)
	if typeInfo.Stars != "" {
		return o.processPtrType(typeInfo, tops)
	}
	// KindAny has no case on purpose: the empty interface is encoded by the
	// generated union (anyBinco), not by an anonymous container serializer.
	switch typeInfo.Kind {
	case scanner.KindPrim:
		anonType, ok, err = o.makePrimType(typeInfo, tops)
	case scanner.KindArray:
		anonType, ok, err = o.makeArrayType(typeInfo, tops)
	case scanner.KindSlice:
		if typeInfo.ElemType == "byte" || typeInfo.ElemType == "uint8" {
			anonType, ok, err = o.makeByteSliceType(typeInfo, tops)
		} else {
			anonType, ok, err = o.makeSliceType(typeInfo, tops)
		}
	case scanner.KindMap:
		anonType, ok, err = o.makeMapType(typeInfo, tops)
	}
	if err != nil {
		return err
	}
	if ok {
		o.putToMap(anonType)
	}
	return
}

// NewCollectAnonOp returns a collector that writes into m.
func NewCollectAnonOp(m map[spec.AnonSerializerName]spec.AnonType,
	converter TypeNameConverter) *CollectAnonOp {
	return &CollectAnonOp{m: m, converter: converter}
}

// MakeAnonDataOp stops at the first anonymous serializer it meets, which is what a single field needs.
type MakeAnonDataOp struct {
	*CollectAnonOp
	anonType spec.AnonType
	ok       bool
}

// AnonType returns the anonymous serializer the probe found.
func (o *MakeAnonDataOp) AnonType() (anonType spec.AnonType, ok bool) {
	return o.anonType, o.ok
}

// ProcessType stops the scan once the first anonymous serializer is known.
func (o *MakeAnonDataOp) ProcessType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options) (err error) {
	if err = o.CollectAnonOp.ProcessType(t, tops); err != nil {
		return
	}
	if anonType := o.FirstAnonType(); anonType != nil {
		o.anonType = *anonType
		o.ok = true
	}
	return ErrStop
}

// NewMakeAnonOp returns a probe that reports the first anonymous type of a name.
func NewMakeAnonOp(converter TypeNameConverter) *MakeAnonDataOp {
	m := map[spec.AnonSerializerName]spec.AnonType{}
	return &MakeAnonDataOp{CollectAnonOp: NewCollectAnonOp(m, converter)}
}

// hashedName derives a stable, unique name from everything that changes the
// generated code of an anonymous serializer.
//
// The name keys the map of one serializer per distinct definition, so every
// dimension the templates read has to be part of it. Hashing the package of the
// length encoding was not enough: the default, varnum.Int and varnum.PositiveInt
// live in one package but produce different bytes. The array length was not part
// of the name at all, so [3]int and [4]int collapsed into one serializer and the
// generated code did not compile.
func hashedName(kind spec.AnonKind, t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options) spec.AnonSerializerName {
	var b strings.Builder
	for _, part := range []string{
		kind.String(),
		string(t.Name),
		string(t.PkgPath),
		string(t.Package),
		t.Stars,
		string(t.ElemType),
		string(t.KeyType),
		t.ArrLength,
		tops.LenEnc.LenSerializer(),
		tops.LenValidator,
		tops.ElemValidator,
		tops.KeyValidator,
	} {
		b.WriteString(part)
		b.WriteByte(0)
	}
	// SHA-256 truncated to 128 bits: the name only has to be unique and stable,
	// but md5 is flagged by every linter and offers no collision guarantee.
	sum := sha256.Sum256([]byte(b.String()))
	return spec.AnonSerializerName(kind.String() + Base64KeywordEncoding.EncodeToString(sum[:16]))
}

// stripStar removes one leading star.
func stripStar(s string) string {
	return strings.TrimPrefix(s, "*")
}

// ErrStop tells the scanner to stop: the probe already has its answer.
var ErrStop = errors.New("stop")

// anyProbe records whether a scanned name mentions the empty interface.
type anyProbe struct{ saw bool }

func (p *anyProbe) ProcessType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options,
) error {
	if t.Kind == scanner.KindAny {
		p.saw = true
	}
	return nil
}

func (p *anyProbe) ProcessLeftSquare()  {}
func (p *anyProbe) ProcessComma()       {}
func (p *anyProbe) ProcessRightSquare() {}

// usesAnyName reports whether name mentions the empty interface anywhere.
func usesAnyName(name typename.FullName) bool {
	p := &anyProbe{}
	_ = scanner.Scan(scanner.Config{}, name, p, tpopts.Options{})
	return p.saw
}
