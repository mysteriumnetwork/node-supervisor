// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package daemon

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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
	cmd := exec.Cmd{
		Path: d.cfg.MystPath,
		Args: []string{d.cfg.MystPath, "--openvpn.binary", d.cfg.OpenVPNPath, "daemon"},
		Env: []string{
			"HOME=" + d.cfg.MystHome,
			"PATH=" + strings.Join(paths, ":"),
		},
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	pid := cmd.Process.Pid
	if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0700); err != nil {
		return err
	}

	out, err := cmd.CombinedOutput()
	log.Printf("myst [err=%s] output: %s\n", err, out)
	return err
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
