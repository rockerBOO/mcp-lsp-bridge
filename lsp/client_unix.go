//go:build !windows

package lsp

import (
	"os/exec"
	"syscall"
)

// setProcAttributes sets Unix-specific process attributes
func setProcAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
		Pgid:    0,    // 0 means use the PID as PGID
	}
}