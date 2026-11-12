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

# Issues the certificates of the internal mTLS (edge -> centraldata).
#
#   hack/mkcert-internal.sh <out-dir> <service> [<san> ...]
#
# Creates <out-dir>/ca.pem and ca-key.pem once (the internal CA), then
# <out-dir>/<service>.pem and <service>-key.pem, signed by it. <service>
# is the CN: it is the name of the service token of the edge, so use
# the same one for `szservicetoken -name`. A server also needs its DNS
# names / IPs as <san> (ex: centraldata sentinez.apps.centraldata).
#
# Keep ca-key.pem offline; services get ca.pem, their cert and their key.

set -euo pipefail

if [ "$#" -lt 2 ]; then
  echo "usage: $0 <out-dir> <service> [<san> ...]" >&2
  exit 1
fi

out="$1"
name="$2"
shift 2

umask 077
mkdir -p "$out"

if [ ! -f "$out/ca.pem" ]; then
  openssl ecparam -name prime256v1 -genkey -noout -out "$out/ca-key.pem"
  openssl req -x509 -new -key "$out/ca-key.pem" -sha256 -days 3650 \
    -subj "/CN=sentinez-internal-ca" -out "$out/ca.pem"
fi

san="DNS:$name"
for s in "$@"; do
  if [[ "$s" =~ ^[0-9.]+$ || "$s" == *:* ]]; then
    san="$san,IP:$s"
  else
    san="$san,DNS:$s"
  fi
done

openssl ecparam -name prime256v1 -genkey -noout -out "$out/$name-key.pem"
openssl req -new -key "$out/$name-key.pem" -subj "/CN=$name" \
  -out "$out/$name.csr"

ext=$(mktemp)
trap 'rm -f "$ext" "$out/$name.csr"' EXIT
cat >"$ext" <<EOT
subjectAltName=$san
extendedKeyUsage=serverAuth,clientAuth
keyUsage=digitalSignature
EOT

openssl x509 -req -in "$out/$name.csr" -CA "$out/ca.pem" \
  -CAkey "$out/ca-key.pem" -CAcreateserial -sha256 -days 365 \
  -extfile "$ext" -out "$out/$name.pem"

echo "issued $out/$name.pem (CN=$name, $san)"
