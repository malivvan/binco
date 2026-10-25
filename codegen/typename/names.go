package typename

import (
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/malivvan/binco/codegen/classifier"
)

// CompleteName is a type name as reflect produces it: the package path, the package name and the type name.
type CompleteName string

// TypeCompleteName builds the complete name of t.
func TypeCompleteName(t reflect.Type) (name CompleteName, err error) {
	var stars string
	if strings.HasPrefix(t.String(), "*") {
		if stars, t, err = parsePtr(t); err != nil {
			return
		}
	}
	return typeCompleteName(stars, t)
}

// MustTypeCompleteName is TypeCompleteName, panicking on failure.
func MustTypeCompleteName(t reflect.Type) (name CompleteName) {
	name, err := TypeCompleteName(t)
	if err != nil {
		panic(err)
	}
	return
}

// BaseTypeCompleteName builds the complete name of the underlying type of a defined type.
func BaseTypeCompleteName(t reflect.Type) (name CompleteName, err error) {
	if !classifier.DefinedType(t) {
		err = ErrTypeMismatch
		return
	}
	switch {
	case classifier.PtrType(t):
		var stars string
		if stars, t, err = parsePtr(t); err != nil {
			return
		}
		return typeCompleteName(stars, t)
	case classifier.PrimitiveType(t):
		return primitiveTypeCompleteName("", t)
	case classifier.ArrayType(t):
		return arrayTypeCompleteName("", t)
	case classifier.SliceType(t):
		return sliceTypeCompleteName("", t)
	case classifier.MapType(t):
		return mapTypeCompleteName("", t)
	default:
		err = NewUnsupportedTypeError(t)
		return
	}
}

// ParsePtr strips the pointers from t and reports how many there were.
func ParsePtr(t reflect.Type) (stars string, at reflect.Type) {
	at = t
	for {
		if !classifier.PtrType(at) {
			return
		}
		at = at.Elem()
		stars += "*"
		continue
	}
}

// parsePtr is ParsePtr but rejects more than one pointer.
func parsePtr(t reflect.Type) (stars string, at reflect.Type, err error) {
	stars, at = ParsePtr(t)
	if len(stars) > 1 {
		err = NewMultiPointerError(t)
	}
	return
}

// typeCompleteName builds a complete name for any supported type.
func typeCompleteName(stars string, t reflect.Type) (name CompleteName,
	err error) {
	switch {
	case classifier.DefinedType(t):
		return definedTypeCompleteName(stars, t)
	case classifier.PrimitiveType(t):
		return primitiveTypeCompleteName(stars, t)
	case classifier.ArrayType(t):
		return arrayTypeCompleteName(stars, t)
	case classifier.SliceType(t):
		return sliceTypeCompleteName(stars, t)
	case classifier.MapType(t):
		return mapTypeCompleteName(stars, t)
	default:
		err = NewUnsupportedTypeError(t)
		return
	}
}

// definedTypeCompleteName builds a complete name for a defined type.
func definedTypeCompleteName(stars string, t reflect.Type) (name CompleteName,
	err error) {
	name = CompleteName(stars + filepath.Join(t.PkgPath(), t.String()))
	return
}

// primitiveTypeCompleteName builds a complete name for a built-in type.
func primitiveTypeCompleteName(stars string, t reflect.Type) (
	name CompleteName, err error) {
	name = CompleteName(stars + t.Kind().String())
	return
}

// arrayTypeCompleteName builds a complete name for an array type.
func arrayTypeCompleteName(stars string, t reflect.Type) (name CompleteName,
	err error) {
	elemName, err := TypeCompleteName(t.Elem())
	if err != nil {
		return
	}
	name = CompleteName(stars + "[" + strconv.Itoa(t.Len()) + "]" +
		string(elemName))
	return
}

// sliceTypeCompleteName builds a complete name for a slice type.
func sliceTypeCompleteName(stars string, t reflect.Type) (name CompleteName,
	err error) {
	elemName, err := TypeCompleteName(t.Elem())
	if err != nil {
		return
	}
	name = CompleteName(stars + "[]" + string(elemName))
	return
}

// mapTypeCompleteName builds a complete name for a map type.
func mapTypeCompleteName(stars string, t reflect.Type) (name CompleteName,
	err error) {
	keyName, err := TypeCompleteName(t.Key())
	if err != nil {
		return
	}
	elemName, err := TypeCompleteName(t.Elem())
	if err != nil {
		return
	}
	name = CompleteName(stars + "map[" + string(keyName) + "]" +
		string(elemName))
	return
}

// FullName is "pkg.Type" as the generator templates write it.
type FullName string

func (n FullName) TypeName() TypeName {
	match := pkgTypeNameRe.FindStringSubmatch(string(n))
	if len(match) != 3 {
		return TypeName(n)
	}
	return TypeName(match[2])
}

func (n FullName) Package() Package {
	match := pkgTypeNameRe.FindStringSubmatch(string(n))
	if len(match) != 3 {
		return ""
	}
	return Package(match[1])
}

// MakeFullName joins a package and a type name.
func MakeFullName(pkg Package, name TypeName) FullName {
	if pkg == "" {
		return FullName(name)
	}
	return FullName(strings.Join([]string{string(pkg), string(name)}, "."))
}

// RelativeName is a full name that omits the package of the generated file.
type RelativeName FullName

func (n RelativeName) WithoutSquares() (str string) {
	str = string(n)
	before, _, ok := strings.Cut(str, "[")
	if !ok {
		return str
	}
	return before
}

// TypeName is a type name without a package.
type TypeName string

// pkgTypeNamePattern splits "pkg.Type" into its two parts.
const pkgTypeNamePattern = `^(\w+)\.(.+)$`

// pkgTypeNameRe implements pkgTypeNamePattern.
var pkgTypeNameRe = regexp.MustCompile(pkgTypeNamePattern)
