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

// Command szservicetoken signs the token a service (eg an edge) presents to
// the control plane, with SENZ_SERVICE_KEY. Run it where the key lives and
// give the token to the service as SENZ_SERVICE_TOKEN.
//
//	SENZ_SERVICE_KEY=... szservicetoken -name edge-1 -ttl 720h
//
// The name must be the common name of the certificate of the service.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/pkg/szconfig"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "szservicetoken:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("szservicetoken", flag.ContinueOnError)
	name := fs.String("name", "", "service name, the CN of its certificate")
	ttl := fs.Duration("ttl", 30*24*time.Hour, "validity of the token")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("-name is required")
	}

	if *ttl <= 0 {
		return fmt.Errorf("-ttl must be positive")
	}

	szconfig.SetEnv(settingpb.Senz_SENZ_SERVICE_KEY,
		os.Getenv(settingpb.Senz_SENZ_SERVICE_KEY.String()))

	token, err := crypto.ServiceTokenGenerator(*name, *ttl)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(out, token)

	return err
}
