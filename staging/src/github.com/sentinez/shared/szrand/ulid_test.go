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

package szrand

import (
	"testing"

	"buf.build/go/protovalidate"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/shared/szprefix"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// eventMsgs builds every event message whose ids NewULID generates.
func eventMsgs() []proto.Message {
	reqId := NewULID(szprefix.Request)

	return []proto.Message{
		&httppb.Request{Id: reqId},
		&httppb.RequestEvent{Id: reqId},
		&httppb.Event{Id: reqId},
		&secrulepb.RuleEvent{
			Id:        NewULID(szprefix.SecRuleEvent),
			RequestId: reqId,
		},
		&corerulesetpb.CoreRuleEvent{
			Id:        NewULID(szprefix.CoreRulesetEvent),
			RequestId: reqId,
		},
	}
}

// setZone makes the zone id prefix zone ("" for none) for the test.
func setZone(t *testing.T, zone string) {
	t.Helper()

	prev := _zone
	t.Cleanup(func() { _zone = prev })
	t.Setenv("SENZ_ZONE", zone)
	_zone = loadZone
}

func TestULIDMatchesProtoPattern(t *testing.T) {
	tests := []struct {
		name     string
		giveZone string
	}{
		{name: "no zone", giveZone: ""},
		{name: "zone", giveZone: "SGN"},
		{name: "dashed zone", giveZone: "ap-southeast-1a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setZone(t, tt.giveZone)

			for _, msg := range eventMsgs() {
				assert.NoError(t, protovalidate.Validate(msg), msg)
			}
		})
	}
}

func TestEventIdRejectsWrongFormat(t *testing.T) {
	tests := []struct {
		name   string
		giveId string
	}{
		{name: "xid", giveId: "req_csmqt6lv0bqs73ffrl4g"},
		{name: "wrong prefix", giveId: NewULID(szprefix.Event)},
		{name: "no separator", giveId: "req01ARZ3NDEKTSV4RRFFQ69G5FAV"},
		{name: "empty", giveId: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := protovalidate.Validate(&httppb.Request{Id: tt.giveId})
			assert.Error(t, err)
		})
	}
}
