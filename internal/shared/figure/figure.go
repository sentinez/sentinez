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

package figure

import (
	"fmt"
	"strings"
	"sync"

	"github.com/common-nighthawk/go-figure"
	"github.com/sentinez/sentinez"
	"github.com/sentinez/shared/color"
)

var (
	once sync.Once
)

func INFO(serviceName string, key string, msgs ...string) {
	msgs = append([]string{sentinez.Code + " " + sentinez.Version}, msgs...)
	gts := color.Green.Add(">")
	msg := strings.Join(msgs, "\n"+gts+" ")

	once.Do(func() {
		fmt.Print(Gen(serviceName, key) + gts + " " + msg + "\n\n")
	})
}

// Gen generates the ASCII art of the project.
func Gen(header string, footer string) string {
	fig := figure.NewFigure(strings.ToLower(sentinez.Code), "speed", true)
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
