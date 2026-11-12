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


# Puts the SENZ_* lines of an env file in the Consul KV, where the services
# read them (see pkg/appconfig): shared by all of them under sentinez/env/,
# or for one service under sentinez/env/<service_key>/.
#
#   consul-env-put.sh ENV_FILE [SERVICE_KEY]
#
# CONSUL_ADDR is the address of Consul (default http://localhost:8500). The
# values are never printed. The keys that only the environment may set are
# skipped: they are ignored in Consul anyway.

set -e
set -o pipefail

# the same keys as _envOnly in pkg/appconfig/overlay.go: keep them equal
ENV_ONLY=" SENZ_CONSUL_URI SENZ_ADDRESS SENZ_MEMBERSHIP_ADDRESS SENZ_DISCOVERY_ADDRESS SENZ_ZONE "

error_exit() {
  echo "Error: $1" >&2
  exit 1
}

env_file="${1:-}"
service="${2:-}"
consul_addr="${CONSUL_ADDR:-http://localhost:8500}"

[ -n "$env_file" ] || error_exit "usage: $0 ENV_FILE [SERVICE_KEY]"
[ -f "$env_file" ] || error_exit "no such env file: $env_file"

prefix="sentinez/env/"
[ -z "$service" ] || prefix="${prefix}${service}/"

count=0
while IFS= read -r line || [ -n "$line" ]; do
  line="${line#"${line%%[![:space:]]*}"}"
  case "$line" in "" | "#"*) continue ;; esac
  line="${line#export }"

  key="${line%%=*}"
  value="${line#*=}"
  key="${key%"${key##*[![:space:]]}"}"

  case "$key" in
    SENZ_*) ;;
    *) continue ;;
  esac

  case "$ENV_ONLY" in
    *" $key "*)
      echo "skip $key: only the environment may set it"
      continue
      ;;
  esac

  # a value in matching quotes, as godotenv reads it
  case "$value" in
    \"*\") value="${value#\"}"; value="${value%\"}" ;;
    \'*\') value="${value#\'}"; value="${value%\'}" ;;
  esac

  [ -n "$value" ] || continue

  printf '%s' "$value" |
    curl -sSf -X PUT --data-binary @- "$consul_addr/v1/kv/${prefix}${key}" >/dev/null ||
    error_exit "put ${prefix}${key} failed"
  echo "put ${prefix}${key}"
  count=$((count + 1))
done <"$env_file"

echo "$count keys put under ${prefix}"
