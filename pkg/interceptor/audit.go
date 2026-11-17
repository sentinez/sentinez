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

// Package interceptor holds the gRPC server interceptors of the control
// plane: the authorization of the calls and their audit log.
package interceptor

import (
	"context"
	"time"

	"github.com/sentinez/controlplane/pkg/headers"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szrand"
	auditpb "github.com/sentinez/sentinez/api/proto/sentinez/types/audit/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const _auditMessage = "audit"

// Audit returns a server interceptor that writes to log an entry for each
// call of a method whose x_method option sets audit, also when the call
// fails. Install it before the authorization interceptor to record the
// calls it denies. The request is logged as protojson, its debug_redact
// fields hidden; the response is only read for the tenant id.
func Audit(log szlog.Log) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (any, error) {

		if !MethodOf(info.FullMethod).GetAudit() {
			return handler(ctx, req)
		}

		start := time.Now()
		slot := &caller{}
		resp, err := handler(context.WithValue(ctx, callerKey{}, slot), req)
		call := &call{method: info.FullMethod, start: start,
			req: messageOf(req), resp: messageOf(resp), err: err,
			user: slot.userOf(ctx), addr: headers.ClientAddr(ctx),
			agent: headers.UserAgent(ctx)}
		log.Info(_auditMessage, call.entry())

		return resp, err
	}
}

// callerKey is the context key of the *caller [Audit] places for
// [Authorize] to fill.
type callerKey struct{}

// caller is the caller [Authorize] verified, so that [Audit] does not
// verify its token again.
type caller struct {
	auth *headers.Auth
	// seen is whether Authorize ran: auth is nil without a valid token.
	seen bool
}

// setCaller hands auth, nil without a valid token, to the [Audit] of ctx.
func setCaller(ctx context.Context, auth *headers.Auth) {
	if c, ok := ctx.Value(callerKey{}).(*caller); ok {
		c.auth, c.seen = auth, true
	}
}

// userOf returns the caller Authorize saw, else the one of the token of
// ctx; nil without a valid token.
func (c *caller) userOf(ctx context.Context) *typepb.Context {
	auth := c.auth
	if !c.seen {
		auth, _ = headers.GetAuth(ctx)
	}
	if auth == nil {
		return nil
	}

	return auth.Context
}

// call is one finished call of an audited method.
type call struct {
	method      string
	start       time.Time
	req, resp   proto.Message
	err         error
	user        *typepb.Context
	addr, agent string
}

func (c *call) entry() *auditpb.AuditLog {
	entry := &auditpb.AuditLog{
		Id:           szrand.NewUUID(szprefix.AuditLog),
		Time:         timestamppb.New(c.start),
		Method:       c.method,
		RemoteAddr:   c.addr,
		UserAgent:    c.agent,
		Request:      redact(c.req),
		Code:         uint32(status.Code(c.err)),
		Latency:      durationpb.New(time.Since(c.start)),
		TenantId:     tenantIdOf(c.req, c.resp),
		UserId:       c.user.GetUserId(),
		Name:         c.user.GetName(),
		ControlPlane: c.user.GetControlPlane(),
	}
	if c.err != nil {
		entry.Error = status.Convert(c.err).Message()
	}

	return entry
}

func messageOf(v any) proto.Message {
	m, _ := v.(proto.Message)
	return m
}
