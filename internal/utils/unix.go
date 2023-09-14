//go:build !windows

package utils

import (
	"bytes"
	"os/exec"
)

func Exec(name string, stdout *bytes.Buffer, stderr *bytes.Buffer, args ...string) error {
	command := exec.Command(name, args...)
	command.Stdout = stdout
	command.Stderr = stderr

	return command.Run()
}
