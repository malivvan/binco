package codegen

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/spec"
	cnvtr "github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
)

// SerializerFns answers the naming questions the templates ask: how a serializer, a typed serializer or a type identifier is named, and which codec a Go type maps to.
//
// Every method that can fail returns an error instead of panicking. The
// templates propagate it and Generate wraps it, so a bad name is reported with
// the type that caused it.
type SerializerFns struct {
	converter       cnvtr.TypeNameConverter
	builder         TypeBuilder
	foreignTypes    map[typename.FullName]struct{}
	serializerNames map[typename.FullName]string
	names           map[string]typename.FullName
	utilFns         TmplFns
	gopts           genopts.Options
}

func (f SerializerFns) lowerFirst(str string) string {
	if str == "" {
		return str
	}
	r := []rune(str)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func (f SerializerFns) upperFirst(str string) string {
	if str == "" {
		return str
	}
	r := []rune(str)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// codecName returns the base name the generated code uses for a type, and
// records it so that two types that would produce the same name are rejected.
//
// Two types with the same name in different packages used to collapse into one
// serializer. The name is now checked: a collision names both types and asks for
// an explicit name through genopts.WithSerializer.
func (f SerializerFns) codecName(name typename.FullName) (string, error) {
	if serializerName, pst := f.serializerNames[name]; pst {
		return serializerName, nil
	}
	var (
		rname string
		err   error
	)
	if _, pst := f.foreignTypes[name]; pst {
		rname = string(typename.RelativeName(name.TypeName()).WithoutSquares())
	} else {
		var rel typename.RelativeName
		rel, err = f.converter.ConvertToRelativeName(name)
		if err != nil {
			return "", err
		}
		rname = rel.WithoutSquares()
	}
	if f.names != nil {
		if other, dup := f.names[rname]; dup && other != name {
			return "", fmt.Errorf("%w: %s and %s would both be generated as %q; give one of them a name with genopts.WithSerializer",
				ErrNameCollision, other, name, rname)
		}
		f.names[rname] = name
	}
	return rname, nil
}

// MapOpts renders the option list of a map serializer.
func (f SerializerFns) MapOpts(anonType spec.AnonType) (string, error) {
	var (
		b        = strings.Builder{}
		keyType  typename.RelativeName
		elemType typename.RelativeName
		err      error
	)
	if keyType, err = f.converter.ConvertToRelativeName(anonType.KeyType); err != nil {
		return "", err
	}
	if elemType, err = f.converter.ConvertToRelativeName(anonType.ElemType); err != nil {
		return "", err
	}
	if anonType.LenSerializer != "" && anonType.LenSerializer != "nil" {
		fmt.Fprintf(&b, "mapopts.WithLenSerializer[%v, %v](%v)", keyType, elemType,
			anonType.LenSerializer)
	}
	if anonType.LenVl != "" && anonType.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "mapopts.WithLenValidator[%v, %v](%v)", keyType, elemType,
			anonType.LenVl)
	}
	if anonType.KeyVl != "" && anonType.KeyVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "mapopts.WithKeyValidator[%v, %v](%v)", keyType, elemType,
			anonType.KeyVl)
	}
	if anonType.ElemVl != "" && anonType.ElemVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "mapopts.WithValueValidator[%v, %v](%v)", keyType, elemType, anonType.ElemVl)
	}
	return b.String(), nil
}

// SliceOpts renders the option list of a slice serializer.
func (f SerializerFns) SliceOpts(anonType spec.AnonType) (string, error) {
	var (
		b        = strings.Builder{}
		elemType typename.RelativeName
		err      error
	)
	if elemType, err = f.converter.ConvertToRelativeName(anonType.ElemType); err != nil {
		return "", err
	}
	if anonType.LenSerializer != "" && anonType.LenSerializer != "nil" {
		fmt.Fprintf(&b, "slopts.WithLenSerializer[%v](%v)", elemType, anonType.LenSerializer)
	}
	if anonType.LenVl != "" && anonType.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "slopts.WithLenValidator[%v](%v)", elemType, anonType.LenVl)
	}
	if anonType.ElemVl != "" && anonType.ElemVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "slopts.WithElemValidator[%v](%v)", elemType, anonType.ElemVl)
	}
	return b.String(), nil
}

// ByteSliceOpts renders the option list of a byte slice serializer.
func (f SerializerFns) ByteSliceOpts(anonType spec.AnonType) (string, error) {
	b := strings.Builder{}
	if anonType.LenSerializer != "" && anonType.LenSerializer != "nil" {
		fmt.Fprintf(&b, "bslopts.WithLenSerializer(%v)", anonType.LenSerializer)
	}
	if anonType.LenVl != "" && anonType.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "bslopts.WithLenValidator(%v)", anonType.LenVl)
	}
	return b.String(), nil
}

// ArrayOpts renders the option list of an array serializer.
func (f SerializerFns) ArrayOpts(anonType spec.AnonType) (string, error) {
	var (
		b        = strings.Builder{}
		elemType typename.RelativeName
		err      error
	)
	if elemType, err = f.converter.ConvertToRelativeName(anonType.ElemType); err != nil {
		return "", err
	}
	if anonType.LenSerializer != "" && anonType.LenSerializer != "nil" {
		fmt.Fprintf(&b, "arropts.WithLenSerializer[%v](%v)", elemType, anonType.LenSerializer)
	}
	if anonType.ElemVl != "" && anonType.ElemVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "arropts.WithElemValidator[%v](%v)", elemType, anonType.ElemVl)
	}
	return b.String(), nil
}

// StringOpts renders the option list of a string serializer.
func (f SerializerFns) StringOpts(anonData spec.AnonType) (string, error) {
	b := strings.Builder{}
	if anonData.LenSerializer != "" && anonData.LenSerializer != "nil" {
		fmt.Fprintf(&b, "stropts.WithLenSerializer(%v)", anonData.LenSerializer)
	}
	if anonData.LenVl != "" && anonData.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "stropts.WithLenValidator(%v)", anonData.LenVl)
	}
	return b.String(), nil
}

// FieldTmplPipe builds the data one field template renders.
func (f SerializerFns) FieldTmplPipe(structType spec.StructType,
	fieldType spec.FieldType, index int, opts genopts.Options) FieldTmplPipe {
	pipe := FieldTmplPipe{
		Val:         f.Val(structType),
		FieldsCount: len(structType.SerializedFields()),
		Field:       fieldType,
		Index:       index,
		Gops:        opts,
	}
	// index is the position among serialized fields; the options are indexed by
	// the position in the struct, which fieldType.Index carries.
	if len(structType.Sops.Fields) > 0 && fieldType.Index < len(structType.Sops.Fields) {
		pipe.Tops = structType.Sops.Fields[fieldType.Index].Type
	}
	return pipe
}

// RelName returns the name the generated file uses for a type.
func (f SerializerFns) RelName(name typename.FullName) (string, error) {
	rname, err := f.converter.ConvertToRelativeName(name)
	if err != nil {
		return "", err
	}
	return string(rname), nil
}

// Elem returns the serializer expression for the element type of a container.
func (f SerializerFns) Elem(anonType spec.AnonType, opts genopts.Options) (string, error) {
	return f.Of(anonType.ElemType, tpopts.Options{}, opts)
}

// Key returns the serializer expression for the key type of a map.
func (f SerializerFns) Key(anonType spec.AnonType, opts genopts.Options) (string, error) {
	return f.Of(anonType.KeyType, tpopts.Options{}, opts)
}

// Of returns the expression that names the serializer of a type.
func (f SerializerFns) Of(fullName typename.FullName, tops tpopts.Options,
	opts genopts.Options,
) (string, error) {
	anonType, ok, err := f.builder.BuildAnonType(fullName, tops)
	if err != nil {
		return "", fmt.Errorf("cannot get a serializer for %v: %w", fullName, err)
	}
	if ok {
		return string(anonType.SerializerName), nil
	}
	var (
		pkg            string
		serializerName string
		rules          map[string]string
	)

	switch opts.Mode {
	case genopts.ModeSafe:
		rules = portableRules(tops)
	case genopts.ModeNotUnsafe:
		rules = hybridRules(tops, opts)
	case genopts.ModeUnsafe:
		rules = nativeRules(tops, opts)
	}
	pkg, pst := rules[string(fullName)]
	if !pst {
		return f.Var(fullName)
	}
	switch fullName {
	case "int", "int64", "int32", "int16", "int8":
		serializerName = f.upperFirst(string(fullName))
		if tops.NumEncoding == tpopts.NumEncodingVarintPositive {
			serializerName = "Positive" + serializerName
		}
	case "uint", "uint64", "uint32", "uint16", "uint8", "float64", "float32", "byte":
		serializerName = f.upperFirst(string(fullName))
	case "bool", "string":
		serializerName = f.upperFirst(string(fullName))
	case "[]byte", "[]uint8":
		serializerName = "ByteSlice"
	case "time.Time":
		serializerName = rules["time_serializer"]
	default:
		return "", fmt.Errorf("internal bug: unpredicted case %v", fullName)
	}
	return pkg + "." + serializerName, nil
}

// Val is the name of the value parameter in generated methods.
func (f SerializerFns) Val(tp any) string {
	return "v"
}

// Type returns the name of the generated serializer type.
func (f SerializerFns) Type(fullName typename.FullName) (string, error) {
	v, err := f.Var(fullName)
	if err != nil {
		return "", err
	}
	return f.lowerFirst(v), nil
}

// SchemaVar is the name of the generated constant that identifies the wire
// layout of a struct. It is built from the codec name rather than from the
// relative type name, because a relative name of a foreign type contains a
// package qualifier ("pkg.Foo") and is not a valid identifier.
func (f SerializerFns) SchemaVar(fullName typename.FullName) (string, error) {
	name, err := f.codecName(fullName)
	if err != nil {
		return "", err
	}
	return name + "SchemaFingerprint", nil
}

// DTMVar returns the name of the generated identifier constant.
func (f SerializerFns) DTMVar(fullName typename.FullName) (string, error) {
	name, err := f.codecName(fullName)
	if err != nil {
		return "", err
	}
	return name + SuffixDTM, nil
}

// TypedVar returns the name of the generated typed serializer.
func (f SerializerFns) TypedVar(fullName typename.FullName) (string, error) {
	name, err := f.codecName(fullName)
	if err != nil {
		return "", err
	}
	return name + SuffixTypedBinco, nil
}

// Var returns the name of the generated serializer variable.
func (f SerializerFns) Var(name typename.FullName) (string, error) {
	codec, err := f.codecName(name)
	if err != nil {
		return "", err
	}
	return codec + SuffixBinco, nil
}

// NewSerializerFns prepares the naming helpers for one generator run.
func NewSerializerFns(converter cnvtr.TypeNameConverter,
	builder TypeBuilder,
	foreignTypes map[typename.FullName]struct{},
	utilFns TmplFns,
	gopts genopts.Options,
) (SerializerFns, error) {
	serializerNames := map[typename.FullName]string{}
	for t, serializerName := range gopts.SerializerNames {
		cname, err := typename.TypeCompleteName(t)
		if err != nil {
			return SerializerFns{}, err
		}
		fullName, err := converter.ConvertToFullName(cname)
		if err != nil {
			return SerializerFns{}, err
		}
		serializerNames[fullName] = serializerName
	}

	return SerializerFns{
		converter:       converter,
		builder:         builder,
		foreignTypes:    foreignTypes,
		serializerNames: serializerNames,
		names:           map[string]typename.FullName{},
		utilFns:         utilFns,
		gopts:           gopts,
	}, nil
}

// TypeBuilder turns reflect types into the specifications the templates consume.
type TypeBuilder interface {
	BuildDefinedType(t reflect.Type, tops tpopts.Options) (spec.DefinedType, error)
	BuildAnonType(name typename.FullName, topts tpopts.Options) (
		spec.AnonType, bool, error)
}

// The suffixes appended to a type name: SuffixBinco for a serializer, SuffixTypedBinco for a typed serializer and SuffixDTM for a type identifier.
const (
	SuffixBinco      = "Binco"
	SuffixTypedBinco = "TypedBinco"
	SuffixDTM        = "DTM"
)
