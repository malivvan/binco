#!/bin/sh
# portable.sh proves that binco builds on every target Go supports.
#
# usage: sh scripts/portable.sh [--quick]
#
# Stage 1 (full run only) cross-compiles every target from `go tool dist list`.
# Stage 2 compiles the test files on a representative cross-section, which is
# what catches test-only platform code (the `go test` binary is not built by
# `go build`).
#
# `--quick` skips the full matrix and the long test-compile list, for a fast
# local check. CI runs the full script and so should a release.
#
# /bin/true is what `go test -exec` runs instead of the test binary: the tests
# are compiled and linked, never executed. A POSIX shell is required.

set -eu

if ! command -v go >/dev/null 2>&1; then
	echo "portable: the go tool is not in PATH" >&2
	exit 1
fi

# The test files are compiled on these targets as well: both byte orders, every
# integer width, the desktop OS families and both wasm environments.
test_targets="
linux/386
linux/amd64
linux/arm
linux/arm64
linux/loong64
linux/mips
linux/mips64
linux/ppc64
linux/ppc64le
linux/riscv64
linux/s390x
darwin/amd64
darwin/arm64
windows/amd64
windows/arm64
freebsd/amd64
js/wasm
wasip1/wasm
plan9/amd64
aix/ppc64
"

quick_targets="
linux/386
linux/ppc64
windows/amd64
js/wasm
"

# run_for sets GOOS and GOARCH for one "os/arch" word and runs the rest.
run_for() {
	target=$1
	shift
	GOOS=${target%%/*} GOARCH=${target#*/} "$@"
}

# Android and iOS link the runtime through cgo, so the Go tool refuses to link a
# main package for them with CGO_ENABLED=0 ("requires external (cgo) linking").
# The library packages build unchanged; only the example main packages would need
# an NDK or Xcode. These are the only targets handled differently.
needs_toolchain() {
	case "$1" in
	android/*|ios/*) return 0 ;;
	*) return 1 ;;
	esac
}

quick=0
if [ "${1:-}" = "--quick" ]; then
	quick=1
fi

library_pkgs=$(go list ./... | grep -v '/demos')

if [ "$quick" -eq 0 ]; then
	count=0
	for target in $(go tool dist list); do
		count=$((count + 1))
		printf 'build %-18s ' "$target"
		if needs_toolchain "$target"; then
			if ! run_for "$target" env CGO_ENABLED=0 go build $library_pkgs; then
				echo "portable: build failed for $target" >&2
				exit 1
			fi
			echo "ok (library; example main packages need a cgo toolchain)"
			continue
		fi
		if ! run_for "$target" go build ./...; then
			echo "portable: build failed for $target" >&2
			exit 1
		fi
		echo ok
	done
	echo "portable: $count targets build"
else
	test_targets=$quick_targets
fi

for target in $test_targets; do
	printf 'test-compile %-18s ' "$target"
	if ! run_for "$target" go test -exec=/bin/true -count=1 ./...; then
		echo "portable: the test files do not compile for $target" >&2
		exit 1
	fi
	echo ok
done
echo "portable: the test files compile on every listed target"
