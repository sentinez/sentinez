// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secure

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/variable"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/shared/bytesconv"
	"google.golang.org/protobuf/types/known/structpb"
)

// Action params keys (Action.params).
const (
	// _paramLocation is the REDIRECT target and the ROUTE_TO location.
	_paramLocation = "location"
	// _paramStatus is the REDIRECT status code, 302 by default.
	_paramStatus = "status"
)

// _maxActions sizes the stack buffer of the non-terminal actions a request
// matches; more spill to the heap.
const _maxActions = 8

// evaluate runs chain in priority order. It returns the terminal action
// that ends the request, ranked block > challenge > redirect with ties to
// the earlier rule, and appends the other matched actions to acts. A block
// stops the scan; a challenge or redirect does not, so a later block still
// applies to a client holding a pass.
func evaluate(
	ctx szhttp.Context,
	chain []secrules.Entry,
	mr *rulepb.MatchedRules,
	acts []*rulepb.Action,
) (*rulepb.Action, []*rulepb.Action) {
	var terminal *rulepb.Action

	for _, e := range chain {
		if !e.Eval(ctx, mr) {
			continue
		}

		action := e.Rule.GetAction()
		rank := terminalRank(action.GetType())
		switch {
		case action.GetType() == rulepb.ActionType_ACTION_TYPE_BLOCK:
			return action, acts
		case rank == 0:
			acts = append(acts, action)
		case rank > terminalRank(terminal.GetType()):
			terminal = action
		}
	}

	return terminal, acts
}

// terminalRank orders the actions that end a request; 0 for the others.
func terminalRank(t rulepb.ActionType) int {
	switch t {
	case rulepb.ActionType_ACTION_TYPE_BLOCK:
		return 3
	case rulepb.ActionType_ACTION_TYPE_CHALLENGE:
		return 2
	case rulepb.ActionType_ACTION_TYPE_REDIRECT:
		return 1
	default:
		return 0
	}
}

// applyActions applies the non-terminal actions in reverse, so the rule
// with the higher priority has the last word on a header, tag or route.
// LOG needs nothing beyond the matched event.
func applyActions(ctx szhttp.Context, acts []*rulepb.Action) {
	for i := len(acts) - 1; i >= 0; i-- {
		switch acts[i].GetType() {
		case rulepb.ActionType_ACTION_TYPE_MODIFY_HEADER:
			modifyHeaders(ctx, acts[i].GetParams())
		case rulepb.ActionType_ACTION_TYPE_SET_TAG:
			setTags(ctx, acts[i].GetParams())
		case rulepb.ActionType_ACTION_TYPE_ROUTE_TO:
			if loc := paramString(acts[i], _paramLocation); loc != "" {
				ctx.X().RouteLocation = loc
			}
		}
	}
}

// loggedAction is the action a match is logged with: the terminal one, else
// the highest priority one.
func loggedAction(
	terminal *rulepb.Action, acts []*rulepb.Action,
) *rulepb.Action {
	if terminal != nil || len(acts) == 0 {
		return terminal
	}

	return acts[0]
}

// modifyHeaders sets each param as a request header sent upstream.
func modifyHeaders(ctx szhttp.Context, params *structpb.Struct) {
	for k, v := range params.GetFields() {
		if !variable.IsValidHeaderKey(k) {
			continue
		}

		ctx.SetHeader(bytesconv.S2b(k), bytesconv.S2b(valueString(v)))
	}
}

// setTags records each param as a request tag.
func setTags(ctx szhttp.Context, params *structpb.Struct) {
	fields := params.GetFields()
	if len(fields) == 0 {
		return
	}

	x := ctx.X()
	if x.Tags == nil {
		x.Tags = make(map[string]string, len(fields))
	}

	for k, v := range fields {
		x.Tags[k] = valueString(v)
	}
}

// redirect answers with the action's location and status. It reports
// false, writing nothing, when the location is missing or unsafe.
func redirect(ctx szhttp.Context, action *rulepb.Action) (bool, error) {
	location := paramString(action, _paramLocation)
	if location == "" || strings.ContainsAny(location, "\r\n") {
		return false, nil
	}

	ctx.SetResponseHeader(szbytestr.HeaderLocation, []byte(location))

	return true, ctx.String(redirectStatus(action), nil)
}

// redirectStatus returns the status param when it is a redirect status,
// else 302.
func redirectStatus(action *rulepb.Action) int {
	status, err := strconv.Atoi(paramString(action, _paramStatus))
	if err != nil {
		return http.StatusFound
	}

	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return status
	default:
		return http.StatusFound
	}
}

func paramString(action *rulepb.Action, key string) string {
	return valueString(action.GetParams().GetFields()[key])
}

// valueString formats a param value: YAML numbers and booleans arrive as
// such, not as strings.
func valueString(v *structpb.Value) string {
	switch k := v.GetKind().(type) {
	case *structpb.Value_StringValue:
		return k.StringValue
	case *structpb.Value_NumberValue:
		return strconv.FormatFloat(k.NumberValue, 'f', -1, 64)
	case *structpb.Value_BoolValue:
		return strconv.FormatBool(k.BoolValue)
	default:
		return ""
	}
}
