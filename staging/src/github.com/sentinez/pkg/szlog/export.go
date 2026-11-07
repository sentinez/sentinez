// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package szlog

import (
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/protobuf/proto"
)

// EventExporter receives the events of the loggers given WithEventExporter,
// as the proto messages they were logged with, e.g. a Kafka producer
// (shared/szkafka).
type EventExporter interface {
	// Export must not keep event once it returns: the caller may release
	// it to a pool right after. It should not block.
	Export(kind typepb.LogType, level Level, msg string, event proto.Message)
}

// WithEventExporter hands this logger's events to e in place of OTLP: the
// OTLP options are then ignored.
func WithEventExporter(e EventExporter) Option {
	return func(o *loggingOptions) { o.exporter = e }
}
