// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package daemon

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const pidFile = "/var/run/myst.pid"

// RunMyst runs mysterium node daemon. Blocks.
func (d *Daemon) RunMyst() error {
	paths := []string{
		d.cfg.OpenVPNPath,
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
		"/usr/local/bin",
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Cmd{
		Path: d.cfg.MystPath,
		Args: []string{
			d.cfg.MystPath,
			"--openvpn.binary", d.cfg.OpenVPNPath,
			"--mymysterium.enabled=false",
			"--ui.enable=false",
			"daemon",
		},
		Env: []string{
			"HOME=" + d.cfg.MystHome,
			"PATH=" + strings.Join(paths, ":"),
		},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	err := runWithSuccessTimeout(cmd.Wait, 5*time.Second)
	if err != nil {
		log.Printf("myst output [err=%s]:\n%s\n%s\n", err, stderr.String(), stdout.String())
		return err
	}

	pid := cmd.Process.Pid
	if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0700); err != nil {
		return err
	}

	return nil
}

func runWithSuccessTimeout(f func() error, timeout time.Duration) error {
	done := make(chan error)
	go func() {
		done <- f()
	}()
	select {
	case <-time.After(timeout):
		return nil
	case err := <-done:
		return err
	}
}

func (d *Daemon) KillMyst() error {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return nil
	}
	pidFileContent, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("could not read %q: %w", pidFile, err)
	}
	pid, err := strconv.Atoi(string(pidFileContent))
	if err != nil {
		return fmt.Errorf("invalid content of %q: %w", pidFile, err)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("could not find process %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGINT); err != nil {
		return fmt.Errorf("could not interrupt process %d: %w", pid, err)
	}
	// TODO kill if doesn't terminate in 5secs
	_ = os.Remove(pidFile)
	return nil
}
