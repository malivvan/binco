#!/usr/bin/env python3
"""Verify that README.md and AGENTS.md match the repository.

The checks are mechanical so that documentation drift fails CI instead of
surviving until a user hits it:

  * every internal link (``](#anchor)``) resolves to a heading;
  * every ``package.Symbol`` mentioned in either file is exported by that
    package (looked up with ``go doc``);
  * README section and subsection numbers are sequential, and the table of
    contents lists exactly those headings, in order;
  * every ``README §N`` reference in AGENTS.md names an existing section;
  * none of the retired names in RETIRED appears.

Run it with ``make doc-check``; the ``Docs`` CI job runs it on every push.
"""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

# The import aliases the documentation uses, mapped to the package directory.
ALIASES = {
    "binco": ".",
    "varnum": "varnum",
    "scalar": "scalar",
    "composite": "composite",
    "zerocopy": "zerocopy",
    "ptrmap": "ptrmap",
    "typeid": "typeid",
    "wirefmt": "wirefmt",
    "wcomposite": "wirefmt/composite",
    "wscalar": "wirefmt/scalar",
    "wvarnum": "wirefmt/varnum",
    "wzerocopy": "wirefmt/zerocopy",
    "wptrmap": "wirefmt/ptrmap",
    "wtypeid": "wirefmt/typeid",
    "genopts": "codegen/config/gen",
    "tpopts": "codegen/config/type",
    "stopts": "codegen/config/struct",
    "fldopts": "codegen/config/field",
    "veropts": "codegen/config/versioned",
    "intropts": "codegen/config/interface",
    "arropts": "config/array",
    "bslopts": "config/byte_slice",
    "mapopts": "config/map",
    "slopts": "config/slice",
    "stropts": "config/string",
    "codegen": "codegen",
}

# Names that must not come back: they describe removed API or old spellings.
RETIRED = {
    "zerocpy": "the package is called zerocopy",
    "NewValid": "every constructor honours validators; there is no NewValid* family",
    "ErrTooSmallByteSlice": "split into ErrTruncated and ErrBufferTooSmall",
    "UintMaxVarintLen": "int and uint use the uint64 limits",
    "UintMaxLastByte": "int and uint use the uint64 limits",
}

SYMBOL_RE = re.compile(r"\b(" + "|".join(map(re.escape, ALIASES)) + r")\.([A-Z]\w*)")


def slug(text: str) -> str:
    """Return the GitHub anchor a heading generates."""
    text = text.lower()
    text = re.sub(r"[^\w\- ]", "", text)
    return text.replace(" ", "-")


def exported(dirname: str) -> set[str]:
    """Return the exported identifiers of one package, constants included."""
    result = subprocess.run(
        ["go", "doc", "-all", f"./{dirname}"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise SystemExit(f"doccheck: go doc ./{dirname} failed:\n{result.stderr}")

    names: set[str] = set()
    in_block = False
    for raw in result.stdout.splitlines():
        line = raw.strip()
        if re.match(r"^(const|var) \($", line):
            in_block = True
            continue
        if in_block and line == ")":
            in_block = False
            continue
        if in_block:
            match = re.match(r"^([A-Z]\w*)", line)
            if match:
                names.add(match.group(1))
            continue
        match = re.match(r"^(?:func|type|var|const)\s+(?:\([^)]*\)\s*)?([A-Z]\w*)", line)
        if match:
            names.add(match.group(1))
    return names


def headings(text: str) -> list[tuple[int, str]]:
    """Return (level, title) for every level-2 and level-3 heading."""
    found = []
    for line in text.splitlines():
        match = re.match(r"^(#{2,3}) (.+)$", line)
        if match:
            found.append((len(match.group(1)), match.group(2)))
    return found


def check_anchors(text: str, name: str, problems: list[str]) -> None:
    known = {slug(title) for _, title in headings(text)}
    for match in re.finditer(r"\]\(#([^)]+)\)", text):
        if match.group(1) not in known:
            problems.append(f"{name}: link #{match.group(1)} has no matching heading")


def check_symbols(text: str, name: str, api: dict[str, set[str]], problems: list[str]) -> None:
    for match in SYMBOL_RE.finditer(text):
        alias, symbol = match.group(1), match.group(2)
        if symbol not in api[alias]:
            problems.append(
                f"{name}: {alias}.{symbol} is not exported by ./{ALIASES[alias]}"
            )


def check_numbering(text: str, problems: list[str]) -> None:
    current: int | None = None
    section = 0
    subsections: dict[int, int] = {}
    for level, title in headings(text):
        if level == 2:
            match = re.match(r"^(\d+)\. ", title)
            if not match:
                current = None
                continue
            current = int(match.group(1))
            section += 1
            if current != section:
                problems.append(
                    f"README.md: {title} is section {current}, want {section}"
                )
            continue
        match = re.match(r"^(\d+)\.(\d+) ", title)
        if not match:
            continue
        parent, sub = int(match.group(1)), int(match.group(2))
        if current is None or parent != current:
            problems.append(
                f"README.md: {title} is numbered under section {current}, not {parent}"
            )
        want = subsections.get(parent, 0) + 1
        if sub != want:
            problems.append(f"README.md: {title} is subsection {sub}, want {want}")
        subsections[parent] = sub


def check_contents(text: str, problems: list[str]) -> None:
    lines = text.splitlines()
    try:
        start = lines.index("## Contents") + 1
    except ValueError:
        problems.append("README.md: no '## Contents' heading")
        return
    end = next(
        (i for i in range(start, len(lines)) if lines[i].startswith("## ")),
        len(lines),
    )

    listed: list[tuple[int, str]] = []
    for line in lines[start:end]:
        match = re.match(r"^(\s*)\* \[(.+)\]\(#([^)]+)\)$", line)
        if match:
            listed.append((len(match.group(1)) // 2, match.group(3)))

    expected: list[tuple[int, str]] = []
    for level, title in headings(text):
        if level == 2 and re.match(r"^\d+\. ", title):
            expected.append((0, slug(title)))
        elif level == 3 and re.match(r"^\d+\.\d+ ", title):
            expected.append((1, slug(title)))

    if listed == expected:
        return
    problems.append("README.md: the table of contents does not match the headings")
    for index, (have, want) in enumerate(zip(listed, expected)):
        if have != want:
            problems.append(f"  entry {index}: contents has {have}, headings have {want}")
            break
    if len(listed) != len(expected):
        problems.append(
            f"  contents lists {len(listed)} entries, the headings need {len(expected)}"
        )


def check_readme_refs(text: str, numbers: set[int], problems: list[str]) -> None:
    for match in re.finditer(r"README §(\d+)", text):
        if int(match.group(1)) not in numbers:
            problems.append(
                f"AGENTS.md: README §{match.group(1)} does not name an existing section"
            )


def check_retired(text: str, name: str, problems: list[str]) -> None:
    for term, why in RETIRED.items():
        if term in text:
            problems.append(f"{name}: retired name {term!r} ({why})")


def main() -> int:
    readme = (ROOT / "README.md").read_text(encoding="utf-8")
    agents = (ROOT / "AGENTS.md").read_text(encoding="utf-8")

    api = {alias: exported(dirname) for alias, dirname in ALIASES.items()}
    problems: list[str] = []

    check_anchors(readme, "README.md", problems)
    check_symbols(readme, "README.md", api, problems)
    check_symbols(agents, "AGENTS.md", api, problems)
    check_numbering(readme, problems)
    check_contents(readme, problems)

    numbers = {int(m.group(1)) for m in re.finditer(r"^## (\d+)\. ", readme, re.M)}
    check_readme_refs(agents, numbers, problems)
    check_retired(readme, "README.md", problems)
    check_retired(agents, "AGENTS.md", problems)

    if problems:
        for problem in problems:
            print(f"doccheck: {problem}", file=sys.stderr)
        return 1

    print(
        f"doccheck: {len(api)} packages, README.md and AGENTS.md are consistent",
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
