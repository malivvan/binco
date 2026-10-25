# Validation

Validating data while it is unmarshalled.

- [string](string/): a string length validator.
- [slice](slice/): slice length and element validators.
- [map](map/): map length, key and value validators.

Validation is a property of the *reader*: `Marshal` never validates, and the
plain serializers accept whatever the validating ones reject. Every `NewValid*`
constructor takes the same options as its plain counterpart, so a validator can
be added without touching the rest of the code.

Validators run as early as possible:

| Validator | Runs after | Before |
|-----------|------------|--------|
| length | the length is decoded | any element is read |
| key | the key is decoded | the value is read |
| value / element | that element is decoded | the next element is read |

That keeps a hostile length from turning into a huge allocation, which is the
main reason to validate on untrusted input.

## string

```
"hello"        → accepted: "hello"
"aaaaaaaaaaa"  → rejected: string too long: 11 > 10 (errors.Is: true)
plain reader accepts 11 characters
ValidatorFn(3) = odd length
```

The validator can be any type implementing `binco.Validator[int]`, or a plain
function wrapped with `binco.ValidatorFn`. Returning a wrapped sentinel error
(`fmt.Errorf("%w: ...", ErrTooLong)`) keeps `errors.Is` working for the caller.

## slice

```
["hello", "world"]           → accepted: ["hello", "world"]
["hello", "world", "again"]  → rejected: too many elements: 3 > 2
["hello", ""]                → rejected: empty element
Skip: 19 bytes, no validators called
```

`Skip` checks the structure but does not decode elements, so no validator runs.

## map

```
{1:"one"}              → accepted: {1:"one"}
{1:"one", 2:"two"}     → rejected: too many entries: 2 > 1
{-1:"minus one"}       → rejected: negative key: -1
{1:""}                 → rejected: empty value
```
