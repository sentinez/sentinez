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

// Package flags provide flag variable props
package flags

import (
	"fmt"
	"os"
	"sync"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/pkg/client/names"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	"github.com/sentinez/sentinez/pkg/std/version"
	"google.golang.org/protobuf/proto"

	"github.com/spf13/pflag"
)

var (
	once sync.Once
)

var (
	// ascii art use in console with --help option
	asciiConsole = version.FigureGen("SENTINEZ // CONSOLE", "sentinez.console")
)

// flags global variable
var flags = &common.Flag{
	Name:      "sntz.server.default",
	Mode:      "dev",
	LogLevel:  "debug",
	ConsulUrl: "http://localhost:8500",
}

// Parse flag args
func Parse() *common.Flag {
	once.Do(func() {
		pflag.StringVarP(&flags.Mode, "mode", "m",
			flags.GetMode(), "run mode (dev|prod|sandbox)")

		pflag.StringVar(&flags.LogLevel, "log-level",
			flags.GetLogLevel(), "log level (debug|info|warn|error)")

		pflag.StringVar(&flags.ConsulUrl, "consul-url",
			flags.GetConsulUrl(), "consul url")

		pflag.Usage = func() {
			fmt.Print(asciiConsole)
			fmt.Println("Usage: <service> [Flags]")
			pflag.PrintDefaults()
			os.Exit(0)
		}

		pflag.Parse()
	})

	return flags
}

// SetConsole set default flag values
func SetConsole(content string, name names.Namespace, mode string) {
	flags.Name = name.String()
	flags.Mode = mode
	asciiConsole = content
}

// Validate used to validate flags
func Validate(flag proto.Message) error {
	if err := protobuf.Validate(flag); err != nil {
		fmt.Print(asciiConsole)
		return err
	}

	return nil
}
