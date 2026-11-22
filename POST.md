# After execution: what is left

The plan in `PLAN.md` is implemented and verified; see §14 of that file for the
commit map and the deliberate deviations. This file records what is **not** done
and what a maintainer has to decide.

## 1. The tag is local and the CHANGELOG is still `Unreleased`

The work ends with an annotated tag **`v0.11.0`** on the last commit (the previous
tags are `v0.9.0` and `v0.10.0`; there is no `v1.0.0` tag in this repository,
although `CHANGELOG.md` has a `[1.0.0]` section). Nothing was pushed.

Before pushing the tag — which makes CI's `Publish release` job create a GitHub
release — run the release checklist from `AGENTS.md` §12:

1. Move `## [Unreleased]` in `CHANGELOG.md` into `## [0.11.0] - YYYY-MM-DD`.
2. Review the `### Changed` entry that was already in the working tree before
   this change: *"Removed the standalone `minlz` compression package and its
   `xxhash` dependency."* It is **not** part of this change and does not match the
   current tree (`minlz/` and `xxhash/` still exist). Either that other change
   lands, or the line should be removed before the release notes are published.
3. Update the compare links at the bottom of `CHANGELOG.md` for the new tag.
4. Re-run `make portable` (the full matrix, not just `portable-quick`) and tag
   again if anything changed.

If the maintainer prefers to keep the repository at "1.0.0 is next", drop the
`v0.11.0` tag and cut `v1.0.0` once the `[1.0.0]` section is final.

## 2. Scope that was deliberately not implemented

These are documented behaviours, not defects:

1. **Predeclared basic values cannot go into an `any` union.** The union accepts
   what `RegisterInterface` already accepts: defined structs and defined types.
   `map[string]any{"name": "x"}` panics at encode time with the dynamic type
   named, because a plain `string` has no generated serializer/DTM. Workaround:
   wrap the scalar in a defined type (`type Text string`). `PLAN.md` §13 sketches
   `intropts.WithPrimitiveImpl` for the follow-up; it needs a typed-serializer
   path that references the library codec instead of a generated variable.
2. **`type Value any` is unsupported.** Only containers of `any` are accepted
   (`type Values []any` works). `BuildDefinedType` requires a defined basic type.
3. **Anonymous non-empty interfaces** (`interface{ Foo() }`) are still rejected:
   only the predeclared empty interface is recognised.
4. **`map[any]T` is rejected on purpose**: a decoded key could hold a dynamic type
   that is not comparable and panic inside the map assignment. `map[string]any`
   is the supported shape.
5. **One union per implementation.** A type registered in two interfaces declares
   two DTM constants with one name and fails with `binco.ErrDuplicateDTM`; this
   is the pre-existing rule, not new.
6. **No runtime `any` codec.** Dispatch is generated, so hand-written code still
   reads the DTM and switches (README §13.10). Adding one would need `reflect` in
   the codec hot path or a closed registry type.

## 3. Watch items

* The statement-coverage gate was removed at the user's request: `NOCOVER.md`
  and `scripts/covercheck.sh` are deleted, `make cover-check` is gone, and CI
  only measures coverage for Coveralls. `make cover` still reports the module
  total; a partial percentage is no longer a failure.
* The interface templates now read the identifier through
  `typeid.OptionalDTMSerializer`. If the reserved nil marker ever changes, that
  codec (both trees), `TmplFns.NilInterfaceMarshal`, README §13.10 and AGENTS §6
  have to change together.
* `TestGeneratedCodeCompilesAndRuns` exports the `any` codec with
  `genopts.WithSerializer(any, "Any")`. Without that option the default is the
  unexported `anyBinco`, which is intended: a generated file's consumers do not
  need to name the union codec.
* The scanner's `ParseAnyType` is an exact match on purpose. Do not fold `any`
  into `PrimitivePattern`: that regex is unanchored and would accept prefixes.
