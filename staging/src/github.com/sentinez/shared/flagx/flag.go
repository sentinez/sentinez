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

package flagx

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sentinez/core/console"
	flagpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/flag/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/protobuf"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
)

var (
	once sync.Once
)

// flags global variable
var flags = &flagpb.Flag{
	EnvMode:  "dev",
	LogLevel: "debug",
}

func info(meta *typepb.XMeta) string {
	service := strings.Replace(meta.GetServiceName(), "_", " // ", 1)
	return console.GenFigure(service, meta.GetServiceKey())
}

// Parse flag args
func Parse(meta *typepb.XMeta) {
	once.Do(func() {

		pflag.StringVarP(&flags.EnvMode, "mode", "m",
			flags.GetEnvMode(), "run mode (dev|prod|sandbox)")

		pflag.StringVar(&flags.LogLevel, "log_level",
			flags.GetLogLevel(), "log level (debug|info|warn|error)")

		pflag.Usage = func() {
			fmt.Print(info(meta))
			fmt.Println("Usage: <service> [Flags]")
			pflag.PrintDefaults()
			os.Exit(0)
		}

		pflag.Parse()
	})
}

func Get() *flagpb.Flag {
	return flags
}

// Validate used to validate flags
func Validate(flag proto.Message) error {
	if err := protobuf.Validate(flag); err != nil {
		return err
	}

	return nil
}
