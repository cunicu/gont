// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"fmt"
	"math/rand"
	"path/filepath"
)

func (n *BaseNode) Run(cmd string, args ...any) (*Cmd, error) {
	c := n.Command(cmd, args...)
	return c, c.Run()
}

func (n *BaseNode) Start(cmd string, args ...any) (*Cmd, error) {
	c := n.Command(cmd, args...)
	return c, c.Start()
}

func (n *BaseNode) StartGo(script string, args ...any) (*Cmd, error) {
	tmp := filepath.Join(n.network.VarPath, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))

	c := n.network.HostNode.Command("go", "build", "-o", tmp, script)

	if out, err := c.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to compile Go code: %w\n%s", err, string(out))
	}

	return n.Start(tmp, args...)
}

func (n *BaseNode) RunGo(script string, args ...any) (*Cmd, error) {
	tmp := filepath.Join(n.network.TmpPath, fmt.Sprintf("go-build-%d", rand.Intn(1<<16)))

	if _, err := n.network.HostNode.Run("go", "build", "-o", tmp, script); err != nil {
		return nil, fmt.Errorf("failed to compile Go code: %w", err)
	}

	return n.Run(tmp, args...)
}
