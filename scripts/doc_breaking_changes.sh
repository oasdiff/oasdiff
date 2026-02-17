#!/bin/sh

# Use C locale for deterministic sort order across macOS and Linux
export LC_ALL=C

echo "# Examples of Breaking and Non-Breaking Changes"
echo "These examples are automatically generated from unit tests."

echo "## Examples of breaking changes"
grep -n '// BC:' checker/*.go | grep "is breaking" | sort -u -k2 | sed -n "s/\([^0-9]*\):\([0-9]*\):\/\/ BC: \(.*\)/[\3](\.\.\/\1?plain=1#L\2)  /p" | sort

echo ""
echo "## Examples of non-breaking changes"
grep -n '// BC:' checker/*.go | grep "is not breaking" | sort -u -k2 | sed -n "s/\([^0-9]*\):\([0-9]*\):\/\/ BC: \(.*\)/[\3](\.\.\/\1?plain=1#L\2)  /p" | sort

echo ""
echo "## Examples of info-level changes for changelog"
grep -n '// CL:' checker/*.go | sort -u -k2 | sed -n "s/\([^0-9]*\):\([0-9]*\):\/\/ CL: \(.*\)/[\3](\.\.\/\1?plain=1#L\2)  /p" | sort
