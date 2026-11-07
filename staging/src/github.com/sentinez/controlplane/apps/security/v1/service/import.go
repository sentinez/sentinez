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

package securitysvc

import (
	"context"
	"errors"
	"fmt"

	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// errUnnamedRule is returned for a proxy.yaml rule without a name: rules
// are matched by name on import, so it would be duplicated every time.
var errUnnamedRule = errors.New("rule without a name")

// ImportSecurity stores the security block of a proxy.yaml under owner:
// the core ruleset (one per resource, named name) and the sec_rules and
// limiters, matched by name against the rules owner already has. Stored
// rules missing from sec are left untouched. sec must be normalized
// (sznormalize.NormalizeEdgeSetting) so its rules carry their runtime.
func (srv *SecurityService) ImportSecurity(ctx context.Context,
	owner *typepb.Owner, name string, sec *edgepb.Security) error {

	err := srv.importCoreRuleset(ctx, owner, name, sec.GetCoreRuleset())
	if err != nil {
		return fmt.Errorf("import core ruleset: %w", err)
	}

	if err := srv.importSecRules(ctx, owner, sec.GetSecRules()); err != nil {
		return fmt.Errorf("import sec rules: %w", err)
	}

	if err := srv.importRateLimits(ctx, owner, sec.GetLimiters()); err != nil {
		return fmt.Errorf("import limiters: %w", err)
	}

	return nil
}

// importCoreRuleset updates the core ruleset of the resource of owner, or
// creates it when the resource has none.
func (srv *SecurityService) importCoreRuleset(ctx context.Context,
	owner *typepb.Owner, name string,
	content *corerulesetpb.CoreRuleset) error {

	if content == nil {
		return nil
	}

	stored, _, err := srv.coreRulesetRepo.List(ctx,
		&securitypb.ListCoreRulesetsRequest{
			ResourceIds: []string{owner.GetResourceId()},
		})
	if err != nil {
		return err
	}

	crs := &securitypb.CoreRuleset{Owner: owner, Name: name, Content: content}
	if len(stored) == 0 {
		_, err = srv.coreRulesetRepo.Create(ctx, crs)
		return err
	}

	crs.Id = stored[0].GetId()
	return srv.coreRulesetRepo.Update(ctx, crs)
}

// importSecRules upserts rules by name among the sec rules of owner.
func (srv *SecurityService) importSecRules(ctx context.Context,
	owner *typepb.Owner, rules []*edgepb.SecRule) error {

	stored, _, err := srv.secRuleRepo.List(ctx, &securitypb.ListSecRulesRequest{
		ResourceIds: []string{owner.GetResourceId()},
	})
	if err != nil {
		return err
	}

	models := make([]*securitypb.SecRule, 0, len(rules))
	for _, rule := range rules {
		models = append(models, secRuleFromEdge(owner, rule.GetIngressRuntime()))
	}

	return upsertByName(ctx, stored, models,
		func(ctx context.Context, m *securitypb.SecRule, id string) error {
			if id == "" {
				_, err := srv.secRuleRepo.Create(ctx, m)
				return err
			}
			m.Id = id
			return srv.secRuleRepo.Update(ctx, m)
		})
}

// importRateLimits upserts limiters by name among the rate limits of
// owner.
func (srv *SecurityService) importRateLimits(ctx context.Context,
	owner *typepb.Owner, limiters []*edgepb.RateLimit) error {

	stored, _, err := srv.rateLimitRepo.List(ctx,
		&securitypb.ListRateLimitsRequest{
			ResourceIds: []string{owner.GetResourceId()},
		})
	if err != nil {
		return err
	}

	models := make([]*securitypb.RateLimit, 0, len(limiters))
	for _, limiter := range limiters {
		model, err := rateLimitFromEdge(owner, limiter)
		if err != nil {
			return err
		}
		models = append(models, model)
	}

	return upsertByName(ctx, stored, models,
		func(ctx context.Context, m *securitypb.RateLimit, id string) error {
			if id == "" {
				_, err := srv.rateLimitRepo.Create(ctx, m)
				return err
			}
			m.Id = id
			return srv.rateLimitRepo.Update(ctx, m)
		})
}

// namedModel is a rule matched by name on import.
type namedModel interface {
	GetId() string
	GetName() string
}

// upsertByName writes each of models with write, given the id of the
// stored model of the same name: empty when none has it, to create it.
func upsertByName[M namedModel](ctx context.Context, stored, models []M,
	write func(ctx context.Context, m M, id string) error) error {

	ids := make(map[string]string, len(stored))
	for _, m := range stored {
		ids[m.GetName()] = m.GetId()
	}

	for _, m := range models {
		if m.GetName() == "" {
			return errUnnamedRule
		}

		if err := write(ctx, m, ids[m.GetName()]); err != nil {
			return fmt.Errorf("%q: %w", m.GetName(), err)
		}
	}

	return nil
}

// secRuleFromEdge maps the runtime of a proxy.yaml sec rule to the API
// model.
func secRuleFromEdge(owner *typepb.Owner,
	rule *secrulepb.SecRule) *securitypb.SecRule {

	return &securitypb.SecRule{
		Owner:       owner,
		Name:        rule.GetName(),
		Description: rule.GetDescription(),
		Expr:        rule.GetExpr(),
		Action:      actionName(rule),
		Status:      rule.GetStatus(),
		Priority:    rule.GetPriority(),
	}
}

// rateLimitFromEdge maps a proxy.yaml limiter to the API model.
func rateLimitFromEdge(owner *typepb.Owner,
	limiter *edgepb.RateLimit) (*securitypb.RateLimit, error) {

	rule := limiter.GetIngressRuntime()
	model := &securitypb.RateLimit{
		Owner:       owner,
		Name:        rule.GetName(),
		Description: rule.GetDescription(),
		Expr:        rule.GetExpr(),
		Action:      actionName(rule),
		Status:      rule.GetStatus(),
		Priority:    rule.GetPriority(),
		TimeWindow:  limiter.GetTimeWindow(),
		MaxRequests: limiter.GetMaxRequests(),
		Timeout:     limiter.GetTimeout(),
	}
	if err := validateRateLimit(model); err != nil {
		return nil, fmt.Errorf("%q: %w", model.GetName(), err)
	}

	return model, nil
}

// actionName is the enum name of the action of rule, as the console
// stores it; empty without an action.
func actionName(rule *secrulepb.SecRule) string {
	if rule.GetAction() == nil {
		return ""
	}

	return rule.GetAction().GetType().String()
}
