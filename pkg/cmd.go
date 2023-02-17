package gont

import "os/exec"

type CmdOption interface {
	Apply(*exec.Cmd)
}
