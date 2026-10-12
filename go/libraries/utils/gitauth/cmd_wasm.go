//go:build js && wasm
// +build js,wasm

package gitauth

import (
	"os"
	"os/exec"
)

func CmdSetsid(cmd *exec.Cmd) {}

// Browser builds cannot create subprocesses or process groups.
func killProcessGroup(cmd *exec.Cmd) error { return os.ErrProcessDone }
