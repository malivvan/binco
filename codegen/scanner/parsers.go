package scanner

import (
	"regexp"
	"strings"

	"github.com/malivvan/binco/codegen/typename"
)

// DefinedType parses "pkg.Type", "pkg.Type[params]" and the pointer forms of both.
type DefinedType[T QualifiedName] string

func (s DefinedType[T]) splitTypeParams(params string) (result []T) {
	var buf strings.Builder
	depth := 0
	for _, char := range params {
		switch char {
		case '[':
			depth++
		case ']':
			depth--
		case ',':
			if depth == 0 {

				result = append(result, T(strings.TrimSpace(buf.String())))
				buf.Reset()
				continue
			}
		}
		buf.WriteRune(char)
	}

	if buf.Len() > 0 {
		result = append(result, T(strings.TrimSpace(buf.String())))
	}
	return
}

func (s DefinedType[T]) makeDefinedType(match []string) (t TypeInfo[T]) {
	t.PkgPath = typename.PkgPath(match[2])
	t.Stars = match[1]
	t.Package = typename.Package(match[3])
	t.Name = typename.TypeName(match[4])
	t.Params = s.splitTypeParams(match[5])
	return
}

func (s DefinedType[T]) Parse() (t TypeInfo[T], ok bool) {
	match := FullNameRe.FindStringSubmatch(string(s))
	if len(match) != 6 {
		return
	}
	return s.makeDefinedType(match), true
}

// ParseDefinedType reports whether name is a defined type.
func ParseDefinedType[T QualifiedName](name T) (t TypeInfo[T], ok bool) {
	return DefinedType[T](name).Parse()
}

// FullNamePattern matches a defined type name.
const (
	// Works for strings like "*github.com/user-name/project/some/some.Type" and
	// "foo.Type" (when module name is just foo).
	FullNamePattern = `^(\**)(?:([.\w/-]+)/)*(?:([\w-]+)\.)([\w]+)(?:·\d+)*(?:\[(.+)\](?:·\d+)*)?$`
)

// FullNameRe matches a defined type name.
//
// It is compiled once: Parse runs for every type name the generator meets, and
// compiling the expression per call was measurable.
var FullNameRe = regexp.MustCompile(FullNamePattern)

// ContainerType parses array, slice and map names.
type ContainerType[T QualifiedName] string

func (s ContainerType[T]) ParseMap() (t TypeInfo[T], keyType, elemType T, ok bool) {
	str := string(s)

	i := 0
	for i < len(str) && str[i] == '*' {
		i++
	}
	starsTmp := str[:i]
	str = str[i:]

	const prefix = "map["
	if !strings.HasPrefix(str, prefix) {
		ok = false
		return
	}
	stars := starsTmp

	start := len(prefix)
	depth := 1
	j := start

	for j < len(str) && depth > 0 {
		switch str[j] {
		case '[':
			depth++
		case ']':
			depth--
		}
		j++
	}

	if depth != 0 {
		ok = false
		return
	}

	t = TypeInfo[T]{
		Stars: stars,
	}
	t.KeyType = T(strings.TrimSpace(str[start : j-1]))
	t.ElemType = T(strings.TrimSpace(str[j:]))
	t.Name = typename.TypeName("map[" + string(t.KeyType) + "]" + string(t.ElemType))

	keyType = t.KeyType
	elemType = t.ElemType
	ok = true
	return
}

func (s ContainerType[T]) ParseSlice() (t TypeInfo[T], elemType T, ok bool) {
	match := SliceRe.FindStringSubmatch(string(s))
	if len(match) != 4 {
		return
	}
	t = TypeInfo[T]{
		Stars:    match[1],
		Name:     typename.TypeName(match[2]),
		ElemType: T(match[3]),
	}
	elemType = t.ElemType
	ok = true
	return
}

func (s ContainerType[T]) ParseArray() (t TypeInfo[T], elemType T, ok bool) {
	match := ArrayRe.FindStringSubmatch(string(s))
	if len(match) != 5 {
		return
	}
	t = TypeInfo[T]{
		Stars:     match[1],
		Name:      typename.TypeName(match[2]),
		ArrLength: match[3],
		ElemType:  T(match[4]),
	}
	elemType = t.ElemType
	ok = true
	return
}

// ParseContainerType reports whether name is a container, and of which kind.
func ParseContainerType[T QualifiedName](name T) (t TypeInfo[T], keyType, elemStr T,
	kind Kind, ok bool) {
	if t, elemStr, ok = ContainerType[T](name).ParseArray(); ok {
		kind = KindArray
		return
	}
	if t, elemStr, ok = ContainerType[T](name).ParseSlice(); ok {
		kind = KindSlice
		return
	}
	if t, keyType, elemStr, ok = ContainerType[T](name).ParseMap(); ok {
		kind = KindMap
		return
	}
	return
}

// The patterns that match an array, a slice and a primitive name.
const (
	ArrayPattern = `^(\**)(\[(\d+)\](.+$))`
	SlicePattern = `^(\**)(\[\](.+$))`
)

// ArrayRe matches an array name.
var (
	ArrayRe = regexp.MustCompile(ArrayPattern)
	SliceRe = regexp.MustCompile(SlicePattern)
)

// PrimitiveType parses the built-in scalar names.
type PrimitiveType[T QualifiedName] string

func (s PrimitiveType[T]) Parse() (t TypeInfo[T], ok bool) {
	match := PrimitiveRe.FindStringSubmatch(string(s))
	if len(match) != 3 {
		return
	}
	t = TypeInfo[T]{
		Stars: match[1],
		Name:  typename.TypeName(match[2]),
	}
	ok = true
	return
}

// ParsePrimitiveType reports whether name is a built-in scalar.
func ParsePrimitiveType[T QualifiedName](name T) (t TypeInfo[T], ok bool) {
	return PrimitiveType[T](name).Parse()
}

// PrimitivePattern matches a built-in scalar name.
const PrimitivePattern = `^(\**)(uint8|uint16|uint32|uint64|int8|int16|int32|` +
	`int64|uint|int|byte|bool|string|float32|float64)`

// PrimitiveRe matches a built-in scalar name.
var PrimitiveRe = regexp.MustCompile(PrimitivePattern)
