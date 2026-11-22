# Plan: generate codecs for `any` (empty-interface) fields via DTM dispatch

## 1. Goal

Make the code generator accept a struct field (or defined type) whose type is the
**predeclared empty interface** — `any` / `interface{}` — and encode it as a DTM
followed by the concrete implementation's payload, exactly like the existing
generated codec for a *named* interface.

Today `g.AddStruct(reflect.TypeFor[Event]())` for

```go
type Event struct {
	Kind    string
	Payload any
	Meta    map[string]any
	Items   []any
}
```

fails in `codegen/typename.TypeCompleteName` with

```
can't get a name for the interface {} type
```

because `classifier.DefinedType(any)` is false (`any` has no `PkgPath`) and
`typeCompleteName` has no case for it. That is the only reason a caller has to
keep a heterogeneous map (`map[string]any`) as a hand-encoded `[]byte` envelope
(e.g. eidos' `internal/record` union).

After this change:

```go
g.RegisterInterface(reflect.TypeFor[any](),
	intropts.WithStructImpl(reflect.TypeFor[Login]()),
	intropts.WithDefinedTypeImpl(reflect.TypeFor[Count]()),
)
g.AddStruct(reflect.TypeFor[Event]()) // Payload, Meta and Items are generated
```

* `Payload any` → writes `LoginDTM`/`CountDTM` + the implementation payload;
  `nil` is the reserved id `0`.
* `Meta map[string]any` → the ordinary generated map codec whose value codec is
  the `any` dispatcher.
* `Items []any` → the ordinary generated slice codec whose element codec is the
  `any` dispatcher.

**Wire format is unchanged.** An `any` value uses the existing named-interface
encoding (DTM + payload, `0` for nil), so no golden vector, README §13 table or
data layout changes. There is **no new runtime codec for `any`** — dispatch is
generated, like it is for a named interface.

**One pre-existing bug must be fixed first.** The generated interface codecs
write `0` for a nil value but read the identifier with `typeid.DTMSerializer`,
which rejects `0` with `binco.ErrOverflow` (`typeid/meta.go`). The `case 0:` in
`interface_unmarshal.tmpl`/`interface_skip.tmpl` is therefore unreachable and a
nil interface does **not** round-trip today — for named interfaces either. Since
`any` is nil very often (`map[string]any{"x": nil}`), this plan adds the small
companion codec `typeid.OptionalDTMSerializer` (and its `wirefmt/typeid` twin)
and switches the interface templates to it. That is §4.0.

A verified prototype of §4 (hand-written in `$TMPDIR`, using the library's real
codecs) round-trips `nil`, defined structs, defined scalars and recursive `any`
values through the buffer and stream trees, including `map[string]any` and
`[]any`. Follow it.

## 2. Non-goals (state them in the docs, do not implement here)

1. **No runtime `any` codec.** The set of implementations is a compile-time
   property; generating a type switch (as for named interfaces) avoids `reflect`
   in the codec hot path. Hand-written code keeps doing what README §13.10 shows
   (read the DTM, switch, call the typed serializer).
2. **No support for a defined type whose underlying type is `any`**
   (`type Value any`). `BuildDefinedType` requires a defined *basic* type and that
   stays. Document it. (`type Values []any` **is** supported, because the
   underlying type is a container.)
3. **No support for anonymous non-empty interfaces** (`interface{ Foo() }`).
4. **No support for predeclared basic types as union implementations** in this
   change (`string`, `int`, `[]byte`, …). The union accepts what
   `RegisterInterface` already accepts: defined structs and defined types. A bare
   `"text"` stored in an `any` **panics** at encode time; wrap it in a defined
   type. This is a real limitation for JSON-like maps; §13 sketches the follow-up
   (`intropts.WithPrimitiveImpl`) so it can be added without redesigning anything.
5. **No empty interface as a map key** (`map[any]T`): rejected on purpose (a
   decoded key may hold an unhashable dynamic type and panic; see §4.5).
6. **No new fuzz target and no new CI matrix entry.** Extend the existing
   `FuzzDTM` and scanner seeds instead (§6).

## 3. Design summary

| Concern | Decision |
|---|---|
| Name of the empty interface | Canonical spelling `any` (reflect prints `interface {}`; normalize it). Complete name = `"any"`, full name = `"any"`, relative name = `"any"`. |
| Recognition | New `classifier.AnyType(t)` = `t == reflect.TypeFor[any]()`. Reliable: a *defined* empty interface (e.g. `type Empty interface{}`) is a different `reflect.Type` and keeps the existing named-interface path. |
| Scanner | New `scanner.KindAny` + `ParseAnyType`, accepting `any`, `interface{}`, `interface {}` and leading stars (`*any`). |
| Registration | `g.RegisterInterface(reflect.TypeFor[any](), ...)` / `g.AddInterface(reflect.TypeFor[any](), ...)`. No new generator method. |
| Generated symbols | Serializer var `anyBinco`, serializer type `anyBincoImpl` (produced by the existing `SerializerFns.Type` unexported-name rule). Rename with `genopts.WithSerializer(reflect.TypeFor[any](), "Any")` → `AnyBinco`. |
| Codec body | Reuse `serializer_interface.tmpl` + `interface_{marshal,unmarshal,size,skip}.tmpl`. The type spelling is `any`, which is valid Go. |
| Nil identifier | Add `typeid.OptionalDTMSerializer` / `wirefmt/typeid.OptionalDTMSerializer`: same bytes as `DTMSerializer` but `0` decodes as the reserved nil marker. The interface templates read through it (unchanged wire format). |
| Anonymous codecs | `any` has **no** anonymous serializer: `Serializer.Of("any", …)` falls through to `Serializer.Var("any")`. `map[string]any` / `[]any` / `*any` still get their ordinary hashed container codecs, with `anyBinco` as the element codec. |
| Ordering | `RegisterInterface(any, …)` must run **before** `AddStruct`/`AddDefinedType` for a type that holds `any`; otherwise those methods fail with the new `codegen.ErrNoAnySerializer`. `RegisterInterface(any, …)` sets the flag before it adds its own implementations, so an implementation may itself contain `any` (recursive unions work). |
| Map keys | `map[any]T` (and any key name that contains `any`, e.g. `map[[2]any]int`) is rejected with a clear error. |

## 4. File-by-file implementation

Work in this order; each step leaves the tree compiling. Use the identifiers
exactly as given — the docs and tests below quote them.

### 4.0 `typeid`: read the reserved zero identifier (do this first)

**Why.** `codegen/templates/interface_unmarshal.tmpl` and
`interface_skip.tmpl` call `typeid.DTMSerializer.Unmarshal`, which returns
`binco.ErrOverflow` for identifier `0` (`typeid/meta.go` has `if num <= 0`). So
`case 0:` (nil) never runs. `binco.Encode(UniversalBinco, nil)` writes `[0]` and
`binco.Decode(UniversalBinco, [0])` fails — a pre-existing, untested bug that the
new `any` codec would inherit.

**`typeid/meta.go`** — add next to `DTMSerializer`:

```go
// OptionalDTMSerializer reads and writes the identifier an interface payload
// carries.
//
// It is DTMSerializer with one difference: identifier zero is reserved for a nil
// interface value, so it is written and decoded instead of being rejected. Every
// non-zero identifier behaves exactly like DTMSerializer and produces the same
// bytes.
var OptionalDTMSerializer = optionalMetaCodec{}

type optionalMetaCodec struct{}

func (s optionalMetaCodec) Marshal(dtm binco.DTM, bs []byte) (n int) {
	if dtm < 0 {
		panic(binco.ErrOverflow)
	}
	return varnum.PositiveInt64.Marshal(int64(dtm), bs)
}

func (s optionalMetaCodec) Unmarshal(bs []byte) (dtm binco.DTM, n int, err error) {
	num, n, err := varnum.PositiveInt64.Unmarshal(bs)
	if err != nil {
		return
	}
	if int64(binco.DTM(num)) != num {
		return 0, n, binco.ErrOverflow
	}
	return binco.DTM(num), n, nil
}

func (s optionalMetaCodec) Size(dtm binco.DTM) (size int) {
	if dtm < 0 {
		panic(binco.ErrOverflow)
	}
	return varnum.PositiveInt64.Size(int64(dtm))
}

// Skip validates exactly like Unmarshal, so the two agree on 32-bit hosts where
// an identifier above MaxInt does not fit binco.DTM.
func (s optionalMetaCodec) Skip(bs []byte) (n int, err error) {
	_, n, err = s.Unmarshal(bs)
	return
}
```

**`wirefmt/typeid/meta.go`** — the streaming twin (same doc, stream signatures):

```go
var OptionalDTMSerializer = optionalMetaCodec{}

type optionalMetaCodec struct{}

func (optionalMetaCodec) Marshal(dtm binco.DTM, w wirefmt.Writer) (n int, err error) {
	if dtm < 0 {
		return 0, binco.ErrOverflow
	}
	return wvarnum.PositiveInt64.Marshal(int64(dtm), w)
}

func (optionalMetaCodec) Unmarshal(r wirefmt.Reader) (dtm binco.DTM, n int, err error) {
	num, n, err := wvarnum.PositiveInt64.Unmarshal(r)
	if err != nil {
		return
	}
	if int64(binco.DTM(num)) != num {
		return 0, n, binco.ErrOverflow
	}
	return binco.DTM(num), n, nil
}

func (optionalMetaCodec) Size(dtm binco.DTM) int {
	if dtm < 0 {
		panic(binco.ErrOverflow)
	}
	return wvarnum.PositiveInt64.Size(int64(dtm))
}

func (optionalMetaCodec) Skip(r wirefmt.Reader) (n int, err error) {
	_, n, err = (optionalMetaCodec{}).Unmarshal(r)
	return
}
```

**Also fix `DTMSerializer.Skip` in both trees.** It currently delegates to
`varnum.PositiveInt64.Skip`, which does not apply the identifier range check, so
on a 32-bit host `Skip` accepts an id that `Unmarshal` rejects — `AGENTS.md`
§4.1 requires them to agree. Reuse the codec's own reader:

```go
// buffer, typeid/meta.go
func (s metaCodec) Skip(bs []byte) (n int, err error) {
	_, n, err = s.Unmarshal(bs)
	return
}

// stream, wirefmt/typeid/meta.go
func (s metaCodec) Skip(r wirefmt.Reader) (n int, err error) {
	_, n, err = s.Unmarshal(r)
	return
}
```

**`codegen/templates/interface_unmarshal.tmpl`** and
**`codegen/templates/interface_skip.tmpl`**: change only the codec name on the
identifier read:

```
dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal({{ $fulp }})
```

Leave `versioned_unmarshal.tmpl` / `versioned_skip.tmpl` on `typeid.DTMSerializer`:
a version identifier is always positive and has no nil marker.

**`codegen/templates.go`** — make the nil *write* go through the same codec, so
the reader and the writer cannot drift and the short-buffer guarantee comes from
`varnum` (which panics with `binco.ErrBufferTooSmall`) instead of a raw
`bs[0] = 0`:

```go
// NilInterfaceMarshal is the statement that writes the reserved zero type
// identifier, which stands for a nil interface value.
func (f TmplFns) NilInterfaceMarshal(opts genopts.Options) string {
	return "return typeid.OptionalDTMSerializer.Marshal(0, " +
		f.MarshalLastParam(true, opts) + ")"
}
```

`NilInterfaceSize` stays as it is (`"1"`). `interface_marshal.tmpl` needs no
change: it already calls the helper inside `if v == nil { ... }`.

### 4.1 `codegen/classifier/kind.go` — recognise the predeclared empty interface

Add after `DefinedInterface`:

```go
// AnyType reports whether t is the predeclared empty interface, spelled any or
// interface{}.
//
// A defined empty interface, such as type Empty interface{}, is not AnyType: it
// has its own package and name, so it takes the ordinary interface path.
func AnyType(t reflect.Type) (ok bool) {
	return t == reflect.TypeFor[any]()
}
```

(`reflect` is already imported.)

### 4.2 `codegen/typename/names.go` — name it `any`

In `typeCompleteName`, add a case **before** the default (keep
`DefinedType`, `PrimitiveType`, `Array`, `Slice`, `Map`, then `AnyType`, so a
defined type still wins):

```go
case classifier.AnyType(t):
	return CompleteName(stars + "any"), nil
```

Do **not** touch `BaseTypeCompleteName`: `type X any` stays unsupported, and
`BaseTypeCompleteName(any)` returning `ErrTypeMismatch` is the documented
behaviour (add a test for it).

### 4.3 `codegen/scanner` — parse the name

`codegen/scanner/types.go`: append `KindAny` to the `Kind` const block (after
`KindPrim`; nothing depends on the numeric values).

`codegen/scanner/parsers.go`: add next to the other parsers:

```go
// ParseAnyType parses the predeclared empty interface: any, interface{} or
// interface {} (the spelling reflect uses), with optional leading stars.
//
// The canonical spelling this module emits is "any".
func ParseAnyType[T QualifiedName](name T) (t TypeInfo[T], ok bool) {
	str := string(name)
	stars := str[:len(str)-len(strings.TrimLeft(str, "*"))]
	switch str[len(stars):] {
	case "any", "interface{}", "interface {}":
		return TypeInfo[T]{
			Stars: stars,
			Name:  typename.TypeName("any"),
			Kind:  KindAny,
		}, true
	default:
		return TypeInfo[T]{}, false
	}
}
```

(`strings` is already imported.)

`codegen/scanner/scanner.go`: in `scan`, insert **after** `ParseContainerType`
and **before** `ParsePrimitiveType`:

```go
if t, ok := ParseAnyType(name); ok {
	return s.scanAny(t, position)
}
```

and add:

```go
func (s Scanner[T]) scanAny(t TypeInfo[T], position Position) (err error) {
	t.Position = position
	return s.op.ProcessType(t, s.tops)
}
```

`ParseAnyType` is an exact match, so it does not shadow `pkg.any` (handled by
`ParseDefinedType` first) or `anything` (still unsupported).

### 4.4 `codegen/spec/converter/converter.go` — write `any`

Both `fullNameOp.ProcessType` and `relativeNameOp.ProcessType` switch on
`t.Kind`; add to **both**:

```go
case scanner.KindAny:
	o.b.WriteString("any")
```

`t.Stars` is already written before the switch, so `*any` becomes `*any`.

### 4.5 `codegen/spec/builder` — accept `any`, probe for it, reject map keys

**`builder.go`, `BuildInterfaceType`**: replace the opening with

```go
if !classifier.DefinedInterface(t) && !classifier.AnyType(t) {
	err = b.notInterfaceError(t)
	return
}
if classifier.AnyType(t) && len(iops.Impls) == 0 {
	err = NewAnyWithoutImplsError()
	return
}
```

Everything else already works once §4.2/§4.3 land: `parseTypeName(any)` yields
`typename.FullName("any")`.

Add next to `CollectAnonTypes`:

```go
// UsesAny reports whether the type name mentions the empty interface.
//
// A type that holds any needs the interface serializer declared by
// RegisterInterface(reflect.TypeFor[any](), ...). A name the scanner cannot
// parse reports false; the caller's own scan reports that error.
func (b TypeBuilder) UsesAny(name typename.FullName) bool {
	return usesAnyName(name)
}
```

**`anon.go`**: add the probe and the map-key guard:

```go
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
```

In `CollectAnonOp.makeMapType`, first statement:

```go
if usesAnyName(t.KeyType) {
	return d, false, NewAnyMapKeyError()
}
```

Leave the `switch typeInfo.Kind` in `CollectAnonOp.ProcessType` untouched:
`KindAny` has no anonymous serializer (no case = no anon), which is what makes
`Serializer.Of("any", …)` fall through to `Serializer.Var("any")`. Add a comment
above the switch:

```go
// KindAny has no case on purpose: the empty interface is encoded by the
// generated union (anyBinco), not by an anonymous container serializer.
```

**`errors.go`**: add two constructors in the existing style:

```go
// NewAnyWithoutImplsError reports an interface{} serializer without implementations.
func NewAnyWithoutImplsError() error {
	return fmt.Errorf("an interface{} serializer needs at least one implementation; add one with intropts.WithStructImpl or intropts.WithDefinedTypeImpl")
}

// NewAnyMapKeyError reports an interface{} map key.
//
// A decoded key would hold a value of one of the registered types, and a type
// that is not comparable would panic inside the map assignment, so the shape is
// rejected while the type is generated instead of on hostile input.
func NewAnyMapKeyError() error {
	return fmt.Errorf("an interface{} cannot be a map key; a decoded key may hold a dynamic type that is not comparable, so use a concrete key type")
}
```

### 4.6 `codegen/generator.go` — the registration flag and the two checks

1. Add a field to `Generator`:

```go
// anyRegistered is set once the empty interface has a generated serializer.
// AddStruct and AddDefinedType refuse a type that holds any until then.
anyRegistered bool
```

2. `AddInterface`: after `BuildInterfaceType` succeeds and before queueing:

```go
if classifier.AnyType(t) {
	g.anyRegistered = true
}
```

3. `RegisterInterface`: after `ApplyRegister` and **before** the impl loops (an
   implementation may itself contain `any`, so the flag must be set first), and
   validate the obvious misuse up front so the flag is not set on a failing call:

```go
if classifier.AnyType(t) {
	if len(rops.StructImpls)+len(rops.DefinedTypeImpls) == 0 {
		return bldr.NewAnyWithoutImplsError()
	}
	g.anyRegistered = true
}
```

(`bldr` is already imported.)

4. `AddStruct`: inside the `for _, fd := range structType.SerializedFields()`
   loop, **after** the existing `CollectAnonTypes` call (so the more specific
   `map[any]T` error wins):

```go
if !g.anyRegistered && g.builder.UsesAny(fd.FullName) {
	return fmt.Errorf("%w: field %s of %s; call RegisterInterface(reflect.TypeFor[any](), ...) first",
		ErrNoAnySerializer, fd.FieldName, structType.FullName)
}
```

Only serialized fields are checked, so an ignored `any` field needs no
registration.

5. `AddDefinedType`: after the existing `CollectAnonTypes` call:

```go
if !g.anyRegistered && g.builder.UsesAny(definedType.UnderlyingTypeName) {
	return fmt.Errorf("%w: the underlying type of %s is %s; call RegisterInterface(reflect.TypeFor[any](), ...) first",
		ErrNoAnySerializer, definedType.FullName, definedType.UnderlyingTypeName)
}
```

6. `codegen/errors.go`: add the sentinel next to the others:

```go
// ErrNoAnySerializer reports a type that holds the empty interface before the
// interface serializer was registered.
//
// Register the implementations with RegisterInterface(reflect.TypeFor[any](),
// ...) before adding the types that hold them.
var ErrNoAnySerializer = errors.New("no serializer for the empty interface")
```

### 4.7 `codegen/config/gen/options_gen.go` — do not silently accept `WithDTM(any, …)`

Before this change `WithDTM(reflect.TypeFor[any](), 1)` failed with *"can't get a
name for the interface {} type"*. After §4.2 it would resolve to `"any"` and be
silently ignored (the union has no DTM of its own). Keep a clear error:

```go
func WithDTM(t reflect.Type, dtm binco.DTM) SetOption {
	return func(o *Options) (err error) {
		if classifier.AnyType(t) {
			return NewUnsupportedDTMTypeError(t)
		}
		if dtm <= 0 {
			return binco.ErrOverflow
		}
		...
```

Add the import `"github.com/malivvan/binco/codegen/classifier"` (no cycle:
`classifier` imports only `reflect`) and in `errs.go`:

```go
// NewUnsupportedDTMTypeError reports a type that cannot carry a type identifier.
func NewUnsupportedDTMTypeError(t reflect.Type) error {
	return fmt.Errorf("cannot pin a type identifier for %v: pin the identifiers of its implementations instead", t)
}
```

(`errs.go` must import `reflect`.)

### 4.8 Templates — no other change needed

`serializer_interface.tmpl` and the four `interface_*.tmpl` includes already
produce valid code when `.FullName` is `"any"` and the identifier is read through
`OptionalDTMSerializer`. Expected output for
`RegisterInterface(any, WithStructImpl(AnyRecord), WithDefinedTypeImpl(AnyNumber))`:

```go
var anyBinco = anyBincoImpl{}
type anyBincoImpl struct{}

func (s anyBincoImpl) Marshal(v any, bs []byte) (n int) {
	if v == nil {
		return typeid.OptionalDTMSerializer.Marshal(0, bs)
	}
	switch t := v.(type) {
	case AnyRecord:
		return AnyRecordTypedBinco.Marshal(t, bs)
	case AnyNumber:
		return AnyNumberTypedBinco.Marshal(t, bs)
	default:
		panic(fmt.Sprintf(binco.ErrorPrefix+"unexpected %v type", t))
	}
}

func (s anyBincoImpl) Unmarshal(bs []byte) (v any, n int, err error) {
	dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case 0:
		// The writer stored a nil interface value.
		return nil, n, nil
	case AnyRecordDTM:
		v, n1, err = AnyRecordTypedBinco.UnmarshalData(bs[n:])
	case AnyNumberDTM:
		v, n1, err = AnyNumberTypedBinco.UnmarshalData(bs[n:])
	default:
		err = binco.NewUnexpectedDTMError(dtm)
		return
	}
	n += n1
	return
}
```

`Size`/`Skip` follow the same shape. Stream mode uses
`wirefmt.Writer`/`wirefmt.Reader`, `wirefmt/typeid`, and
`return typeid.OptionalDTMSerializer.Marshal(0, w)`. `fmt`, `reflect`, `typeid`,
`varnum` and `binco` are all already in `templates/imports.tmpl`;
`internal/importfix` prunes what the output does not use.

## 5. Fixtures (`testdata/model.go`)

Append (this file is copied verbatim into the temporary module of
`TestGeneratedCodeCompilesAndRuns`, so keep it dependency-free beyond the imports
it already has):

```go
// AnyEvent holds a value whose concrete type is chosen at run time.
type AnyEvent struct {
	Kind    string
	Payload any
	Meta    map[string]any
	Items   []any
}

// AnyNumber and AnyRecord are implementations the tests register for any.
type AnyNumber int

type AnyRecord struct {
	Name  string
	Count int
}

// AnyNode nests any inside an implementation, so the union has to exist before
// its own implementations are added.
type AnyNode struct {
	Value int
	Next  any
}

// AnyList is a defined slice of any.
type AnyList []any

// AnyKeyedMap is rejected by the generator: a decoded interface key may not be
// comparable.
type AnyKeyedMap struct {
	M map[any]int
}
```

Do not modify existing fixtures (`FullStruct` etc.); their committed schema
fingerprint and golden bytes must not change.

## 6. Tests

### 6.0 `typeid` / `wirefmt/typeid` — the optional identifier

New `typeid/optional_test.go`:

* `TestOptionalDTMSerializerNilMarker` — `Marshal(0)` is `[0]`, `Size(0)` is 1,
  `Unmarshal([0])` is `(0, 1, nil)`, `Skip([0])` is `(1, nil)`.
* `TestOptionalDTMSerializerMatchesDTMSerializer` — for `1, 7, 127, 128, 300,
  1<<20`: the bytes, `Size` and `Skip` agree with `typeid.DTMSerializer`.
* `TestOptionalDTMSerializerRejectsBadInput` — `nil` → `ErrTruncated`;
  `[0x80, 0x00]` → `ErrNonCanonical`; when `strconv.IntSize == 32`, the varint of
  `MaxInt64` → `ErrOverflow` from **both** `Unmarshal` and `Skip`.
* `TestOptionalDTMSerializerRejectsNegative` — `Marshal(-1)` and `Size(-1)` panic
  with `binco.ErrOverflow`.
* Extend `typeid/coverage_test.go`:
  `DTMSerializer.Skip([]byte{0})` now returns `ErrOverflow` (the §4.0 fix), and
  the same for `wirefmt/typeid/coverage_test.go` (there: `[]byte{0}` through
  `bytes.NewReader`).

New `wirefmt/typeid/optional_test.go`: mirror the above with a reader/writer
(`wirefmt.SliceWriter` / `bytes.NewReader`); an empty reader is `io.EOF` for both
`Unmarshal` and `Skip`. Reuse the helpers the existing
`wirefmt/typeid/typed_test.go` uses.

`wirefmt/golden_test.go`: add a row next to `"typeid typed int32"` (and the
`typeid` / `wtypeid` imports):

```go
{
	name:         "optional dtm nil",
	buffer:       func() []byte { return binco.Encode(typeid.OptionalDTMSerializer, binco.DTM(0)) },
	stream:       func() []byte { return streamBytes(t, wtypeid.OptionalDTMSerializer, binco.DTM(0)) },
	decodeStream: func(bs []byte) error { return decodeStream(t, wtypeid.OptionalDTMSerializer, bs, binco.DTM(0)) },
	decodeBuffer: func(bs []byte) error { return decodeBuffer(t, typeid.OptionalDTMSerializer, bs, binco.DTM(0)) },
},
```

`wirefmt/safety_test.go`: add one `streamCase` with `lengthPrefixed: false`,
`name: "typeid optional dtm"`, encoding `binco.DTM(0)` (one byte) and decoding /
skipping through `wtypeid.OptionalDTMSerializer`. The generic checks (truncation,
stalled reader, skip agreement, no over-read) then cover it.

`typeid/fuzz_test.go`: extend `FuzzDTM` (same target — no CI matrix change) with
the optional codec: `n` within the input; when `err == nil`, `Skip` agrees and
`Size`/`Marshal`/`Unmarshal` round-trip the returned identifier, including `0`.

### 6.1 `codegen/classifier/kind_test.go`

`TestAnyType`: true for `reflect.TypeFor[any]()`; false for `int`,
`testdata.Universal`, `*any`, and a locally declared
`type definedEmpty interface{}`.

### 6.2 `codegen/typename/names_test.go`

* `TestTypeCompleteName`: add `{reflect.TypeFor[any](), "any"}` and
  `{reflect.TypeFor[*any](), "*any"}`.
* `TestBaseTypeCompleteName`: assert `BaseTypeCompleteName(any)` returns
  `typename.ErrTypeMismatch` (documents the `type X any` non-goal).

### 6.3 `codegen/scanner`

* `parsers_test.go`, new `TestParseAnyType`: `"any"` ok, `"*any"` ok with
  `Stars == "*"`, `"interface{}"` ok, `"interface {}"` ok, `"anything"` not ok,
  `"pkg.any"` not ok, `""` not ok. Every accepted case has
  `Name == typename.TypeName("any")` and `Kind == scanner.KindAny`.
* `scanner_test.go`: add rows to `TestScan`:
  `{"empty interface", "any", "type:any"}`,
  `{"pointer to empty interface", "*any", "type:any"}`,
  `{"map of empty interface", "map[string]any",
    "type:map[string]any [ type:string ] type:any"}`.
* `fuzz_test.go`, `addNameSeeds`: append `"any"`, `"*any"`, `"interface{}"`,
  `"map[string]any"`. Do **not** add a `FuzzParseAnyType` target.

### 6.4 `codegen/spec/converter/converter_test.go`

* `TestConvertToRelativeName`: add `{"any", "any"}`.
* The `ConvertToFullName` table(s): add `{"any", "any"}` and
  `{"map[string]any", "map[string]any"}`.

### 6.5 `codegen/spec/builder`

In `builder_test.go`:

* `TestBuildInterfaceType`: also build `reflect.TypeFor[any]()` with two impls and
  assert `FullName == "any"` and the impl list.
* New `TestBuildStructTypeWithAnyFields`: build `testdata.AnyEvent`; assert the
  field names are `"string"`, `"any"`, `"map[string]any"`, `"[]any"`.
* `TestCollectAnonTypes`: add `"any"` → 0 anon types; `"[]any"` → 1
  (`AnonKindSlice`, `ElemType == "any"`); `"map[string]any"` → 1
  (`AnonKindMap`, `ElemType == "any"`); `"*any"` → 1 (`AnonKindPtr`,
  `ElemType == "any"`).
* New `TestUsesAny`: true for `"any"`, `"*any"`, `"[]any"`,
  `"map[string]any"`, `"pkg.Pair[any, int]"`; false for `"int"`, `"[]string"`,
  `"pkg.Type"`, `"chan int"`. The map and generic cases exercise the probe's
  `ProcessLeftSquare`/`ProcessComma`/`ProcessRightSquare`.
* New `TestBuildInterfaceTypeAnyNeedsImpls`: `BuildInterfaceType(any,
  intropts.Options{})` errors and the message mentions `WithStructImpl`.
* New `TestAnyMapKeyIsRejected`: `BuildAnonType("map[any]int")` and
  `BuildAnonType("map[[2]any]int")` error, and so does `CollectAnonTypes` with
  `"map[any]int"`. (`BuildStructType` alone does *not* scan fields — the
  generator calls `CollectAnonTypes` — so the struct-level rejection is covered
  by `TestAnyMapKeyIsRejectedByAddStruct` in §6.8.)
* `errors_test.go` `TestErrorConstructors`: add the exact messages of
  `NewAnyWithoutImplsError` and `NewAnyMapKeyError`.

### 6.6 `codegen/names_test.go`

Add `TestSerializerFnsAny`:

* `Of("any", tpopts.Options{}, gops)` == `"anyBinco"`;
* `Var("any")` == `"anyBinco"`, `Type("any")` == `"anyBincoImpl"`;
* with `genopts.WithSerializer(reflect.TypeFor[any](), "Any")`,
  `Var("any")` == `"AnyBinco"` and `Type("any")` == `"anyBinco"`.

### 6.7 `codegen/config/gen/options_test.go`

New `TestWithDTMRejectsAny`: `WithDTM(reflect.TypeFor[any](), 1)` returns an error
that mentions "implementations". Also add to
`codegen/generator_coverage_test.go`'s `TestNewGeneratorRejectsUnnamedTypes`:

```go
if _, err := codegen.NewGenerator(genopts.WithPath(modelPkgPath),
	genopts.WithDTM(reflect.TypeFor[any](), 1)); err == nil {
	t.Fatal("NewGenerator accepted a pinned identifier for the empty interface")
}
```

### 6.8 `codegen/any_test.go` — generator-level tests

Common helper in the file:

```go
func registerAny(t *testing.T, g *codegen.Generator) {
	t.Helper()
	if err := g.RegisterInterface(reflect.TypeFor[any](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.AnyRecord]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.AnyNumber]()),
	); err != nil {
		t.Fatalf("RegisterInterface(any): %v", err)
	}
}
```

Tests:

1. `TestGenerateAnyInterface` — buffer mode; `registerAny`; `AddStruct(AnyEvent)`;
   `mustContain`:
   * `"var anyBinco = anyBincoImpl{}"`,
   * `"func (s anyBincoImpl) Marshal(v any, bs []byte) (n int)"`,
   * `"func (s anyBincoImpl) Unmarshal(bs []byte) (v any, n int, err error)"`,
   * `"return typeid.OptionalDTMSerializer.Marshal(0, bs)"`,
   * `"dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal(bs)"`,
   * `"case 0:\n\t\t// The writer stored a nil interface value.\n\t\treturn nil, n, nil"`,
   * `"case AnyRecordDTM:"`, `"case AnyNumberDTM:"`,
   * `"var AnyRecordTypedBinco = typeid.NewSerializer(AnyRecordDTM, AnyRecordBinco)"`,
   * `"anyBinco.Marshal(v.Payload, bs[n:])"`,
   * `"composite.NewMapSerializer[string, any](composite.String, anyBinco"`,
   * `"composite.NewSliceSerializer[any](anyBinco)"`.
   Do not assert the hashed anonymous variable names.
2. `TestGenerateAnyStructFieldOrder` — assert the four `AnyEventBinco.Marshal`
   lines appear in field order (`Kind`, `Payload`, `Meta`, `Items`) with
   `strings.Index`.
3. `TestAddStructAnyNeedsRegistration` — `AddStruct(AnyEvent)` returns
   `errors.Is(err, codegen.ErrNoAnySerializer)` and the message names `Payload`.
4. `TestAddDefinedTypeAnyNeedsRegistration` — `AddDefinedType(AnyList)` without
   registration is `ErrNoAnySerializer`; with `registerAny` it generates
   `"composite.NewSliceSerializer[any](anyBinco)"` and `"[]any(v)"`.
5. `TestRegisterAnyBeforeItsImpls` — `RegisterInterface(any,
   WithStructImpl(AnyNode))` (AnyNode has `Next any`) succeeds; the source
   contains `AnyNodeBinco` and `anyBinco`.
6. `TestRegisterAnyWithoutImpls` — `RegisterInterface(any)` errors; a following
   `AddStruct(AnyEvent)` still returns `ErrNoAnySerializer` (a failing call does
   not set the flag).
7. `TestGenerateAnyIgnoredField` — `AddStruct(AnyEvent, stopts.WithField(
   fldopts.WithIgnore()) × 4)` succeeds without registration; an ignored `any`
   field needs no union.
8. `TestGenerateStreamAny` — `newGenerator(t, genopts.WithStream())`;
   `registerAny`; `AddStruct(AnyEvent)`; `mustContain`:
   `"func (s anyBincoImpl) Marshal(v any, w wirefmt.Writer) (n int, err error)"`,
   `"func (s anyBincoImpl) Unmarshal(r wirefmt.Reader) (v any, n int, err error)"`,
   `"return typeid.OptionalDTMSerializer.Marshal(0, w)"`,
   `"anyBinco.Marshal(v.Payload, w)"`,
   `"\"github.com/malivvan/binco/wirefmt/typeid\""`.
9. `TestAnyMapKeyIsRejectedByAddStruct` — `AddStruct(AnyKeyedMap)` errors and the
   message mentions "map key".
10. `TestAnyDTMNumberingAndRegistry` (also fine in `codegen/dtm_test.go`) —
    `RegisterInterface(Universal, StructImpl, DefinedImpl)` then
    `RegisterInterface(any, AnyRecord, AnyNumber)`: ids `1..4` (struct impls
    first, then defined-type impls, per interface); `DTMRegistry()` has the four
    implementation names and **not** `"any"`.
11. `TestAnySharesAnImplementationIsRejected` — registering one impl in two
    interfaces still fails with `binco.ErrDuplicateDTM` (documents the existing
    one-union-per-implementation rule; no code change expected).

### 6.9 `codegen/nil_interface_test.go` — the nil fix at the template level

`TestNilInterfaceIsRepresentable` must expect the new writer and reader:

```go
"if v == nil {\n\t\treturn typeid.OptionalDTMSerializer.Marshal(0, bs)\n\t}",
"dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal(bs)",
"case 0:\n\t\t// The writer stored a nil interface value.\n\t\treturn nil, n, nil",
```

and keep the `Skip` fragment consistent (`case 0:\n\t\treturn n, nil`). Add a
`TestGenerateStreamNilInterface` (or extend the existing stream test) asserting
`return typeid.OptionalDTMSerializer.Marshal(0, w)`.

`codegen/templates_coverage_test.go`: `TestNilInterfaceSize` stays (`"1"`);
optionally add a unit test for the exact string `NilInterfaceMarshal` returns in
each mode, like the existing test for `NilInterfaceSize`.

### 6.10 `codegen/generator_test.go` — compile and run

Inside `TestGeneratedCodeCompilesAndRuns`, after the existing
`RegisterInterface(Universal, …)` entry of the `add` slice, add:

```go
func() error {
	return g.RegisterInterface(reflect.TypeFor[any](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.AnyRecord]()),
		intropts.WithStructImpl(reflect.TypeFor[testdata.AnyNode]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.AnyNumber]()))
},
func() error {
	return g.AddStruct(reflect.TypeFor[testdata.AnyEvent]())
},
```

Add `"errors"` to the program's imports and these blocks before the final
`fmt.Println("OK")`:

```go
	// A nil interface is the reserved identifier 0, for a named interface and
	// for any alike.
	if got := binco.Encode(testdata.UniversalBinco, nil); !bytes.Equal(got, []byte{0}) {
		fail("nil interface = %v, want [0]", got)
	}
	if v, err := binco.Decode(testdata.UniversalBinco, []byte{0}); err != nil || v != nil {
		fail("nil interface decode = %v %v", v, err)
	}

	// any: the DTM of the concrete type, then its payload.
	anyBS := binco.Encode(testdata.AnyBinco, testdata.AnyNumber(7))
	if !bytes.Equal(anyBS, binco.Encode(testdata.AnyNumberTypedBinco, testdata.AnyNumber(7))) {
		fail("any encoding differs from the typed serializer: %v", anyBS)
	}
	gotAny, err := binco.Decode(testdata.AnyBinco, anyBS)
	if err != nil {
		fail("any decode: %v", err)
	}
	if v, ok := gotAny.(testdata.AnyNumber); !ok || v != 7 {
		fail("any decode = %#v, want AnyNumber(7)", gotAny)
	}
	if n, err := testdata.AnyBinco.Skip(anyBS); err != nil || n != len(anyBS) {
		fail("any skip: %d %v", n, err)
	}
	if got := binco.Encode(testdata.AnyBinco, nil); !bytes.Equal(got, []byte{0}) {
		fail("nil any = %v, want [0]", got)
	}
	if v, err := binco.Decode(testdata.AnyBinco, []byte{0}); err != nil || v != nil {
		fail("nil any decode = %v %v", v, err)
	}
	if _, err := binco.Decode(testdata.AnyBinco, []byte{99, 0}); err == nil {
		fail("an unknown DTM decoded without an error")
	}
	func() {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok || !errors.Is(err, binco.ErrBufferTooSmall) {
				fail("nil any into an empty buffer panicked with %v", r)
			}
		}()
		testdata.AnyBinco.Marshal(nil, nil)
	}()
	func() {
		defer func() {
			if recover() == nil {
				fail("an unregistered type was encoded without a panic")
			}
		}()
		binco.Encode(testdata.AnyBinco, map[string]int{"a": 1})
	}()

	ev := testdata.AnyEvent{
		Kind:    "login",
		Payload: testdata.AnyRecord{Name: "a", Count: 2},
		Meta:    map[string]any{"n": testdata.AnyNode{Value: 1, Next: testdata.AnyNumber(3)}},
		Items:   []any{testdata.AnyNumber(4), nil, testdata.AnyRecord{Name: "b"}},
	}
	evBS := binco.Encode(testdata.AnyEventBinco, ev)
	gotEv, err := binco.Decode(testdata.AnyEventBinco, evBS)
	if err != nil {
		fail("event with any: %v", err)
	}
	if gotEv.Payload.(testdata.AnyRecord).Name != "a" {
		fail("event payload = %#v", gotEv.Payload)
	}
	if gotEv.Meta["n"].(testdata.AnyNode).Next.(testdata.AnyNumber) != 3 {
		fail("nested any = %#v", gotEv.Meta["n"])
	}
	if len(gotEv.Items) != 3 || gotEv.Items[0].(testdata.AnyNumber) != 4 || gotEv.Items[1] != nil {
		fail("items = %#v", gotEv.Items)
	}
	if !bytes.Equal(evBS, binco.Encode(testdata.AnyEventBinco, gotEv)) {
		fail("event with any is not re-encodable")
	}
```

Keep the existing `fmt.Sprint(bs) != "[216 4 2 104 105]"` assertion for
`SimpleStruct`: the new registrations add DTM constants and codecs but do not
change any existing struct's bytes.

## 7. Documentation (all in the same change)

`make doc-check` enforces links, symbols, numbering and the table of contents:

* Do **not** add a `##`/`###` heading (the TOC and numbering are checked). Use a
  `####` heading (doccheck ignores levels > 3) or a bold lead-in.
* Every `package.Symbol` mentioned in README/AGENTS must be exported. The new
  names exist after §4: `codegen.ErrNoAnySerializer`, `classifier.AnyType`,
  `builder.UsesAny`, `scanner.ParseAnyType`, `scanner.KindAny`,
  `genopts.NewUnsupportedDTMTypeError`, `typeid.OptionalDTMSerializer`.

### 7.1 README §7.6 "Interfaces" — `#### The empty interface`

Append after the nil-interface paragraph:

~~~~markdown
#### The empty interface

`reflect.TypeFor[any]()` registers a closed set of implementations for `any`.
Every `any` field, `[]any` and `map[string]any` in a type the generator renders
then writes the DTM of the value's concrete type followed by its payload; a nil
value is the reserved identifier `0`, so the bytes are the same as a named
interface (§13.10).

```go
type Login struct{ User string }
type Count int

type Event struct {
	Kind    string
	Payload any
	Meta    map[string]any
	Items   []any
}

g.RegisterInterface(reflect.TypeFor[any](),
	intropts.WithStructImpl(reflect.TypeFor[Login]()),
	intropts.WithDefinedTypeImpl(reflect.TypeFor[Count]()),
)
g.AddStruct(reflect.TypeFor[Event]()) // every any field uses anyBinco
```

* Register the union before adding the types that hold it. `AddStruct` and
  `AddDefinedType` reject a field of type `any` with
  `codegen.ErrNoAnySerializer` until the serializer exists, because the generated
  code refers to it.
* Implementations are what `RegisterInterface` already accepts: defined structs
  and defined types. A bare `string`, `int`, `map[string]any` or `[]byte` value
  cannot be stored in the union; wrap it in a defined type first. Encoding an
  unregistered type panics, like a named-interface codec does.
* The union is closed: an unknown identifier decodes to
  `binco.UnexpectedDTMError`.
* An empty interface cannot be a map key: a decoded key may hold a dynamic type
  that is not comparable, which would panic inside the map. Use a concrete key
  type, as in `map[string]any` above.
* Pin the identifiers with `genopts.WithDTM` before payloads outlive a build.
* The generated codec is the unexported `anyBinco` (`anyBincoImpl`).
  `genopts.WithSerializer(reflect.TypeFor[any](), "Any")` declares `AnyBinco`
  instead.
~~~~

### 7.2 README §7.3 "Generator methods"

Under **RegisterInterface**, add two sentences: the empty interface is registered
the same way (`reflect.TypeFor[any]()`), it must come before the structs that hold
`any`, and its implementations receive the same DTM constants and typed
serializers (so `Generator.DTMRegistry` includes them).

### 7.3 README §7.4 "Generated code and naming"

Add a sentence: the empty interface's serializer is `anyBinco` of type
`anyBincoImpl` (unexported, like an anonymous container codec); rename it with
`genopts.WithSerializer`.

### 7.4 README §8.7 "typeid"

Add to the code block and the prose:

```go
typeid.OptionalDTMSerializer                        // the interface identifier, 0 = nil
```

> `DTMSerializer` rejects a zero identifier because an identifier on the wire is
> positive. `OptionalDTMSerializer` has the same encoding and accepts zero, which
> is the reserved marker for a nil interface value; the generated interface codecs
> use it. Both agree on every non-zero identifier.

### 7.5 README §13.10 "Type identifiers"

Add:

```markdown
An empty-interface value has no identifier of its own: the writer stores the DTM
of the concrete type followed by that type's payload, and the reserved identifier
`0` means nil. The bytes are identical to a named interface, and a reader that
does not know the concrete type reports `binco.UnexpectedDTMError`. Interface
codecs read and write the identifier through `typeid.OptionalDTMSerializer`,
which accepts the zero marker; `typeid.DTMSerializer` keeps rejecting it.
```

### 7.6 README §10 "Error reference"

Add a row:

```markdown
| `codegen.ErrNoAnySerializer` | a struct or defined type holds `any` before its implementations were registered |
```

### 7.7 AGENTS.md

* §4.3 (codec families): add a paragraph — an `any` field has no runtime codec;
  the generator emits a DTM dispatcher over the implementations registered with
  `RegisterInterface(reflect.TypeFor[any](), ...)`, reusing the named-interface
  templates, with `0` for nil.
* §5 (errors): add `codegen.ErrNoAnySerializer`.
* §6 (wire format): add a bullet — an interface payload writes the concrete
  type's DTM followed by its payload; nil is the reserved id `0`; the identifier
  is read through `typeid.OptionalDTMSerializer`, so `0` decodes. No new bytes.
* §7 (the generator): add a sentence — `any` is a scanner kind
  (`scanner.KindAny`) with no anonymous serializer; its codec is the generated
  union, and `AddStruct`/`AddDefinedType` require the union first.
* §9 (testing): add "the empty interface: registration order, missing
  registration, map keys, nil identifiers, and both trees" to the list.
* §13 (traps): add three bullets:
  * **`any` registration order** — `RegisterInterface(reflect.TypeFor[any](), ...)`
    must run before `AddStruct`/`AddDefinedType` for a type that holds `any`.
  * **`any` map keys** — `map[any]T` is rejected because a decoded key may hold an
    unhashable dynamic type.
  * **One union per implementation** — a type registered in two interfaces
    declares two constants with one name and is rejected with
    `binco.ErrDuplicateDTM`; register each implementation once.

### 7.8 CHANGELOG.md

Under `## [Unreleased]`, add:

```markdown
### Added

- Struct fields of type `any` (and containers such as `[]any` and
  `map[string]any`) are generated. `RegisterInterface(reflect.TypeFor[any](),
  ...)` registers the closed set of implementations; the generated codec writes
  the concrete type's DTM followed by its payload, and nil is the reserved
  identifier `0`. A heterogeneous field no longer needs a hand-rolled `[]byte`
  envelope. A field of type `any` without a registered union is rejected with
  `codegen.ErrNoAnySerializer`, and `map[any]T` is rejected because a decoded key
  may not be comparable.
- `typeid.OptionalDTMSerializer` (and `wirefmt/typeid.OptionalDTMSerializer`):
  the interface identifier, where zero is the reserved nil marker.

### Fixed

- A nil interface value did not decode: the generated codecs read the reserved
  zero identifier with `typeid.DTMSerializer`, which rejects it. They now use
  `typeid.OptionalDTMSerializer`, so `Marshal(nil)`/`Unmarshal` round-trip, and a
  short output buffer panics with `binco.ErrBufferTooSmall` instead of a runtime
  index error.
- `typeid.DTMSerializer.Skip` applies the same range checks as its `Unmarshal`,
  so the two agree on 32-bit hosts and for the zero identifier.
```

Keep the existing `Unreleased` `### Changed` block.

## 8. NOCOVER.md — refresh the line numbers (do not skip)

> **Superseded after execution.** At the user's request the statement-coverage
> gate was removed from the repository: `NOCOVER.md` and
> `scripts/covercheck.sh` are deleted, `make cover-check` is gone, and CI only
> measures coverage for Coveralls. The steps below describe the original plan
> and no longer apply; `make cover` remains as a reporting target.

`NOCOVER.md` lists exact `file:line` keys, and `scripts/covercheck.sh` fails both
when a listed block is covered and when a listed block no longer exists. Editing
`generator.go` shifts those numbers. Procedure:

1. Run `make cover-check` (§9) and read the failure.
2. Cover every **newly** reported uncovered statement with a test from §6. The
   intended outcome is that the only uncovered blocks are the same ten as today
   (nine in `codegen/generator.go`, including the `registerTemplates` one, and
   one in `codegen/names.go`) — plus none of the new code in `typeid`,
   `wirefmt/typeid` or the builder: every branch there is reachable and must be
   tested (including the negative-`dtm` panic branches and the 32-bit overflow
   case, which runs under `GOARCH=386`).
3. Update the line numbers in both the table and the ```nocover``` block of
   `NOCOVER.md`. Keep the reasons verbatim; only numbers change.
4. Re-run `make cover-check` until it prints 100% with 10 documented exceptions.

If a genuinely new unreachable branch appears, prefer deleting the branch over
adding an entry, as `NOCOVER.md` itself requires.

## 9. Verification

Run from the repository root, in this order, and do not declare done until all
pass:

```bash
gofmt -l .                                   # must print nothing
go vet ./...
go test ./...                                # includes the compile-and-run test
go test -race ./...
GOARCH=386 go test ./...
make cover                                   # reports the module total (no gate)
make doc-check
make portable-quick
go test -tags=fuzz -run='^$' ./...           # fuzz targets still compile
```

Focused loop while developing:

```bash
go test ./typeid/... ./wirefmt/typeid/... ./codegen/... -run 'Optional|Any|NilInterface|DTM' -v
go test -short ./...                         # skips the temp-module compile test
```

Then review the diff:

```bash
git status --short            # no stray files: coverage.out, *.test, temp modules
git diff --stat
git diff                      # only the intended changes
```

`coverage.out` is generated and ignored; delete it if it is left behind.

## 10. Traps and common mistakes

1. **Do not skip §4.0.** Without `OptionalDTMSerializer`, every nil `any` value
   (and every nil named interface) fails to decode with `ErrOverflow`, and the
   `case 0:` branches stay dead. The wire bytes do not change.
2. **Do not add a runtime `any` codec.** Any change that dispatches on
   `reflect.Type` at run time, or a new package, is a design error: the closed
   implementation set only exists at generation time.
3. **Do not renumber `Kind` blindly.** Append `KindAny` after `KindPrim`.
4. **Do not add `any` to `PrimitivePattern`.** That regex is unanchored and would
   accept prefixes; use the dedicated exact-match `ParseAnyType`.
5. **Do not special-case `any` in `codecName`/`Var`.** The existing foreign-type
   path already yields `anyBinco` / `anyBincoImpl`; a special case would create a
   collision with a user type named `Any`.
6. **Order of the `RegisterInterface` impl slices.** Structs are added before
   defined types regardless of the order they were passed in, so DTM numbering
   tests must list struct impls first (`AnyRecord` before `AnyNumber`).
7. **`UsesAny` runs on names, not `reflect.Type`s.** Call it with
   `fd.FullName` / `definedType.UnderlyingTypeName`, after `CollectAnonTypes`, so a
   `map[any]T` field reports the map-key error instead of the registration error.
8. **Ignored fields.** Only `SerializedFields()` are checked; an ignored `any`
   field must not require the union.
9. **Do not change existing golden bytes or fingerprints.** `SimpleStruct`'s
   committed fingerprint, README §13 tables and the existing
   `wirefmt/golden_test.go` rows are untouched: the `any` encoding is the
   named-interface encoding.
10. **Stream and buffer must both be generated from the same templates.** Keep
    `Tmpl.NilInterfaceMarshal` as the only place that spells the nil write, and
    `OptionalDTMSerializer` as the only identifier reader for interfaces.
11. **`versioned_*.tmpl` stay on `DTMSerializer`.** A version identifier is always
    positive and has no nil marker; switching it would weaken the check.
12. **~~`NOCOVER.md` line numbers drift~~** — not applicable: the coverage gate
    and `NOCOVER.md` were removed after execution (§8).
13. **Do not add a README `##`/`###` heading**; use `####`.
14. **A plain `string` inside `any` panics.** That is the documented closed-union
    rule in §2 (item 4), not a bug to paper over in this change; the panic names
    the dynamic type.

## 11. Acceptance checklist

- [x] `binco.Decode(UniversalBinco, []byte{0})` returns `nil, nil` and
      `binco.Encode(UniversalBinco, nil)` is `[0]` (named interface nil works).
- [x] `g.AddStruct` accepts `any`, `[]any`, `map[string]any`, `*any` fields and
      `type Values []any`, and generates a dispatcher that round-trips nil and
      every registered implementation.
- [x] `g.RegisterInterface(reflect.TypeFor[any](), ...)` works, assigns DTM
      constants and typed serializers to the implementations, and accepts an
      implementation that itself holds `any`.
- [x] Without the registration, `AddStruct`/`AddDefinedType` fail with
      `codegen.ErrNoAnySerializer`; `map[any]T` is rejected with a map-key error.
- [x] Generated nil values panic with `binco.ErrBufferTooSmall` on a short
      buffer.
- [x] Buffer and stream generation both work; an `any` value's bytes equal the
      typed serializer's bytes.
- [x] `typeid.OptionalDTMSerializer` and its streaming twin are tested (nil
      marker, parity with `DTMSerializer`, truncation, 32-bit overflow, skip
      agreement) and covered by the golden parity row and the safety row.
- [x] `gofmt`, `go vet`, `go test`, `go test -race`, `GOARCH=386 go test`,
      `make doc-check`, `make portable-quick` and the fuzz compilation all pass.
- [x] README (§7.3, §7.4, §7.6, §8.7, §10, §13.10), AGENTS.md (§4.3, §5, §6, §7,
      §9, §13) and CHANGELOG.md describe the behaviour; `NOCOVER.md` line numbers
      are current.

## 12. Suggested commit order

1. `fix(typeid): decode the reserved nil identifier` — §4.0 (both trees,
   templates, `TmplFns`), tests §6.0, docs §7.4–§7.6/§7.8.
2. `feat(codegen): name and parse the empty interface` — §4.1–§4.4 + tests
   §6.1–§6.4.
3. `feat(codegen): generate DTM dispatch for interface{} fields` — §4.5–§4.8,
   fixtures §5, tests §6.5–§6.9.
4. `test(codegen): compile and run any fields` — §6.10.
5. `docs: document generated any fields` — §7.

Each commit compiles and passes `go test ./codegen/... ./typeid/...`; only the
last one runs the full definition-of-done suite.

## 13. Follow-up (not part of this change): predeclared basic implementations

Today the union only accepts defined structs and defined types, so a JSON-like
`map[string]any{"name": "x"}` panics on the plain `string`. The extension is
self-contained if it is wanted later:

1. `intropts`: add `PrimitiveImpls []reflect.Type` to `RegisterOptions` and
   `WithPrimitiveImpl(t reflect.Type)`; validate `t.PkgPath() == ""`, a scalar
   kind, and a non-empty `t.Name()` (`string`, `bool`, `int*`, `uint*`, `float*`;
   `byte` is `uint8`; `[]byte`/containers have no name and stay out).
2. `RegisterInterface`: for a primitive impl, skip `AddDefinedType`/`AddTyped`;
   append `intropts.WithImpl(t)` (the union template only needs the name) and a
   DTM entry (`typeDTMEntry` works: `TypeCompleteName(string) == "string"`).
3. Queue the typed serializer with a new small template — or teach
   `serializer_typed.tmpl` a branch — that emits
   `var stringTypedBinco = typeid.NewSerializer(stringDTM, composite.String)`
   (the codec expression comes from `Serializer.Of`, so the generator mode is
   honoured). `Serializer.TypedVar("string")` already spells `stringTypedBinco`,
   and `case string:` assigns cleanly into `any`.
4. Tests must cover the DTM constant name (`stringDTM`, unexported), both trees,
   a struct field holding a `string`, and the existing duplicate-name rejection.

The plan in §4 leaves room for this: the union is just a name list plus a generated
switch, so adding primitive cases does not change the wire format or the empty
interface's own encoding.

## 14. Execution log

Implemented on 2026-11-22 in this order:

| Step | Commit | Content |
|---|---|---|
| §4.0 | `d19cc09` | `typeid.OptionalDTMSerializer` + the `wirefmt/typeid` twin, `DTMSerializer.Skip` range checks, interface templates and `NilInterfaceMarshal`, unit/golden/safety/fuzz tests |
| §4.1–§4.4 | `5fa103f` | `classifier.AnyType`, the canonical name `any`, `scanner.KindAny`/`ParseAnyType`, converter cases, tests and fuzz seeds |
| §4.5–§4.8 | `7e74a8a` | builder acceptance + `UsesAny` + map-key guard, the generator flag and `ErrNoAnySerializer`, the `WithDTM(any)` rejection, fixtures, `any_test.go` |
| §6.10 | `ab090b4` | `TestGeneratedCodeCompilesAndRuns` gains the union, `AnyEvent`, nil interfaces and the `any` round trips |
| range check | `92f4beb` | the optional identifier is bounded like the plain one (see deviation 1) |
| §7, §8 | `c58d160` | README, AGENTS, CHANGELOG, `NOCOVER.md` (both removed again with the coverage gate) |

Deliberate deviations from the plan:

1. `OptionalDTMSerializer.Unmarshal` decodes through `varnum.Uint64` and checks
   `num > math.MaxInt64 || int64(binco.DTM(num)) != int64(num)` instead of
   delegating to `PositiveInt64`. The second clause can only fail on a 32-bit
   host, so combining it with the 64-bit-reachable `MaxInt64` test keeps the
   range check covered by `make cover-check` (which runs 64-bit). `Skip` reuses
   `Unmarshal`, so the two agree.
2. The nil interface value is written with
   `typeid.OptionalDTMSerializer.Marshal(0, …)` rather than a bare `bs[0] = 0`.
   That is what makes a short buffer panic with `binco.ErrBufferTooSmall` (the
   check lives in `varnum`) and keeps one codec as both writer and reader.
3. `TestGeneratedCodeCompilesAndRuns` registers the union with
   `genopts.WithSerializer(reflect.TypeFor[any](), "Any")`: the default codec is
   unexported and the generated program is a different package. The `any` values
   are passed to `binco.Encode[any]` explicitly, because Go cannot infer
   `T = any` from an untyped argument.
4. `NOCOVER.md` was refreshed to **nine** exceptions during the work (the
   `AddStruct` `CollectAnonTypes` branch became reachable through the map-key
   test). After that, at the user's request, the 100%-coverage gate itself was
   removed: `NOCOVER.md` and `scripts/covercheck.sh` are gone, `make cover-check`
   no longer exists, and `make cover` only reports the total.
5. `*any` is covered by the `AnyPointer` fixture and
   `TestGenerateAnyPointerField`.

The verification suite from §9 passes: `gofmt`, `go vet`, `go test ./...`,
`go test -race ./...`, `GOARCH=386 go test ./...`, `make cover`, `make doc-check`,
`make portable-quick`, and `go test -tags=fuzz -run='^$' ./...`. The last
coverage run before the gate was removed was 100% with nine documented
exceptions; it is reported, not enforced, from then on.
