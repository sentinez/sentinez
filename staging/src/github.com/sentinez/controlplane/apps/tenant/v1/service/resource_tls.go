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

package tenantsvc

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/sentinez/core/szcrypto/certs"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

const (
	_statusActive  = typepb.Status_STATUS_ACTIVE
	_statusPending = typepb.Status_STATUS_PENDING

	// _usageUpstreamCa is the usage of the CAs the edge verifies the
	// upstreams with.
	_usageUpstreamCa = tenantpb.CaCertificateUsage_CA_CERTIFICATE_USAGE_UPSTREAM
)

// _tlsPaths the fields of a resource its TLS depends on.
var _tlsPaths = []string{
	tenantpb.Resource_Status, tenantpb.Resource_Domain,
	tenantpb.Resource_TlsMode,
}

// touchesTls reports whether an update of paths may change whether the
// TLS of the resource is valid.
func touchesTls(paths []string) bool {
	return slices.ContainsFunc(paths, func(p string) bool {
		return slices.Contains(_tlsPaths, p)
	})
}

// wantsActive reports whether rsc is asked to be active: active, or
// pending until its certificates fit its TLS mode.
func wantsActive(rsc *tenantpb.Resource) bool {
	return rsc.GetStatus() == _statusActive ||
		rsc.GetStatus() == _statusPending
}

// settleStatus sets rsc, when asked to be active, active if the edge can
// serve the TLS of its domain in its mode (see tlsReady), else pending.
// It reports whether the status changed.
func (svc *Service) settleStatus(ctx context.Context,
	rsc *tenantpb.Resource) (bool, error) {
	if !wantsActive(rsc) {
		return false, nil
	}

	ready, err := svc.tlsReady(ctx, rsc)
	if err != nil {
		return false, err
	}

	was := rsc.GetStatus()
	rsc.Status = _statusPending
	if ready {
		rsc.Status = _statusActive
	}

	return rsc.GetStatus() != was, nil
}

// settleUpdate settles the status of rsc, updated under paths, when they
// touch its TLS: the paths, with the status when it changed.
func (svc *Service) settleUpdate(ctx context.Context,
	rsc *tenantpb.Resource, paths []string) ([]string, error) {
	if !touchesTls(paths) {
		return paths, nil
	}

	changed, err := svc.settleStatus(ctx, rsc)
	if err != nil {
		return nil, err
	}

	if !changed || slices.Contains(paths, tenantpb.Resource_Status) {
		return paths, nil
	}

	return append(slices.Clip(paths), tenantpb.Resource_Status), nil
}

// tlsReady reports whether the certificates of rsc are what its TLS mode
// needs, valid now:
//   - passthrough: none, the edge forwards the TLS untouched;
//   - terminate: a server certificate for the domain, with its key;
//   - re-encrypt, and unspecified (an HTTPS upstream re-encrypts): the
//     server certificate, then, when the resource has upstream CAs or
//     upstream client certificates, one of each still valid (the client
//     one with its key), as the edge dials the upstreams with them.
//
// A resource not stored yet has no certificate.
func (svc *Service) tlsReady(ctx context.Context,
	rsc *tenantpb.Resource) (bool, error) {
	mode := rsc.GetTlsMode()
	if mode == edgepb.TlsMode_TLS_MODE_PASSTHROUGH {
		return true, nil
	}

	if rsc.GetId() == "" {
		return false, nil
	}

	resp, err := svc.cacert.List(ctx, &tenantpb.ListCaCertificateRequest{
		ResourceIds: []string{rsc.GetId()},
	})
	if err != nil {
		return false, fmt.Errorf("cacert.List: %w", err)
	}

	byUsage := certificatesByUsage(resp.GetCaCertificates())
	now := svc.now()

	if !anyValid(byUsage[_usageServer], now, serves(rsc.GetDomain())) {
		return false, nil
	}

	if mode == edgepb.TlsMode_TLS_MODE_TERMINATE {
		return true, nil
	}

	return validOrNone(byUsage[_usageUpstreamCa], now, isAny) &&
		validOrNone(byUsage[_usageUpstreamClient], now, hasKey), nil
}

// certificatesByUsage groups certs by their usage.
func certificatesByUsage(certs []*tenantpb.CaCertificate,
) map[tenantpb.CaCertificateUsage][]*tenantpb.CaCertificate {
	byUsage := make(
		map[tenantpb.CaCertificateUsage][]*tenantpb.CaCertificate)
	for _, cert := range certs {
		byUsage[cert.GetUsage()] = append(byUsage[cert.GetUsage()], cert)
	}

	return byUsage
}

// validOrNone reports whether certs is empty, or one of them is valid.
func validOrNone(certs []*tenantpb.CaCertificate, now time.Time,
	ok func(*tenantpb.CaCertificate) bool) bool {
	return len(certs) == 0 || anyValid(certs, now, ok)
}

// anyValid reports whether one of certs is ok and valid at now.
func anyValid(certs []*tenantpb.CaCertificate, now time.Time,
	ok func(*tenantpb.CaCertificate) bool) bool {
	return slices.ContainsFunc(certs, func(c *tenantpb.CaCertificate) bool {
		return !now.Before(c.GetNotBefore().AsTime()) &&
			!now.After(c.GetNotAfter().AsTime()) && ok(c)
	})
}

func isAny(*tenantpb.CaCertificate) bool { return true }

func hasKey(cert *tenantpb.CaCertificate) bool {
	return cert.GetHasPrivateKey()
}

// serves reports whether the edge can serve a certificate for domain: it
// holds its private key and names the domain.
func serves(domain string) func(*tenantpb.CaCertificate) bool {
	return func(cert *tenantpb.CaCertificate) bool {
		leaf := certs.Parse(cert.GetPem())

		return cert.GetHasPrivateKey() && leaf != nil &&
			leaf.VerifyHostname(domain) == nil
	}
}

// recheckStatus settles the status of rsc after its certificates changed:
// a pending resource becomes active once they fit its TLS mode, an active
// one pending once they no longer do.
func (svc *Service) recheckStatus(ctx context.Context,
	rsc *tenantpb.Resource) error {
	changed, err := svc.settleStatus(ctx, rsc)
	if err != nil || !changed {
		return err
	}

	err = svc.resource.Update(ctx, rsc, []string{tenantpb.Resource_Status})
	if err != nil {
		return fmt.Errorf("resource.Update: %w", err)
	}

	return nil
}
