//go:build windows
// +build windows

package util

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

const defaultShell = "cmd /c"

func CmdOutput(cmd *exec.Cmd, buf *bytes.Buffer) error {
	return fmt.Error("unsupported operating system")
}
