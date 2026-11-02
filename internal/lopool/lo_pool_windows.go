//go:build windows
// +build windows

package lopool

// lo_pool_windows.go - LibreOffice pool platform helpers for Windows.

import (
	"crypto/rand"
	"os"
	"os/exec"
	"strconv"
)

// cryptoRandRead - wrapper for randomHexLo
func cryptoRandRead(b []byte) error {
	_, err := rand.Read(b)
	return err
}

// killProcessGroupLo - on Windows, the process tree is handled via
// taskkill /T; no separate preparation step is required.
func killProcessGroupLo(cmd *exec.Cmd) {
	// no-op
}

// killLibreOfficeTreeLo - kills soffice.exe and all its descendants
// using taskkill /T /F.
func killLibreOfficeTreeLo(cmd *exec.Cmd) {

	if cmd.Process == nil {
		return
	}

	_ = exec.Command("taskkill",
		"/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()

	// final fallback
	_ = cmd.Process.Kill()

	// We signal that the process is no longer waiting (Wait)
	_ = os.ErrProcessDone
}
