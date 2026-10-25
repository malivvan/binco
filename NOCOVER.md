# Coverage exceptions

`binco` aims for **100% statement coverage** of every library package.

Every package in the module is measured.

```bash
make cover-check   # go test -coverprofile over every package, then
                   # scripts/covercheck.sh fails unless everything is covered
```

`scripts/covercheck.sh` treats a statement listed in the `nocover` block below as
covered, and rejects an entry that no longer names an uncovered block, so the
list cannot rot unnoticed.

## Policy

A statement may only stay uncovered when covering it would require changing the
library in a way that makes it worse — for example a branch that cannot run on
any supported platform, or an error path that would need a test-only hook in
production code. Every such case is listed below with the reason.

An exception is a debt, not a default: the preferred fix is to delete the
unreachable branch, or to make the dependency injectable, rather than to add an
entry here, and most branches that used to be unreachable have been deleted or
made testable over time.

## Exceptions

| Statement | Why it cannot run |
|-----------|-------------------|
| `codegen/generator.go:258` | `RegisterVersioned` checks the error of `AddTyped` after `AddStruct`/`AddDefinedType` succeeded for the same type: `BuildTyped` accepts every type those accept, and parses the same already-converted name. Kept so a future builder change fails loudly instead of silently. |
| `codegen/generator.go:262` | `RegisterVersioned` checks the error of `typeDTMEntry` after the same type was added: its two conversions (the complete name and the full name) already succeeded in `AddStruct`/`AddTyped`. |
| `codegen/generator.go:296` | `RegisterInterface` checks the error of `AddTyped` for a struct implementation after `AddStruct` succeeded, for the same reason as `:258`. |
| `codegen/generator.go:301` | `RegisterInterface` checks the error of `typeDTMEntry` for a struct implementation after it was added, for the same reason as `:262`. |
| `codegen/generator.go:312` | `RegisterInterface` checks the error of `AddTyped` for a defined-type implementation after `AddDefinedType` succeeded, for the same reason as `:258`. |
| `codegen/generator.go:317` | `RegisterInterface` checks the error of `typeDTMEntry` for a defined-type implementation after it was added, for the same reason as `:262`. |
| `codegen/generator.go:408` | `AddStruct` checks the error of `CollectAnonTypes` for a field after `BuildStructType` validated the same options and converted the same field name. |
| `codegen/generator.go:430` | `AddDefinedType` checks the error of `CollectAnonTypes` for the underlying name after `BuildDefinedType` converted it. |
| `codegen/generator.go:456` | `NewGenerator` checks the error of `registerTemplates`, which parses the embedded template files: a failure is a build-time defect, not a runtime condition. |
| `codegen/names.go:285` | `Of` reports an "internal bug" for a type that has a codec-package rule but no serializer-name case. The rule tables (`portableRules`, `hybridRules`, `nativeRules`) contain exactly the handled keys, so no input reaches it. |

## The `nocover` block

```nocover
codegen/generator.go:258
codegen/generator.go:262
codegen/generator.go:296
codegen/generator.go:301
codegen/generator.go:312
codegen/generator.go:317
codegen/generator.go:408
codegen/generator.go:430
codegen/generator.go:456
codegen/names.go:285
```
