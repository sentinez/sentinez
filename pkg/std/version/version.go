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

// Package version provides the version of the package.
package version

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/common-nighthawk/go-figure"
)

var (

	// Package is filled at linking time
	Package = "github.com/sentinez/sentinez"

	// Version holds the complete version number. Filled in at linking time.
	Version = "v0.0.1-beta"

	// GoVersion is Go tree's version.
	GoVersion = runtime.Version()

	// Name is the full name of the project.
	Name = "SENTINEZ"

	// Code is the code of the project.
	Code = "STNZ"
)

// ASCII prints the ASCII art of the project.
func ASCII(header string, footer string) {
	fmt.Print(FigureGen(header, footer))
}

// FigureGen generates the ASCII art of the project.
func FigureGen(header string, footer string) string {
	fig := figure.NewFigure("setnz", "speed", true)
	figureLines := strings.Split(fig.String(), "\n")
	sideText := []string{
		"",
		"",
		header,
		"------------",
		footer,
	}

	var v string
	for i, line := range figureLines {
		side := ""
		if i < len(sideText) {
			side = sideText[i]
		}
		v += fmt.Sprintf("%-40s %s\n", line, side)
	}

	return v
}
