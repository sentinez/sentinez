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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	"github.com/sentinez/sentinez/pkg/version"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
)

var (
	once sync.Once
)

// flags global variable
var flags = &common.Flag{
	EnvMode:  "dev",
	LogLevel: "debug",
}

func info(meta *common.SntzMeta) string {
	service := strings.Replace(meta.GetServiceName(), "_", " // ", 1)
	return version.FigureGen(service, meta.GetServiceKey())
}

// Parse flag args
func Parse(meta *common.SntzMeta) {
	once.Do(func() {

		pflag.StringVarP(&flags.EnvMode, "mode", "m",
			flags.GetEnvMode(), "run mode (dev|prod|sandbox)")

		pflag.StringVar(&flags.LogLevel, "log-level",
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

func Get() *common.Flag {
	return flags
}

// Validate used to validate flags
func Validate(flag proto.Message) error {
	if err := protobuf.Validate(flag); err != nil {
		return err
	}

	return nil
}
