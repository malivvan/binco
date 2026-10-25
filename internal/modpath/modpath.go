// Package modpath validates Go module paths and import paths.
//
// The code generator needs two checks: the import path of the package it writes
// into must be a valid module path, and an extra import must be a valid import
// path. This package implements exactly those two, following the rules of the
// Go toolchain: path elements are separated by slashes and may contain ASCII
// letters, digits and a small set of punctuation.
package modpath

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// kind selects the character set and the extra rules of a path.
type kind int

const (
	// moduleKind is used for module paths: the first element is a domain name
	// and the last one may carry a major-version suffix.
	moduleKind kind = iota
	// importKind is used for import paths: a superset of the module characters
	// (it also allows '+'), with no domain or version-suffix rules.
	importKind
)

// CheckImportPath reports whether path is a valid import path.
func CheckImportPath(path string) error {
	if err := checkPath(path, importKind); err != nil {
		return fmt.Errorf("invalid import path %q: %w", path, err)
	}
	return nil
}

// CheckPath reports whether path is a valid module path.
//
// A module path is an import path whose first element looks like a domain name
// (lower-case letters, digits, dots and dashes, at least one dot, no leading
// dash) and whose final element is not a malformed major-version suffix such as
// "/v1" or "/v2.0".
func CheckPath(path string) error {
	if err := checkPath(path, moduleKind); err != nil {
		return fmt.Errorf("invalid module path %q: %w", path, err)
	}
	first := path
	if i := strings.IndexByte(first, '/'); i >= 0 {
		first = first[:i]
	}
	// checkPath has already rejected an empty first element and a leading dash,
	// so only the domain-name rule is left to check here.
	if !strings.Contains(first, ".") {
		return fmt.Errorf("invalid module path %q: missing dot in first path element", path)
	}
	for _, r := range first {
		if !firstElementOK(r) {
			return fmt.Errorf("invalid module path %q: invalid character %q in first path element", path, r)
		}
	}
	if !validVersionSuffix(path) {
		return fmt.Errorf("invalid module path %q: invalid version suffix", path)
	}
	return nil
}

// checkPath applies the rules shared by both kinds: valid UTF-8, no empty
// elements, no leading or trailing slash and no double slash.
func checkPath(path string, k kind) error {
	if !utf8.ValidString(path) {
		return fmt.Errorf("invalid UTF-8")
	}
	if path == "" {
		return fmt.Errorf("empty path")
	}
	if path[0] == '-' {
		return fmt.Errorf("leading dash")
	}
	if strings.Contains(path, "//") {
		return fmt.Errorf("double slash")
	}
	if path[len(path)-1] == '/' {
		return fmt.Errorf("trailing slash")
	}
	start := 0
	for i, r := range path {
		if r != '/' {
			continue
		}
		if err := checkElem(path[start:i], k); err != nil {
			return err
		}
		start = i + 1
	}
	return checkElem(path[start:], k)
}

// checkElem applies the rules of one path element.
func checkElem(elem string, k kind) error {
	if elem == "" {
		return fmt.Errorf("empty path element")
	}
	if strings.Count(elem, ".") == len(elem) {
		return fmt.Errorf("invalid path element %q", elem)
	}
	if elem[0] == '.' && k == moduleKind {
		return fmt.Errorf("leading dot in path element")
	}
	if elem[len(elem)-1] == '.' {
		return fmt.Errorf("trailing dot in path element")
	}
	for _, r := range elem {
		if !elementRuneOK(r, k) {
			return fmt.Errorf("invalid character %q in path element %q", r, elem)
		}
	}
	// The prefix before the first dot must not be a reserved file name on
	// Windows, and the element must not look like a Windows short name
	// ("PROGRA~1"), so that a path stays usable as a file path there.
	short := elem
	if i := strings.IndexByte(short, '.'); i >= 0 {
		short = short[:i]
	}
	for _, reserved := range badWindowsNames {
		if strings.EqualFold(reserved, short) {
			return fmt.Errorf("%q is a reserved name on Windows", short)
		}
	}
	if tilde := strings.LastIndexByte(short, '~'); tilde >= 0 && tilde < len(short)-1 {
		if allDigits(short[tilde+1:]) {
			return fmt.Errorf("trailing tilde and digits in path element %q", elem)
		}
	}
	return nil
}

// validVersionSuffix reports whether the last element of a module path is not a
// malformed major-version suffix.
//
// A suffix of the form "/vN" is allowed for N >= 2 only: "/v1" is implicit in
// the path and "/v2.0" or "/v02" are ambiguous with a real package name.
// gopkg.in uses ".vN" instead and requires it on every path.
func validVersionSuffix(path string) bool {
	if strings.HasPrefix(path, "gopkg.in/") {
		return validGopkgInSuffix(path)
	}
	i := len(path)
	dotted := false
	for i > 0 && (isDigit(path[i-1]) || path[i-1] == '.') {
		if path[i-1] == '.' {
			dotted = true
		}
		i--
	}
	if i <= 1 || i == len(path) || path[i-1] != 'v' || path[i-2] != '/' {
		return true
	}
	major := path[i-2:]
	return !dotted && len(major) > 2 && major[2] != '0' && major != "/v1"
}

// validGopkgInSuffix applies the gopkg.in naming convention: every path ends in
// ".vN", with an optional "-unstable" suffix, and N is not zero-padded.
func validGopkgInSuffix(path string) bool {
	i := len(path)
	if strings.HasSuffix(path, "-unstable") {
		i -= len("-unstable")
	}
	for i > 0 && isDigit(path[i-1]) {
		i--
	}
	if i <= 1 || path[i-1] != 'v' || path[i-2] != '.' {
		return false
	}
	major := path[i-2:]
	return len(major) > 2 && (major[2] != '0' || major == ".v0")
}

// firstElementOK reports whether r may appear in the first element of a module
// path, which by convention is a domain name.
func firstElementOK(r rune) bool {
	return r == '-' || r == '.' || '0' <= r && r <= '9' || 'a' <= r && r <= 'z'
}

// elementRuneOK reports whether r may appear in a path element of the given
// kind.
func elementRuneOK(r rune, k kind) bool {
	if k == importKind && r == '+' {
		return true
	}
	return r == '-' || r == '.' || r == '_' || r == '~' ||
		'0' <= r && r <= '9' || 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z'
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isDigit(s[i]) {
			return false
		}
	}
	return s != ""
}

// badWindowsNames are the file names Windows reserves in every directory.
var badWindowsNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}
