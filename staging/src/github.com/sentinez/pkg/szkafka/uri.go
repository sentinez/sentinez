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

package szkafka

import (
	"crypto/tls"

	"github.com/sentinez/pkg/sztls"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Endpoint is where a client connects.
type Endpoint struct {
	// Brokers are the seed brokers, e.g. ["localhost:9094"].
	Brokers []string
	// TLS is nil for plaintext.
	TLS *tls.Config
}

// ParseURI returns the endpoint of uri, "kafka://host:port[,host:port...]"
// or "kafkas://..." for TLS: the mTLS of files when they are set, else
// TLS trusting the system roots.
func ParseURI(uri string, files sztls.Files) (Endpoint, error) {
	ep, err := sztls.ParseURI(uri, "kafka", "kafkas", files)
	if err != nil {
		return Endpoint{}, err
	}

	return Endpoint{Brokers: ep.Hosts, TLS: ep.TLS}, nil
}

// dialOpts the options dialing with tlsConf, none for plaintext.
func dialOpts(tlsConf *tls.Config) []kgo.Opt {
	if tlsConf == nil {
		return nil
	}

	return []kgo.Opt{kgo.DialTLSConfig(tlsConf)}
}
