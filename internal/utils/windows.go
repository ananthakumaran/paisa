//go:build windows

package utils

import (
	"bytes"
	"os/exec"
	"syscall"
)

func Exec(name string, stdout *bytes.Buffer, stderr *bytes.Buffer, args ...string) error {
	command := exec.Command(name, args...)
	command.Stdout = stdout
	command.Stderr = stderr

	command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	return command.Run()
}
