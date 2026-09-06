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

FROM golang:1.27.1-alpine AS builder

# Move to working directory (/src).
WORKDIR /src

# Set the timezone and install CA certificates
RUN apk --no-cache add ca-certificates tzdata
RUN apk add --no-cache git

COPY . ./
RUN go mod tidy

WORKDIR /src

# Set necessary environment variables needed for our image and build the API server.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go mod tidy
RUN go build -ldflags="-s -w" -o /bin/main ./cmd/greeter/v1

# Remove all source code files
RUN rm -r *
# Clean up
RUN go clean -modcache

# Stage run
FROM alpine AS runner

WORKDIR /app

COPY --from=builder /bin/main /bin/main

# Set the timezone and install CA certificates
RUN apk --no-cache add ca-certificates tzdata

CMD [ "main" ]