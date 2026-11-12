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

package connection

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TLS the PEM files of a service of the internal mTLS: the internal CA
// and the certificate and key this service presents.
type TLS struct {
	CAFile   string
	CertFile string
	KeyFile  string
}

// TLSFromConfig the files of SENZ_TLS_CA, SENZ_TLS_CERT and SENZ_TLS_KEY.
func TLSFromConfig(conf *settingpb.Config) TLS {
	return TLS{
		CAFile:   conf.Get(settingpb.Senz_SENZ_TLS_CA),
		CertFile: conf.Get(settingpb.Senz_SENZ_TLS_CERT),
		KeyFile:  conf.Get(settingpb.Senz_SENZ_TLS_KEY),
	}
}

// Enabled reports whether any file is set; unset means plaintext.
func (t TLS) Enabled() bool {
	return t.CAFile != "" || t.CertFile != "" || t.KeyFile != ""
}

// ServerCredentials the credentials of a server that requires a client
// certificate signed by the internal CA.
func (t TLS) ServerCredentials() (credentials.TransportCredentials, error) {
	conf, err := t.config()
	if err != nil {
		return nil, err
	}

	conf.ClientAuth = tls.RequireAndVerifyClientCert
	conf.ClientCAs = conf.RootCAs
	conf.RootCAs = nil

	return credentials.NewTLS(conf), nil
}

// ClientCredentials the credentials of a client that presents its
// certificate and trusts only servers of the internal CA. serverName is
// the name the server certificate is verified against.
func (t TLS) ClientCredentials(
	serverName string) (credentials.TransportCredentials, error) {

	conf, err := t.config()
	if err != nil {
		return nil, err
	}

	conf.ServerName = serverName

	return credentials.NewTLS(conf), nil
}

// config the TLS 1.3 configuration holding the certificate and the CA.
func (t TLS) config() (*tls.Config, error) {
	if t.CAFile == "" || t.CertFile == "" || t.KeyFile == "" {
		return nil, errors.New("tls: ca, cert and key are all required")
	}

	pair, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: load key pair: %w", err)
	}

	pem, err := os.ReadFile(t.CAFile)
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

// DialOptions the options of a client conn: the internal mTLS when t is
// enabled, else plaintext.
func (t TLS) DialOptions(serverName string) ([]grpc.DialOption, error) {
	if !t.Enabled() {
		return []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}, nil
	}

	creds, err := t.ClientCredentials(serverName)
	if err != nil {
		return nil, err
	}

	return []grpc.DialOption{grpc.WithTransportCredentials(creds)}, nil
}
