//go:build unix
// +build unix

package lopool

// lo_pool_unix.go - LibreOffice pool platform helpers for Unix.

import (
	"crypto/rand"
	"os/exec"
	"syscall"
)

// cryptoRandRead - wrapper for randomHexLo
func cryptoRandRead(b []byte) error {
	_, err := rand.Read(b)
	return err
}

// killProcessGroupLo - launch soffice in its own process group,
// so that we can subsequently kill the entire tree (soffice spawns
// child processes such as oosplash/soffice.bin).
func killProcessGroupLo(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killLibreOfficeTreeLo - Sends SIGKILL to the entire process group, identified by its PGID.
func killLibreOfficeTreeLo(cmd *exec.Cmd) {

	if cmd.Process == nil {
		return
	}

	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		// fallback: to kill off the process itself, at the very least
		_ = cmd.Process.Kill()
		return
	}

	// negative PID = kill the entire process group
	_ = syscall.Kill(-pgid, syscall.SIGKILL)

	// we wait, so as not to leave any zombies behind
	_, _ = cmd.Process.Wait()
}
