#!/bin/sh
# covercheck.sh reports the statement coverage of a cover profile and fails
# unless every statement is covered.
#
# usage: covercheck.sh [coverage.out] [NOCOVER.md]
#
# The profile covers every package in the module. A statement listed in the
# "nocover" block of NOCOVER.md, one
# module-relative "path:line" per line, is treated as covered. Every other
# statement must be covered, and every entry must still name an uncovered
# block, so the exception list cannot rot.

set -eu

profile=${1:-coverage.out}
nocover=${2:-NOCOVER.md}
module=$(go list -m)

awk -v module="$module" -v nocover="$nocover" '
BEGIN {
	while ((getline line < nocover) > 0) {
		sub(/#.*/, "", line)
		gsub(/^[ \t]+|[ \t]+$/, "", line)
		if (line ~ /^[A-Za-z0-9_.\/-]+:[0-9]+$/) {
			exceptions[line] = 1
		}
	}
}
NR == 1 { next }
{
	split($1, parts, ":")
	start = parts[2]
	sub(/\..*$/, "", start)
	file = parts[1]
	sub("^" module "/", "", file)
	key = file ":" start

	total += $2
	if ($3 + 0 > 0) {
		covered += $2
		if (key in exceptions) {
			print "covercheck: NOCOVER.md lists " key ", but that statement is covered" > "/dev/stderr"
			stale = 1
		}
	} else if (key in exceptions) {
		covered += $2
		seen[key] = 1
	}
}
END {
	if (total == 0) {
		print "covercheck: the profile has no statements" > "/dev/stderr"
		exit 1
	}
	for (key in exceptions) {
		if (!(key in seen)) {
			print "covercheck: NOCOVER.md lists " key ", which is not an uncovered block" > "/dev/stderr"
			stale = 1
		}
	}

	printf "covercheck: %.1f%% of %d statements", covered * 100 / total, total
	if (length(exceptions) > 0) {
		printf " (%d documented exceptions)", length(exceptions)
	}
	print ""

	if (covered < total) {
		printf "covercheck: %d statements are neither covered nor documented in NOCOVER.md\n", total - covered > "/dev/stderr"
		exit 1
	}
	if (stale) {
		exit 1
	}
}' "$profile"
