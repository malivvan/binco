# Code generation with options

```
value:   {Num:1 Str:hello}
size:    14 bytes (Num alone takes 8: fixed width, not varint)
encoded: [1 0 0 0 0 0 0 0 5 104 101 108 108 111]
decoded: {Num:1 Str:hello}

string of 101 characters: string too long: 101 > 100
errors.Is(err, pkg.ErrTooLong): true
string of 100 characters: accepted, "aaaaaaaa"…
```

The same type as in [gen](../gen), but with three generator options:

| Option | Effect |
|--------|--------|
| `genopts.WithSerName(Foo, "CustomFoo")` | the serializer is named `CustomFooBinco` |
| `fldopts.WithType(tpopts.WithNumEncoding(tpopts.NumEncodingRaw))` | `Num` uses `scalar.Int` (8 bytes) instead of `varnum.Int` |
| `fldopts.WithType(tpopts.WithLenValidator("ValidateLength"))` | the generated code validates `Str`'s length while unmarshalling |

```go
g, _ := codegen.NewGenerator(
	genopts.WithPkgPath("github.com/malivvan/binco/demos/gen_options/pkg"),
	genopts.WithSerName(reflect.TypeFor[pkg.Foo](), "CustomFoo"),
)
g.AddStruct(reflect.TypeFor[pkg.Foo](),
	stopts.WithField(fldopts.WithType(
		tpopts.WithNumEncoding(tpopts.NumEncodingRaw))),
	stopts.WithField(fldopts.WithType(
		tpopts.WithLenValidator("ValidateLength"))),
)
```

The generated file shows the effect directly:

```go
var CustomFooBinco = customFooBinco{}

func (s customFooBinco) Marshal(v Foo, bs []byte) (n int) {
	n = scalar.Int.Marshal(v.Num, bs)            // fixed width, 8 bytes
	return n + stringkΔf...ΞΞ.Marshal(v.Str, bs[n:]) // validating serializer
}
```

The long, unreadable name is the generated anonymous serializer for the
validating string; it is created once per distinct type+options combination.

## Why this example exists

Per-field encoding and validation are usually chosen while writing the type, not
in a separate generator program. If you mark up your types (for example with
comments) or let an assistant derive the options from them, the generator calls
above are what the output must be equivalent to: the same serializers, the same
options, the same wire format.

## Notes

* Validators are referenced by name, so the function must be visible in the
  target package (`pkg.ValidateLength` here).
* `WithNumEncoding` applies to numbers only; unsupported options for a type are
  ignored.
* Regenerate after changing the options: `go generate ./demos/gen_options`.
