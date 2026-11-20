#!/bin/bash

# Copyright 2026 Duc-Hung Ho.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# bench.sh runs the benchmarks of the edge request path and writes their
# medians (ns/op, B/op, allocs/op) into the README, between the
# <!-- bench:start --> and <!-- bench:end --> markers.
#
# Usage: hack/bench.sh [count]
#   count  runs of each benchmark, the median is kept (default
#          $BENCH_COUNT, else 6)

set -e
set -o pipefail

# a decimal point, whatever the locale of the machine
export LC_ALL=C

root="$(cd "$(dirname "$0")/.." && pwd)"
readme="$root/README.md"
count="${1:-${BENCH_COUNT:-6}}"

# the packages of the edge request path: the chain plugins and the
# net/http context they run on
packages=(./internal/handlers/... ./pkg/network/httpx/)

_start='<!-- bench:start -->'
_end='<!-- bench:end -->'

if ! grep -qF "$_start" "$readme" || ! grep -qF "$_end" "$readme"; then
  echo "bench: $readme lacks the $_start / $_end markers" >&2
  exit 1
fi

raw="$(mktemp)"
table="$(mktemp)"
trap 'rm -f "$raw" "$table"' EXIT

cd "$root"
echo "bench: ${packages[*]}, $count runs each"
go test -run '^$' -bench . -benchmem -count "$count" "${packages[@]}" |
  tee "$raw" | grep -E '^(ok|FAIL|---)' || true

if grep -q '^FAIL' "$raw"; then
  echo "bench: go test failed" >&2
  exit 1
fi

if ! grep -q '^Benchmark' "$raw"; then
  echo "bench: no benchmark result" >&2
  exit 1
fi

# the table of the medians, in the order the benchmarks ran
awk -v count="$count" -v goversion="$(go env GOVERSION)" \
  -v platform="$(go env GOOS)/$(go env GOARCH)" \
  -v date="$(date -u +%Y-%m-%d)" '
function median(values, name, n,   i, j, v, sorted) {
  for (i = 1; i <= n; i++) {
    v = values[name, i] + 0
    for (j = i - 1; j >= 1 && sorted[j] > v; j--) {
      sorted[j + 1] = sorted[j]
    }
    sorted[j + 1] = v
  }
  if (n % 2) {
    return sorted[(n + 1) / 2]
  }
  return (sorted[n / 2] + sorted[n / 2 + 1]) / 2
}

function number(v) {
  return v < 100 && v != int(v) ? sprintf("%.1f", v) : sprintf("%.0f", v)
}

/^cpu:/ { sub(/^cpu: */, ""); cpu = $0 }
/^pkg:/ { pkg = $2; sub(/^github\.com\/sentinez\/sentinez\//, "", pkg) }
/^Benchmark/ {
  name = $1
  sub(/^Benchmark/, "", name)
  sub(/-[0-9]+$/, "", name)
  if (!(name in runs)) {
    order[++total] = name
    packageOf[name] = pkg
  }
  k = ++runs[name]
  ns[name, k] = $3
  bytes[name, k] = $5
  allocs[name, k] = $7
}

END {
  printf "_%s, %s, %s; median of %d runs, %s._\n\n", \
    goversion, platform, cpu, count, date
  print "| Package | Benchmark | ns/op | B/op | allocs/op |"
  print "|---|---|--:|--:|--:|"
  for (i = 1; i <= total; i++) {
    name = order[i]
    n = runs[name]
    printf "| `%s` | %s | %s | %s | %s |\n", packageOf[name], name, \
      number(median(ns, name, n)), number(median(bytes, name, n)), \
      number(median(allocs, name, n))
  }
}' "$raw" >"$table"

# the table replaces what is between the markers
awk -v table="$table" -v start="$_start" -v end="$_end" '
index($0, start) {
  print
  print ""
  while ((getline line < table) > 0) {
    print line
  }
  print ""
  skip = 1
  next
}
index($0, end) { skip = 0 }
!skip { print }' "$readme" >"$readme.tmp"
mv "$readme.tmp" "$readme"

echo "bench: README.md updated"
