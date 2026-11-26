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

// Package sztls loads the PEM files of the internal mTLS (SZ_TLS_CA,
// SZ_TLS_CERT and SZ_TLS_KEY) into the tls.Config of a server or a client.
package sztls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// Files the PEM files of a service of the internal mTLS: the internal CA
// and the certificate and key this service presents.
type Files struct {
	CAFile   string
	CertFile string
	KeyFile  string
}

// FilesOf the files of SZ_TLS_CA, SZ_TLS_CERT and SZ_TLS_KEY.
func FilesOf(conf *settingpb.Config) Files {
	return Files{
		CAFile:   conf.Get(settingpb.Sz_SZ_TLS_CA),
		CertFile: conf.Get(settingpb.Sz_SZ_TLS_CERT),
		KeyFile:  conf.Get(settingpb.Sz_SZ_TLS_KEY),
	}
}

// Enabled reports whether any file is set.
func (f Files) Enabled() bool {
	return f.CAFile != "" || f.CertFile != "" || f.KeyFile != ""
}

// Config the TLS 1.3 config holding the certificate, and the CA as
// RootCAs; all three files are required.
func (f Files) Config() (*tls.Config, error) {
	if f.CAFile == "" || f.CertFile == "" || f.KeyFile == "" {
		return nil, errors.New("tls: ca, cert and key are all required")
	}

	pair, err := tls.LoadX509KeyPair(f.CertFile, f.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: load key pair: %w", err)
	}

	pem, err := os.ReadFile(f.CAFile)
	if err != nil {
		return nil, fmt.Errorf("tls: read ca: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("tls: ca holds no certificate")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{pair},
		RootCAs:      pool,
	}, nil
}

// ClientConfig the config of a client: the mTLS of Config when f is
// enabled, else TLS 1.2+ trusting the system roots. The server name is
// left for the dialer to set from the address it dials.
func (f Files) ClientConfig() (*tls.Config, error) {
	if !f.Enabled() {
		return &tls.Config{MinVersion: tls.VersionTLS12}, nil
	}

	return f.Config()
}
