# Code generation

```
Int:     [216 4] (varint)
Foo:     [11 104 101 108 108 111 32 119 111 114 108 100 10] (13 bytes)
decoded: {S:hello world T:5} (read 13 bytes)
skipped: 13 bytes
through binco.Serializer[int]: size=2
```

Writing serializers by hand does not scale. `github.com/malivvan/binco/codegen`
inspects types through `reflect` and writes the same code you would write
yourself, for the buffer API or (with `WithStream`) for the streaming API.

## Layout

```
demos/gen/
  main.go                 // uses the generated serializers; holds go:generate
  pkg/types.go            // the types
  pkg/binco.gen.go         // generated, committed
  gen/main.go             // the generator program
```

## Run it

```bash
go generate ./demos/gen   # rewrites pkg/binco.gen.go
go run ./demos/gen        # uses it
```

`go generate` executes the directive at the top of `main.go`:

```go
//go:generate go run ./gen
```

## The generator program

```go
g, err := codegen.NewGenerator(
	genopts.WithPkgPath("github.com/malivvan/binco/demos/gen/pkg"),
)
if err != nil {
	log.Fatal(err)
}

if err = g.AddDefinedType(reflect.TypeFor[pkg.Int]()); err != nil { ... }
if err = g.AddStruct(reflect.TypeFor[pkg.Foo[pkg.Int]]()); err != nil { ... }

bs, err := g.Generate()
if err != nil {
	os.WriteFile("failed.gen.go", bs, 0o644) // inspect the output
	log.Fatal(err)
}
os.WriteFile("./pkg/binco.gen.go", bs, 0o644)
```

`WithPkgPath` is the only required option: it is the import path of the package
the generated file belongs to, and usually the `reflect.Type.PkgPath()` of one of
the types.

## What is generated

For `type Int int` and `type Foo[T any] struct { S string; T T }`:

```go
var IntBinco = intBinco{}
var FooBinco = fooBinco{}

func (s fooBinco) Marshal(v Foo[Int], bs []byte) (n int)
func (s fooBinco) Unmarshal(bs []byte) (v Foo[Int], n int, err error)
func (s fooBinco) Size(v Foo[Int]) (size int)
func (s fooBinco) Skip(bs []byte) (n int, err error)
```

Naming: the serializer variable is `<Type>Binco`, a DTM constant would be
`<Type>DTM`, and a typed serializer `<Type>TypedBinco`. Use `genopts.WithSerName`
to change it (see [gen_options](../gen_options)).

## Related

* [gen_options](../gen_options): per-field options, custom serializer names,
  validators.
* The [library README](../../README.md#code-generation) documents every
  generator method, mode and option.
