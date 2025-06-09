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

// Package greeterdomain provides the business logic for the greeter service.
package greeterdomain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sentinez/sentinez/pkg/std/zlog"

	"github.com/sentinez/sentinez/pkg/std/errors"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/domain/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/shared/v1"
)

// New creates a new Greeter module.
func New() IGreeter {
	return &Greeter{}
}

// IGreeter defines the interface for the Greeter biz module.
type IGreeter interface {
	domain.GreeterDomainServiceServer
}

// Greeter implements GreeterServiceServer, business logic for the service.
type Greeter struct {
	domain.UnimplementedGreeterDomainServiceServer
}

// SayHello implements GreeterServiceServer.
func (g *Greeter) SayHello(_ context.Context,
	msg *domain.SayHelloRequest) (*domain.SayHelloResponse, error) {

	msgs, _ := json.Marshal(msg)
	zlog.Debug("Received a SayHello request" + string(msgs))

	name := msg.GetName()
	t := time.Now().String()

	if name == "error" {
		return nil, errors.StatusForbidden
	}

	return &domain.SayHelloResponse{
		Response: &shared.SayHelloResponse{
			Message: "Reply " + name + " at " + t,
		},
	}, nil
}
