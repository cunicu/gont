// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package execvpe

// The following is a port of glibc execvep(2) implementation to Go
// See: https://sourceware.org/git/?p=glibc.git;a=blob;f=posix/execvpe.c

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// The file is accessible but it is not an executable file.
// Invoke the shell to interpret it as a script.
func maybeScriptExecute(argv0 string, argv []string, envp []string) error {
	argc := len(argv)

	newArgv := []string{"/bin/sh", argv0}
	if argc > 1 {
		newArgv = append(newArgv, argv[1:]...)
	}

	return syscall.Exec(newArgv[0], newArgv, envp) //nolint:gosec
}

// Execvpe searches the executable binary or a shell script in the currently configured
// path using a custom environment
func Execvpe(argv0 string, argv []string, envp []string) error {
	var err error

	// We check the simple case first.
	if argv0 == "" {
		return syscall.ENOENT
	}

	// Don't search when it contains a slash.
	if strings.Contains(argv0, "/") {
		err := syscall.Exec(argv0, argv, envp)

		if errors.Is(err, syscall.ENOEXEC) {
			if err := maybeScriptExecute(argv0, argv, envp); err != nil {
				return err
			}
		}

		return syscall.EINVAL
	}

	path := os.Getenv("PATH")
	if path == "" {
		path = "/bin:/usr/bin"
	}

	gotEacces := false
	for _, p := range strings.Split(path, ":") {
		argv0 := filepath.Join(p, argv0)
		if err = syscall.Exec(argv0, argv, envp); errors.Is(err, syscall.ENOEXEC) {
			err = maybeScriptExecute(argv0, argv, envp)
		}

		switch {
		case errors.Is(err, syscall.EACCES):
			// Record that we got a 'Permission denied' error.  If we end
			// up finding no executable we can use, we want to diagnose
			// that we did find one but were denied access.
			gotEacces = true

		case errors.Is(err, syscall.ENOENT):
			fallthrough
		case errors.Is(err, syscall.ESTALE):
			fallthrough
		case errors.Is(err, syscall.ENOTDIR):
			fallthrough

		// Those errors indicate the file is missing or not executable
		// by us, in which case we want to just try the next path
		// directory.
		case errors.Is(err, syscall.ENODEV):
			fallthrough
		case errors.Is(err, syscall.ETIMEDOUT):

			// Some strange file systems like AFS return even
			// stranger error numbers.  They cannot reasonably mean
			// anything else so ignore those, too.
		default:
			// Some other error means we found an executable file, but
			// something went wrong executing it; return the error to our
			// caller.
			return syscall.EINVAL
		}
	}

	// We tried every element and none of them worked.
	if gotEacces {
		return syscall.EACCES
	}

	return err
}
