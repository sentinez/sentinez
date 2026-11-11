// Copyright 2025 Sentinéz Labs.
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

package iamsvc

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/sentinez/controlplane/pkg/loginlimit"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// The failed sign-ins a client may make in _failureWindow: against one
// account, which stops guessing its password, and in all, which stops
// trying one password on many accounts.
const (
	_accountFailures = 10
	_clientFailures  = 50
	_failureWindow   = 15 * time.Minute
)

var errTooManyAttempts = status.Error(codes.ResourceExhausted,
	"too many failed attempts, try again later")

type signInLimits struct {
	byAccount *loginlimit.Limiter
	byClient  *loginlimit.Limiter
}

func newSignInLimits() *signInLimits {
	return &signInLimits{
		byAccount: loginlimit.New(_accountFailures, _failureWindow),
		byClient:  loginlimit.New(_clientFailures, _failureWindow),
	}
}

// attempt is who tries to sign in, and as whom.
type attempt struct {
	client string
	// account is empty when the attempt names none.
	account string
}

func attemptOf(ctx context.Context, username string) attempt {
	client := clientAddr(ctx)
	if username == "" {
		return attempt{client: client}
	}

	return attempt{
		client:  client,
		account: client + "|" + strings.ToLower(username),
	}
}

func (l *signInLimits) check(a attempt) error {
	if l.byClient.Blocked(a.client) ||
		a.account != "" && l.byAccount.Blocked(a.account) {

		return errTooManyAttempts
	}

	return nil
}

func (l *signInLimits) fail(a attempt) {
	l.byClient.Fail(a.client)

	if a.account != "" {
		l.byAccount.Fail(a.account)
	}
}

// succeed forgets the failures against the account, not those of the
// client: a user with an account must not clear what it did to others.
func (l *signInLimits) succeed(a attempt) {
	if a.account != "" {
		l.byAccount.Reset(a.account)
	}
}

// clientAddr returns the address of the client: the last of the
// X-Forwarded-For the gateway sets, which is the peer it saw, not what the
// client wrote in front of it.
func clientAddr(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if forwarded := md.Get("x-forwarded-for"); len(forwarded) != 0 {
			hops := strings.Split(forwarded[len(forwarded)-1], ",")
			if addr := strings.TrimSpace(hops[len(hops)-1]); addr != "" {
				return addr
			}
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
			return host
		}
	}

	return "unknown"
}

// limited runs call unless the client failed too often, and counts its
// failure.
func limited[Resp any](l *signInLimits, a attempt, call func() (Resp, error),
) (Resp, error) {
	var none Resp

	if err := l.check(a); err != nil {
		return none, err
	}

	resp, err := call()
	if err != nil {
		l.fail(a)
		return none, err
	}

	return resp, nil
}

// Login signs the user in, unless the client failed too often.
func (srv *IAMService) Login(ctx context.Context,
	req *iampb.LoginRequest) (*iampb.LoginResponse, error) {

	a := attemptOf(ctx, req.GetEmailOrUsername())

	resp, err := limited(srv.limits, a, func() (*iampb.LoginResponse, error) {
		return srv.login(ctx, req)
	})
	if err == nil {
		srv.limits.succeed(a)
	}

	return resp, err
}

// PasskeyLoginChallenge starts a passkey sign-in, unless the client failed
// too often. Asking for an unknown account counts as a failure: it is how
// the accounts would be listed.
func (srv *IAMService) PasskeyLoginChallenge(ctx context.Context,
	req *iampb.PasskeyLoginChallengeRequest,
) (*iampb.PasskeyLoginChallengeResponse, error) {
	a := attemptOf(ctx, req.GetEmailOrUsername())

	return limited(srv.limits, a,
		func() (*iampb.PasskeyLoginChallengeResponse, error) {
			return srv.passkeyLoginChallenge(ctx, req)
		})
}

// PasskeyLoginVerify finishes a passkey sign-in, unless the client failed
// too often.
func (srv *IAMService) PasskeyLoginVerify(ctx context.Context,
	req *iampb.PasskeyLoginVerifyRequest,
) (*iampb.PasskeyLoginVerifyResponse, error) {
	a := attemptOf(ctx, "")

	return limited(srv.limits, a,
		func() (*iampb.PasskeyLoginVerifyResponse, error) {
			return srv.passkeyLoginVerify(ctx, req)
		})
}
