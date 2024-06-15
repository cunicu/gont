// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"os"
	"os/exec"
)

type GoBuildFlags []string

type GoBuildFlagsOption interface {
	ApplyGoBuildFlags(*GoBuildFlags)
}

func (n *NamespaceNode) Run(cmd string, args ...any) (*Cmd, error) {
	c := n.Command(cmd, args...)
	return c, c.Run()
}

func (n *NamespaceNode) Start(cmd string, args ...any) (*Cmd, error) {
	c := n.Command(cmd, args...)
	return c, c.Start()
}

func (n *NamespaceNode) StartGo(fileOrPkg string, args ...any) (*Cmd, error) {
	bin, err := n.BuildGo(fileOrPkg, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to build: %w", err)
	}

	return n.Start(bin.Name(), args...)
}

func (n *NamespaceNode) RunGo(fileOrPkg string, args ...any) (*Cmd, error) {
	bin, err := n.BuildGo(fileOrPkg, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to build: %w", err)
	}

	return n.Run(bin.Name(), args...)
}

func (n *NamespaceNode) BuildGo(fileOrPkg string, args ...any) (*os.File, error) {
	if err := os.MkdirAll(n.network.TmpPath, 0o644); err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	tmp, err := os.CreateTemp(n.network.TmpPath, "go-build-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	flags := GoBuildFlags{}
	for _, arg := range args {
		if opt, ok := arg.(GoBuildFlagsOption); ok {
			opt.ApplyGoBuildFlags(&flags)
		}
	}

	bArgs := []string{"build"}
	for _, flag := range flags {
		bArgs = append(bArgs, flag)
	}
	bArgs = append(bArgs, "-o", tmp.Name(), fileOrPkg)

	c := exec.Command("go", bArgs...)
	if out, err := c.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to compile Go code: %w\n%s", err, string(out))
	}

	return tmp, nil
}
