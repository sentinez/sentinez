// Copyright 2025 Duc-Hung Ho.
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

package apiserver

import (
	"github.com/sentinez/pkg/szkafka"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsdklog"
	"github.com/sentinez/pkg/sztopic"
	apiserverpb "github.com/sentinez/sentinez/api/proto/sentinez/gateway/apiserver/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/pkg/interceptor"
	"google.golang.org/grpc"
)

// auditInterceptor returns the interceptor that writes the audit log of the
// methods that set x_method.audit, and the exporter it writes through, for
// the caller to flush at shutdown. The log goes where the edge traffic logs
// do (see szsdklog.New): Kafka, else OTLP, else the console only.
func auditInterceptor(conf *settingpb.Config,
) (grpc.UnaryServerInterceptor, *szsdklog.Exporter, error) {
	exporter, err := szsdklog.New(conf, szkafka.WithTopics(
		map[typepb.LogType]sztopic.Topic{
			typepb.LogType_LOG_TYPE_AUDIT: sztopic.EventActivityAuditLog,
		}))
	if err != nil {
		return nil, nil, err
	}

	log := szlog.NewLog(apiserverpb.GetMetaApiserverServiceKey(),
		typepb.LogType_LOG_TYPE_AUDIT, szlog.LevelInfo, exporter.Option())

	return interceptor.Audit(log), exporter, nil
}
