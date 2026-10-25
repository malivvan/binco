// Package importfix prunes the import block of generated Go source and formats
// the result.
//
// The code generator emits a fixed superset of imports and then fills in the
// ones its types need. Unused entries have to disappear before the source is
// written, because Go rejects an unused import. This package does exactly that:
// it parses the file, keeps the imports whose package qualifier is used, drops
// the rest and runs gofmt over the result. It also removes duplicate imports of
// the same path, which happens when a caller passes an aliased import and the
// type-driven pass finds the same package.
package importfix

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Process prunes the unused imports of src and returns gofmt-formatted source.
func Process(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", src, parser.ParseComments)
	if err != nil {
		return src, err
	}
	specs := prune(collectImports(file), usedQualifiers(file))
	return format.Source(rewrite(fset, src, file, specs))
}

// importSpec is one import of the generated file.
type importSpec struct {
	// alias is the explicit import name, empty when the import is plain.
	alias string
	path  string
}

// collectImports returns the imports of file in source order.
func collectImports(file *ast.File) (specs []importSpec) {
	for _, spec := range file.Imports {
		// The parser rejected every literal it could not decode, so Unquote
		// cannot fail here.
		path, _ := strconv.Unquote(spec.Path.Value)
		var alias string
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		specs = append(specs, importSpec{alias: alias, path: path})
	}
	return specs
}

// usedQualifiers returns the identifiers that appear before a selector, which is
// how a generated file refers to an imported package ("varnum.Int").
//
// The check is syntactic: a local variable that happens to share a package name
// would keep an import that is not needed. Generated code does not do that, and
// erring towards keeping an import only costs a compile error that the caller
// sees immediately, while dropping a used one would be a silent break.
func usedQualifiers(file *ast.File) map[string]struct{} {
	used := make(map[string]struct{})
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if ident, ok := sel.X.(*ast.Ident); ok {
			used[ident.Name] = struct{}{}
		}
		return true
	})
	return used
}

// prune drops the imports whose qualifier is unused and the duplicate imports of
// one path.
func prune(specs []importSpec, used map[string]struct{}) (kept []importSpec) {
	byPath := make(map[string]int, len(specs))
	for _, spec := range specs {
		name := spec.alias
		if name == "" {
			name = path.Base(spec.path)
		}
		// A blank or dot import has side effects, so it is never unused.
		if name != "_" && name != "." {
			if _, ok := used[name]; !ok {
				continue
			}
		}
		if i, dup := byPath[spec.path]; dup {
			// Prefer the explicit alias: it is the one the generated code
			// refers to when the package name differs from the path.
			if kept[i].alias == "" && spec.alias != "" {
				kept[i] = spec
			}
			continue
		}
		byPath[spec.path] = len(kept)
		kept = append(kept, spec)
	}
	return kept
}

// rewrite replaces the import declaration of src with specs, or removes it when
// nothing is left. It returns src unchanged when the file has no import
// declaration and none is needed.
func rewrite(fset *token.FileSet, src []byte, file *ast.File, specs []importSpec) []byte {
	decl := importDecl(file)
	if decl == nil {
		// Every spec comes from an import declaration, so there is nothing to
		// rewrite and nothing to add.
		return src
	}
	start := fset.Position(decl.Pos()).Offset
	end := fset.Position(decl.End()).Offset
	replacement := ""
	if len(specs) > 0 {
		replacement = render(specs)
	}
	out := make([]byte, 0, len(src)+len(replacement))
	out = append(out, src[:start]...)
	out = append(out, replacement...)
	rest := src[end:]
	// Removing the declaration must not leave the blank lines it occupied.
	if replacement == "" {
		out = bytes.TrimRight(out, "\n")
		rest = bytes.TrimLeft(rest, "\n")
		out = append(out, '\n', '\n')
	}
	out = append(out, rest...)
	return out
}

// importDecl returns the import declaration of file, if it has one.
func importDecl(file *ast.File) *ast.GenDecl {
	for _, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
			return gen
		}
	}
	return nil
}

// render writes an import block with the standard library first and one entry
// per line, which is what gofmt would produce.
func render(specs []importSpec) string {
	sorted := make([]importSpec, len(specs))
	copy(sorted, specs)
	sort.SliceStable(sorted, func(i, j int) bool {
		si, sj := isStdlib(sorted[i].path), isStdlib(sorted[j].path)
		if si != sj {
			return si
		}
		return sorted[i].path < sorted[j].path
	})
	var b strings.Builder
	b.WriteString("import (\n")
	for i, spec := range sorted {
		if i > 0 && isStdlib(sorted[i-1].path) && !isStdlib(spec.path) {
			b.WriteByte('\n')
		}
		b.WriteByte('\t')
		if spec.alias != "" {
			b.WriteString(spec.alias)
			b.WriteByte(' ')
		}
		b.WriteString(strconv.Quote(spec.path))
		b.WriteByte('\n')
	}
	b.WriteString(")")
	return b.String()
}

// isStdlib reports whether an import path belongs to the standard library. The
// convention the Go toolchain uses is that the first element has no dot.
func isStdlib(importPath string) bool {
	first := importPath
	if i := strings.IndexByte(first, '/'); i >= 0 {
		first = first[:i]
	}
	return !strings.Contains(first, ".")
}
