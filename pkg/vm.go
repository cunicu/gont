// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

var ErrMissingCloudInitData = errors.New("missing cloud-init user data")

type QEmuVMOption interface {
	ApplyQEmuVM(vm *QEmuVM)
}

var _ Node = (*QEmuVM)(nil)

type CloudInit struct {
	MetaData      map[string]any
	UserData      map[string]any
	NetworkConfig map[string]any
}

type QEmuVM struct {
	*BaseNode

	options []any
	command *Cmd

	// Options
	Arch      string
	CloudInit CloudInit
}

func (n *Network) AddQEmuVM(name string, opts ...Option) (*QEmuVM, error) {
	baseNode, err := n.addBaseNode(name, opts)
	if err != nil {
		return nil, err
	}

	vm := &QEmuVM{
		BaseNode: baseNode,
	}

	n.Register(vm)

	// Apply VM options
	for _, opt := range opts {
		switch opt := opt.(type) {
		case QEmuVMOption:
			opt.ApplyQEmuVM(vm)
		case ExecCmdOption, CmdOption:
			vm.options = append(vm.options, opt)
		}
	}

	// Add links
	if err := vm.configureLinks(); err != nil {
		return nil, fmt.Errorf("failed to configure links: %w", err)
	}

	return vm, nil
}

func (vm *QEmuVM) Close() error {
	if vm.command.Process != nil {
		if err := vm.Stop(); err != nil {
			return fmt.Errorf("failed to stop VM: %w", err)
		}
	}

	return nil
}

func (vm *QEmuVM) configureLinks() error {
	for range vm.ConfiguredInterfaces {
		return errors.ErrUnsupported
	}

	return nil
}

func (vm *QEmuVM) ConfigureInterface(_ *Interface) error {
	return errors.ErrUnsupported
}

func (vm *QEmuVM) Teardown() error {
	return errors.ErrUnsupported
}

func (vm *QEmuVM) Start() (*Cmd, error) {
	name := fmt.Sprintf("/nix/store/v3jjx86y2cs6x9182f3d1i3z1kbap876-qemu-8.2.3/bin/qemu-system-%s", vm.Arch)

	args := slices.Clone(vm.options)

	if vm.CloudInit.UserData != nil {
		fn, err := vm.createCloudInitImage()
		if err != nil {
			return nil, fmt.Errorf("failed to create cloud-init seed image: %w", err)
		}

		args = append(args, "-drive", fmt.Sprintf("if=virtio,format=raw,file=%s", fn))
	}

	vm.command = vm.network.HostNode.Command(name, args...)
	vm.command.Stdin = os.Stdin
	vm.command.Stderr = os.Stderr
	vm.command.Stdout = os.Stdout

	if err := vm.command.Start(); err != nil {
		return nil, err
	}

	return vm.command, nil
}

func (vm *QEmuVM) Stop() error {
	// TODO: Perform orderly shutdown of VM
	return vm.command.Process.Kill()
}

func (vm *QEmuVM) createCloudInitImage() (string, error) {
	if vm.CloudInit.UserData == nil {
		return "", ErrMissingCloudInitData
	}

	fnOut := filepath.Join(vm.BasePath, "cloud-init.img")

	args := []any{
		"--disk-format", "raw",
		"--filesystem", "iso9660",
		fnOut,
	}

	out := []byte("#cloud-config\n")
	outYAML, err := yaml.Marshal(vm.CloudInit.UserData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cloud-init metadata: %w", err)
	}

	out = append(out, outYAML...)

	fnUser := filepath.Join(vm.BasePath, "cloud-init-userdata.yaml")
	if err := os.WriteFile(fnUser, out, 0o600); err != nil {
		return "", fmt.Errorf("failed to write file: %s: %w", fnUser, err)
	}

	args = append(args, fnUser)

	if vm.CloudInit.MetaData != nil {
		meta := maps.Clone(vm.CloudInit.MetaData)
		meta["dsmode"] = "local"

		out, err := yaml.Marshal(meta)
		if err != nil {
			return "", fmt.Errorf("failed to marshal cloud-init metadata: %w", err)
		}

		fnMeta := filepath.Join(vm.BasePath, "cloud-init-metadata.yaml")
		if err := os.WriteFile(fnMeta, out, 0o600); err != nil {
			return "", fmt.Errorf("failed to write file: %s: %w", fnMeta, err)
		}

		args = append(args, fnMeta)
	}

	if vm.CloudInit.NetworkConfig != nil {
		out, err := yaml.Marshal(vm.CloudInit.NetworkConfig)
		if err != nil {
			return "", fmt.Errorf("failed to marshal cloud-init network config: %w", err)
		}

		fnNetCfg := filepath.Join(vm.BasePath, "cloud-init-network.yaml")
		if err := os.WriteFile(fnNetCfg, out, 0o600); err != nil {
			return "", fmt.Errorf("failed to write file: %s: %w", fnNetCfg, err)
		}

		// args = append(args, fnNetCfg)
	}

	cmd := vm.network.HostNode.Command("/nix/store/05dsd12j15sg8qbf7jz6dg5kv7q32z1c-cloud-utils-0.32/bin/cloud-localds", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run cloud-localds: %w", err)
	}

	return fnOut, nil
}
