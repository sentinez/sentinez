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

package centralsync_test

import (
	"context"
	"encoding/base64"
	"io"
	"net"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	centraldatahdl "github.com/sentinez/controlplane/apps/centraldata/v1/handler"
	memrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/memory"
	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/core/szgrpc"
	"github.com/sentinez/pkg/szconfig"
	"github.com/sentinez/pkg/szsign"
	"github.com/sentinez/pkg/sztable"
	"github.com/sentinez/sentinez/api/client"
	"github.com/sentinez/sentinez/api/client/connection"
	"github.com/sentinez/sentinez/api/client/connection/connectiontest"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/centralsync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _syncKey = []byte("0123456789abcdef0123456789abcdef")

// central a centraldata served over mTLS with the real interceptors.
type central struct {
	addr   string
	ca     *connectiontest.CA
	memory *memrepo.Memory
}

func startCentral(t *testing.T) *central {
	t.Helper()

	szconfig.SetEnv(settingpb.Senz_SENZ_SERVICE_KEY,
		base64.StdEncoding.EncodeToString([]byte("service-secret")))
	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY,
		base64.StdEncoding.EncodeToString([]byte("user-secret")))

	memory, err := memrepo.New()
	require.NoError(t, err)
	require.NoError(t, memory.Upsert(sztable.SecuritySecRules,
		&securitypb.SecRule{Id: "srule_1", Name: "one"},
		&securitypb.SecRule{Id: "srule_2", Name: "two"}))

	ca := connectiontest.NewCA(t)
	creds, err := ca.Issue(t, "centraldata").ServerCredentials()
	require.NoError(t, err)

	service := centraldatasvc.New(nil, memory, sztable.SecuritySecRules)
	srv := szgrpc.New(append(authz.ServerOptions(), grpc.Creds(creds))...)
	centraldatapb.RegisterCentralDataServiceServer(srv.AsServer(),
		centraldatahdl.New(service))

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = srv.AsServer().Serve(lis) }()
	t.Cleanup(srv.AsServer().Stop)

	return &central{addr: lis.Addr().String(), ca: ca, memory: memory}
}

func serviceToken(t *testing.T, name string) string {
	t.Helper()

	token, err := crypto.ServiceTokenGenerator(name, time.Hour)
	require.NoError(t, err)

	return token
}

// dialer the Dialer of an edge presenting tls and token.
func (c *central) dialer(tls connection.TLS, token string) centralsync.Dialer {
	return func(context.Context) (centraldatapb.CentralDataServiceClient,
		io.Closer, error) {

		cl, conn, err := client.DialCentralData(c.addr, tls, token)

		return cl, conn, err
	}
}

// pubsub a Source whose changes the test publishes by hand.
type pubsub struct {
	centralsync.Source
	msgs chan *redis.Message
}

func (p pubsub) Subscribe(context.Context,
) (<-chan *redis.Message, io.Closer, error) {
	return p.msgs, io.NopCloser(nil), nil
}

type collectingSink struct{ events chan string }

func (s collectingSink) Replace(c []*centraldatapb.Change) error {
	s.events <- "replace:" + string(rune('0'+len(c)))

	return nil
}

func (s collectingSink) Apply(c *centraldatapb.Change) error {
	s.events <- "apply:" + c.GetId()

	return nil
}

func TestEdgeLoadsAndFollowsCentraldata(t *testing.T) {
	c := startCentral(t)
	edge := c.ca.Issue(t, "edge-1")
	msgs := make(chan *redis.Message, 4)
	source := pubsub{
		Source: centralsync.NewSource(nil,
			c.dialer(edge, serviceToken(t, "edge-1"))),
		msgs: msgs,
	}
	sink := collectingSink{events: make(chan string, 4)}
	f := centralsync.New(source, sink, _syncKey, time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, f.Start(ctx))
	defer f.Close()
	assert.Equal(t, "replace:2", <-sink.events)

	// a change published after the snapshot
	version := c.memory.Version()
	data, err := szsign.Seal(_syncKey, &centraldatapb.Change{
		Table: sztable.SecuritySecRules.String(), Id: "srule_3",
		Op: centraldatapb.Change_OP_UPSERT, Version: version + 1,
		Epoch: c.memory.Epoch(),
	})
	require.NoError(t, err)
	msgs <- &redis.Message{Payload: string(data)}

	select {
	case got := <-sink.events:
		assert.Equal(t, "apply:srule_3", got)
	case <-ctx.Done():
		require.FailNow(t, "the live change was not applied")
	}
}

// firstError the error of the first Recv of the Sync of an edge.
func firstError(t *testing.T, c *central, tls connection.TLS,
	token string) error {
	t.Helper()

	src := centralsync.NewSource(nil, c.dialer(tls, token))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := src.Sync(ctx)
	if err != nil {
		return err
	}

	_, err = stream.Recv()

	return err
}

func TestSyncRefusesWhatIsNotAnEdgeOfTheCertificate(t *testing.T) {
	c := startCentral(t)
	edge := c.ca.Issue(t, "edge-1")
	user, err := crypto.TokenGenerator(&typepb.Context{
		Name:         "root",
		ControlPlane: typepb.ControlPlane_CONTROL_PLANE_ADMIN,
		ExpireAt:     timestamppb.New(time.Now().Add(time.Hour)),
	})
	require.NoError(t, err)

	tests := []struct {
		name      string
		giveToken string
		wantCode  codes.Code
	}{
		{"no token", "", codes.Unauthenticated},
		{"garbage", "not-a-token", codes.Unauthenticated},
		{"a user admin", user, codes.PermissionDenied},
		{"another service", serviceToken(t, "edge-2"), codes.PermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := firstError(t, c, edge, tt.giveToken)

			assert.Equal(t, tt.wantCode, status.Code(err), "%v", err)
		})
	}
}

func TestSyncRefusesAClientOfAnotherCA(t *testing.T) {
	c := startCentral(t)
	stranger := connectiontest.NewCA(t).Issue(t, "edge-1")

	err := firstError(t, c, stranger, serviceToken(t, "edge-1"))

	assert.Error(t, err)
}
