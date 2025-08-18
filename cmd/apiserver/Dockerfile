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

FROM golang:1.24.2-alpine AS builder

# Set necessary environment variables needed for our image and build the API server.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Move to working directory (/src).
WORKDIR /src

COPY resources/api/swagger/ /sentinez/resources/api/swagger/
COPY resources/api/specs/ /sentinez/resources/api/specs/
COPY cmd/apiserver/docker-entrypoint.sh /sentinez/docker-entrypoint.sh

COPY . ./

# Set the timezone and install CA certificates
RUN apk --no-cache add ca-certificates tzdata

RUN apk add --no-cache git
RUN go mod tidy

# Build the application
RUN go build -ldflags="-s -w" -o /bin/apiserver ./cmd/apiserver

# Clean up
RUN go clean -modcache

# Stage run
FROM alpine AS runner

# Set the timezone and install CA certificates
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/apiserver ./bin/apiserver
COPY --from=builder /sentinez/resources/api/swagger ./resources/api/swagger
COPY --from=builder /sentinez/resources/api/specs ./resources/api/specs

COPY --from=builder /sentinez/docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh

CMD [ "./docker-entrypoint.sh" ]