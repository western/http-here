//go:build unix
// +build unix

package app

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

func MakeLibreofficeThumbnail(filepath_tmp, arg_fold_path string) {

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", "libreoffice --headless --norestore --nologo --convert-to png --outdir "+filepath_tmp+" \""+arg_fold_path+"\"")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmdReader, _ := cmd.StderrPipe()
	//cmd.Stdout = cmd.Stderr   // combine ERR + OUT
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	scanner1 := bufio.NewScanner(cmdReader)
	for scanner1.Scan() {
		fmt.Println("libreoffice:", scanner1.Text())
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Command finished on its own
		if err != nil {
			//fmt.Printf("Command finished with error: %v\n", err)
		} else {
			//fmt.Println("Command finished successfully")
		}

	case <-ctx.Done():
		// Context was canceled or timed out
		//fmt.Println("Context done, killing process group...")
		// Get the process group ID (pgid) and kill the entire group using a negative PID
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			// Use -pgid to kill the process group. SIGKILL (9) is forceful.
			if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
				//fmt.Printf("Failed to kill process group: %v\n", err)
			}
		} else {
			//fmt.Printf("Failed to get pgid: %v\n", err)
		}
		// Wait again to reap the process and avoid zombies
		<-done
		//fmt.Printf("Command terminated: %v\n", ctx.Err())
	}

}
