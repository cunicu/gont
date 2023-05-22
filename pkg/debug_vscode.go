// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Limit debugger support to platforms which are supported by Delve
// See: https://github.com/go-delve/delve/blob/master/pkg/proc/native/support_sentinel_linux.go
//go:build linux && (amd64 || arm64 || 386)

package gont

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
)

type vscTasksConfig struct {
	Version string    `json:"version"`
	Tasks   []vscTask `json:"tasks"`
}

type vscTask struct {
	Label          string             `json:"label"`
	Type           string             `json:"type"`
	Command        string             `json:"command"`
	Presentation   *vscPresentation   `json:"presentation,omitempty"`
	Group          string             `json:"group,omitempty"`
	IsBackground   bool               `json:"isBackground"`
	ProblemMatcher *vscProblemMatcher `json:"problemMatcher,omitempty"`
}

type vscProblemMatcher struct {
	Owner      string         `json:"owner,omitempty"`
	Pattern    *vscPattern    `json:"pattern,omitempty"`
	Background *vscBackground `json:"background,omitempty"`
}

type vscBackground struct {
	ActiveOnStart bool   `json:"activeOnStart,omitempty"`
	BeginsPattern string `json:"beginsPattern,omitempty"`
	EndsPattern   string `json:"endsPattern,omitempty"`
}

type vscPattern struct {
	Regexp string `json:"regexp,omitempty"`
}

type vscLaunchConfig struct {
	Version        string             `json:"version,omitempty"`
	Configurations []vscConfiguration `json:"configurations,omitempty"`
	Compounds      []vscCompound      `json:"compounds,omitempty"`
}

type vscPresentation struct {
	Hidden bool   `json:"hidden,omitempty"`
	Reveal string `json:"reveal,omitempty"`
	Panel  string `json:"panel,omitempty"`
}

type vscConfiguration struct {
	Name         string `json:"name,omitempty"`
	Type         string `json:"type,omitempty"`
	StopOnEntry  bool   `json:"stopOnEntry,omitempty"`
	DebugAdapter string `json:"debugAdapter,omitempty"`
	Request      string `json:"request,omitempty"`
	Port         int    `json:"port,omitempty"`
	Host         string `json:"host,omitempty"`
	RemotePath   string `json:"remotePath,omitempty"`
	Mode         string `json:"mode,omitempty"`
}

type vscCompound struct {
	Name           string   `json:"name,omitempty"`
	StopAll        bool     `json:"stopAll,omitempty"`
	Configurations []string `json:"configurations,omitempty"`
	PreLaunchTask  string   `json:"preLaunchTask,omitempty"`
}

// WriteVSCodeConfigs generates Visual Studio Code Launch and Task configuration files
// (tasks.json, launch.json) in the given workspace directory.
// The launch configuration is dynamically generated from the current active Delve debugger
// instances
// If an empty dir is passed, we attempt to find the workspace directory by searching for a
// parent directory which contains either a .vscode, go.mod or .git
func (d *Debugger) WriteVSCodeConfigs(dir string, stopOnEntry bool) error {
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working dir: %w", err)
		}

		var ok bool
		if dir, ok = findWorkspaceDir(wd); !ok {
			return errors.New("failed to find workspace directory")
		}
	}

	if err := os.MkdirAll(path.Join(dir, ".vscode"), 0o755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	for filename, writer := range map[string]func(io.Writer) error{
		"launch.json": d.writeVSCodeLaunchConfig,
		"tasks.json":  d.writeVSCodeTasksConfig,
	} {
		f, err := os.OpenFile(path.Join(dir, ".vscode", filename), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o666)
		if err != nil {
			return err
		}

		if err := writer(f); err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	zap.L().Info("Ready to attach debuggers")

	return nil
}

func (d *Debugger) writeVSCodeTasksConfig(wr io.Writer) error {
	cfg := &vscTasksConfig{
		Version: "0.2.0",
		Tasks: []vscTask{
			{
				Label:   "Start tracing",
				Type:    "shell",
				Command: "wireshark -Xlua_script:${workspaceFolder}/dissector/dissector.lua",
			},
			{
				Label:   "Start debugging",
				Type:    "shell",
				Command: "sudo go test -v -run TestDebug ./pkg/",
				Presentation: &vscPresentation{
					Reveal: "always",
					Panel:  "new",
				},
				IsBackground: true,
				ProblemMatcher: &vscProblemMatcher{
					Owner: "custom",
					Pattern: &vscPattern{
						Regexp: "__________", // Shall never match
					},
					Background: &vscBackground{
						ActiveOnStart: true,
						BeginsPattern: ".*",
						EndsPattern:   ".*Ready to attach debuggers.*",
					},
				},
			},
		},
	}

	enc := json.NewEncoder(wr)
	enc.SetIndent("", "  ")

	return enc.Encode(cfg)
}

func (d *Debugger) writeVSCodeLaunchConfig(wr io.Writer) error {
	lcfgs := []vscConfiguration{}
	lcfgnames := []string{}

	for _, i := range d.instances {
		if i.listenAddr == nil {
			continue
		}

		lcfg := vscConfiguration{
			Name:         fmt.Sprintf("Debug %s (%d)", strings.Join(i.cmd.Args, " "), i.cmd.Process.Pid),
			Type:         "go",
			Request:      "attach",
			DebugAdapter: "dlv-dap",
			Mode:         "remote",
			StopOnEntry:  d.BreakOnEntry,
			RemotePath:   "${workspaceFolder}",
			Port:         i.listenAddr.Port,
		}

		if i.listenAddr.IP == nil || i.listenAddr.IP.IsUnspecified() {
			lcfg.Host = net.IPv6loopback.String()
		} else {
			lcfg.Host = i.listenAddr.IP.String()
		}

		lcfgs = append(lcfgs, lcfg)
		lcfgnames = append(lcfgnames, lcfg.Name)
	}

	cfg := &vscLaunchConfig{
		Version:        "0.2.0",
		Configurations: lcfgs,
		Compounds: []vscCompound{
			{
				Name:           "Debug all processes",
				Configurations: lcfgnames,
				StopAll:        true,
				PreLaunchTask:  "Start debugging",
			},
		},
	}

	enc := json.NewEncoder(wr)
	enc.SetIndent("", "  ")

	return enc.Encode(cfg)
}

func findWorkspaceDir(dir string) (string, bool) {
	if fi, err := os.Stat(path.Join(dir, ".vscode")); err == nil && fi.IsDir() {
		return dir, true
	}

	if _, err := os.Stat(path.Join(dir, "go.mod")); err == nil {
		return dir, true
	}

	if fi, err := os.Stat(path.Join(dir, ".git")); err == nil && fi.IsDir() {
		return dir, true
	}

	if dir == "/" {
		return "", false
	}

	return findWorkspaceDir(path.Dir(dir))
}
