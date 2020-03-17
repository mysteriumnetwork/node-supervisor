// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package install

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const daemonID = "network.mysterium.myst_supervisor"
const plistPath = "/Library/LaunchDaemons/" + daemonID + ".plist"
const plistTpl = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.DaemonID}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.SupervisorPath}}</string>
	</array>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>{{.LogPath}}</string>
	<key>StandardErrorPath</key>
	<string>{{.LogPath}}</string>
</dict>
</plist>
`

func Install(options Options) error {
	if !options.valid() {
		return errors.New("invalid options")
	}
	log.Println("Installing launchd daemon")
	clean()
	tpl, err := template.New("plistTpl").Parse(plistTpl)
	if err != nil {
		return fmt.Errorf("could not create template for %s: %w", plistPath, err)
	}
	fd, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", plistPath, err)
	}
	err = tpl.Execute(fd, map[string]string{
		"DaemonID":       daemonID,
		"LogPath":        "/var/log/myst_supervisor.log",
		"SupervisorPath": options.SupervisorPath,
	})
	if err != nil {
		return fmt.Errorf("could not generate %s: %w", plistPath, err)
	}
	out, err := runV("launchctl", "load", plistPath)
	if err == nil && strings.Contains(out, "Invalid property") {
		err = errors.New("invalid plist file")
	}
	if err != nil {
		return fmt.Errorf("could not load launch daemon: %w", err)
	}
	return nil
}

func clean() {
	log.Println("Cleaning up previous installation")
	_, _ = runV("launchctl", "unload", plistPath)
	_ = os.RemoveAll(plistPath)
}

func runV(c ...string) (string, error) {
	cmd := exec.Command(c[0], c[1:]...)
	output, err := cmd.CombinedOutput()
	log.Printf("[%v] out:\n%s", strings.Join(c, " "), output)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
