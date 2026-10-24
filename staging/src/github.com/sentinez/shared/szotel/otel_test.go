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

package szotel

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// collector is an in-process OTLP/gRPC logs endpoint keeping what it gets.
type collector struct {
	collogspb.UnimplementedLogsServiceServer

	mu      sync.Mutex
	records []*logspb.LogRecord
}

func (c *collector) Export(_ context.Context,
	req *collogspb.ExportLogsServiceRequest,
) (*collogspb.ExportLogsServiceResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, rl := range req.GetResourceLogs() {
		for _, sl := range rl.GetScopeLogs() {
			c.records = append(c.records, sl.GetLogRecords()...)
		}
	}

	return &collogspb.ExportLogsServiceResponse{}, nil
}

func startCollector(t *testing.T) (*collector, string) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	c := &collector{}
	srv := grpc.NewServer()
	collogspb.RegisterLogsServiceServer(srv, c)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(lis)
	}()
	t.Cleanup(func() {
		srv.Stop()
		wg.Wait()
	})

	return c, lis.Addr().String()
}

func newRuleEvent(t *testing.T) *secrulepb.RuleEvent {
	t.Helper()
	params, err := structpb.NewStruct(map[string]any{
		"status":  403,
		"message": "Forbidden",
		"close":   true,
	})
	require.NoError(t, err)

	return &secrulepb.RuleEvent{
		Id:                "sr_1",
		MatchedSecruleIds: []string{"a", "b"},
		Messages:          []string{"sqli"},
		TenantId:          "tn_1",
		MatchedTime:       1761200000000,
		Action: &rulepb.Action{
			Type:   rulepb.ActionType_ACTION_TYPE_BLOCK,
			Params: params,
		},
		Service:       rulepb.RuleService_RULE_SERVICE_SEC_RULE,
		RequestDomain: "example.com",
		RequestIp:     "10.0.0.1",
	}
}

// export logs event through szlog and an OTLP/gRPC exporter, returning the
// record the collector received.
func export(t *testing.T, kind typepb.LogType,
	event proto.Message) *logspb.LogRecord {
	t.Helper()
	c, addr := startCollector(t)
	exp, err := otlploggrpc.New(context.Background(),
		otlploggrpc.WithEndpoint(addr), otlploggrpc.WithInsecure())
	require.NoError(t, err)
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exp)),
	)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	l := szlog.NewLogCloser("test", kind, szlog.LevelInfo,
		szlog.WithOTLPProvider(provider))
	l.Info("test", event, io.NopCloser(nil))

	c.mu.Lock()
	defer c.mu.Unlock()
	require.Len(t, c.records, 1)

	return c.records[0]
}

func newRequestEvent() *httppb.RequestEvent {
	return &httppb.RequestEvent{
		Id:     "rq_1",
		Method: "GET",
		Queries: map[string]*httppb.QueryValue{
			"q": {Values: []string{"1", "2"}},
		},
		Headers: map[string]*httppb.HeaderValue{
			"accept": {Values: []string{"text/html", "*/*"}},
		},
		Status:    200,
		Asn:       13335,
		Timestamp: 1761200000000,
	}
}

func newRequest() *httppb.Request {
	return &httppb.Request{
		Id:   "rq_1",
		Path: []byte("/login"),
		Headers: []*httppb.RequestHeader{
			{Key: []byte("cookie"), Values: [][]byte{[]byte("a"), []byte("b")}},
		},
		Queries: []*httppb.RequestQuery{
			{Key: []byte("q"), Values: [][]byte{[]byte("1")}},
		},
		Timestamp: timestamppb.New(time.UnixMilli(1761200000000)),
	}
}

func TestUnmarshalRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		giveKind  typepb.LogType
		giveEvent proto.Message
	}{
		{
			name:      "secrule",
			giveKind:  typepb.LogType_LOG_TYPE_SEC_RULE,
			giveEvent: newRuleEvent(t),
		},
		{
			name:      "http request event",
			giveKind:  typepb.LogType_LOG_TYPE_HTTP,
			giveEvent: newRequestEvent(),
		},
		{
			name:      "http request",
			giveKind:  typepb.LogType_LOG_TYPE_HTTP,
			giveEvent: newRequest(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := export(t, tt.giveKind, tt.giveEvent)
			got := tt.giveEvent.ProtoReflect().New().Interface()

			require.NoError(t, Unmarshal(rec, got))
			assert.Equal(t, tt.giveKind, Kind(rec))
			assert.True(t, proto.Equal(tt.giveEvent, got), "got %v", got)
		})
	}
}

func TestUnmarshalResetsMessage(t *testing.T) {
	rec := recordOf(kv("tenant_id", strValue("tn_1")))
	got := &secrulepb.RuleEvent{Id: "stale"}

	require.NoError(t, Unmarshal(rec, got))
	assert.Empty(t, got.GetId())
	assert.Equal(t, "tn_1", got.GetTenantId())
}

func TestUnmarshalNoEvent(t *testing.T) {
	err := Unmarshal(&logspb.LogRecord{}, &secrulepb.RuleEvent{})

	require.ErrorIs(t, err, ErrNoEvent)
}

func TestUnmarshalTypeMismatch(t *testing.T) {
	rec := recordOf(kv("id", intValue(1)))

	err := Unmarshal(rec, &secrulepb.RuleEvent{})

	require.ErrorContains(t, err, "id: want string")
}

func TestUnmarshalIgnoresUnknownFields(t *testing.T) {
	rec := recordOf(kv("id", strValue("sr_1")), kv("added", intValue(1)))
	got := &secrulepb.RuleEvent{}

	require.NoError(t, Unmarshal(rec, got))
	assert.Equal(t, "sr_1", got.GetId())
}

func TestKind(t *testing.T) {
	tests := []struct {
		name     string
		giveRec  *logspb.LogRecord
		wantKind typepb.LogType
	}{
		{
			name:     "missing",
			giveRec:  &logspb.LogRecord{},
			wantKind: typepb.LogType_LOG_TYPE_UNSPECIFIED,
		},
		{
			name: "unknown",
			giveRec: &logspb.LogRecord{Attributes: []*commonpb.KeyValue{
				kv(_kindKey, strValue("LOG_TYPE_NOPE")),
			}},
			wantKind: typepb.LogType_LOG_TYPE_UNSPECIFIED,
		},
		{
			name: "waf",
			giveRec: &logspb.LogRecord{Attributes: []*commonpb.KeyValue{
				kv(_kindKey, strValue("LOG_TYPE_WAF")),
			}},
			wantKind: typepb.LogType_LOG_TYPE_WAF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKind, Kind(tt.giveRec))
		})
	}
}

func recordOf(fields ...*commonpb.KeyValue) *logspb.LogRecord {
	event := &commonpb.AnyValue{Value: &commonpb.AnyValue_KvlistValue{
		KvlistValue: &commonpb.KeyValueList{Values: fields},
	}}

	return &logspb.LogRecord{
		Attributes: []*commonpb.KeyValue{kv(_eventKey, event)},
	}
}

func kv(key string, v *commonpb.AnyValue) *commonpb.KeyValue {
	return &commonpb.KeyValue{Key: key, Value: v}
}

func strValue(s string) *commonpb.AnyValue {
	return &commonpb.AnyValue{
		Value: &commonpb.AnyValue_StringValue{StringValue: s},
	}
}

func intValue(n int64) *commonpb.AnyValue {
	return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: n}}
}
