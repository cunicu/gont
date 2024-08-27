// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package vm

import (
	"fmt"
	"strings"

	g "cunicu.li/gont/v2/pkg"
	co "cunicu.li/gont/v2/pkg/options/cmd"
)

type Architecture string

func (a Architecture) ApplyQEmuVM(vm *g.QEmuVM) {
	vm.Arch = string(a)
}

func Option(name string, opts ...any) co.Arguments {
	kvs := []string{}

	for _, opt := range opts {
		switch opt := opt.(type) {
		case map[string]any:
			for key, value := range opt {
				kvs = append(kvs, fmt.Sprintf("%s=%v", key, value))
			}
		case map[string]string:
			for key, value := range opt {
				kvs = append(kvs, fmt.Sprintf("%s=%s", key, value))
			}
		default:
			kvs = append(kvs, fmt.Sprint(opt))
		}
	}

	args := co.Arguments{"-" + name}

	if len(kvs) > 0 {
		args = append(args, strings.Join(kvs, ","))
	}

	return args
}

type CloudInitUserData map[string]any

func (c CloudInitUserData) ApplyQEmuVM(vm *g.QEmuVM) {
	vm.CloudInit.UserData = c
}

type CloudInitMetaData map[string]any

func (c CloudInitMetaData) ApplyQEmuVM(vm *g.QEmuVM) {
	vm.CloudInit.MetaData = c
}

// Shortcuts

func Memory(megs int) co.Arguments {
	return Option("m", map[string]any{"size": megs})
}

//nolint:gochecknoglobals
var NoGraphic = Option("nographic")

func Machine(typ string, opts ...any) co.Arguments {
	args := []any{typ}
	args = append(args, opts...)

	return Option("machine", args...)
}

func CPU(model string) co.Arguments {
	return Option("cpu", model)
}

func Device(driver string, props map[string]any) co.Arguments {
	return Option("device", driver, props)
}

func NetDev(typ string, props map[string]any) co.Arguments {
	return Option("netdev", typ, props)
}

func Drive(props map[string]any) co.Arguments {
	return Option("drive", props)
}

func VNC(display int, opts ...any) co.Arguments {
	args := []any{fmt.Sprintf(":%d", display)}
	args = append(args, opts...)

	return Option("vnc", args...)
}
