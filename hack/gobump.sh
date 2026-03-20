#!/bin/bash

# Copyright 2025 Duc-Hung Ho.
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

set -e
set -o pipefail

if [ -z "$1" ]; then
  echo "Usage: $0 <go-version>"
  echo "Example: $0 1.25.0"
  exit 1
fi

GO_VERSION=$1

echo "Updating Go version to $GO_VERSION across all packages..."

# Update root go.mod if it exists
if [ -f "go.mod" ]; then
  echo "Updating ./go.mod"
  go mod edit -go="$GO_VERSION"
  go mod tidy
fi

# Update go.mod in staging/ and api/ directories
find staging api -type f -name "go.mod" -not -path "*/vendor/*" -print0 2>/dev/null | while IFS= read -r -d '' modfile; do
  dir=$(dirname "$modfile")
  echo "Updating $dir/go.mod"
  (
    cd "$dir"
    go mod edit -go="$GO_VERSION"
    go mod tidy
  )
done

echo "Go version successfully bumped to $GO_VERSION."