// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"strings"

	"github.com/go-delve/delve/service/api"
)

type debugMessage struct {
	format string
	args   []string
}

// parseDebugMessage parses a log message according to the DAP spec:
//
//	"Expressions within {} are interpolated."
func parseDebugMessage(msg string) (*debugMessage, error) {
	// Note: All braces *must* come in pairs, even those within an
	// expression to be interpolated.
	// TODO(suzmue): support individual braces in string values in
	// eval expressions.
	var args []string
	var isArg bool
	var formatSlice, argSlice []rune

	braceCount := 0
	for _, r := range msg {
		if isArg { //nolint:nestif
			switch r {
			case '}':
				if braceCount--; braceCount == 0 {
					argStr := strings.TrimSpace(string(argSlice))
					if len(argStr) == 0 {
						return nil, fmt.Errorf("empty evaluation string")
					}
					args = append(args, argStr)
					formatSlice = append(formatSlice, '%', 's')
					isArg = false
					continue
				}
			case '{':
				braceCount++
			}
			argSlice = append(argSlice, r)
		} else {
			switch r {
			case '}':
				return nil, fmt.Errorf("invalid log point format, unexpected '}'")
			case '{':
				if braceCount++; braceCount == 1 {
					isArg, argSlice = true, []rune{}
					continue
				}
			}
			formatSlice = append(formatSlice, r)
		}
	}
	if isArg || len(formatSlice) == 0 {
		return nil, fmt.Errorf("invalid debug message format")
	}

	return &debugMessage{
		format: string(formatSlice),
		args:   args,
	}, nil
}

func (msg *debugMessage) evaluate(vars []api.Variable) string {
	evaluated := []any{}

	varMap := map[string]string{}
	for _, v := range vars {
		varMap[v.Name] = v.Value
	}

	for _, arg := range msg.args {
		if v, ok := varMap[arg]; ok {
			evaluated = append(evaluated, v)
		} else {
			evaluated = append(evaluated, "<missing>")
		}
	}

	return fmt.Sprintf(msg.format, evaluated...)
}
