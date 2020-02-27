package daemon

import (
	"log"
	"os/exec"
	"strings"
)

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
	run := exec.Cmd{
		Path: d.cfg.MystPath,
		Args: []string{d.cfg.MystPath, "daemon"},
		Env: []string{
			"HOME=" + d.cfg.MystHome,
			"PATH=" + strings.Join(paths, ":"),
		},
	}
	out, err := run.CombinedOutput()
	log.Printf("myst [err=%s] output: %s\n", err, out)
	return err
}
