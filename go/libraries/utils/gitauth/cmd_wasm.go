//go:build js && wasm
// +build js,wasm

package gitauth

import "os/exec"

func CmdSetsid(cmd *exec.Cmd) {}
