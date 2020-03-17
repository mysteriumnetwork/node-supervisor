// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mysteriumnetwork/node-supervisor/config"
	"github.com/mysteriumnetwork/node-supervisor/daemon"
	"github.com/mysteriumnetwork/node-supervisor/install"
)

var (
	FlagInstall     = flag.Bool("install", false, "Install or repair myst supervisor")
	FlagMystHome    = flag.String("mystHome", "", "Home directory for running myst (required for -install)")
	FlagMystPath    = flag.String("mystPath", "", "Path to myst executable (required for -install)")
	FlagOpenVPNPath = flag.String("openvpnPath", "", "Path to openvpn executable (required for -install)")
)

func ensureInstallFlags() {
	if *FlagMystHome == "" || *FlagMystPath == "" || *FlagOpenVPNPath == "" {
		fmt.Println("Error: required flags were not set")
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	if *FlagInstall {
		ensureInstallFlags()
		log.Println("Installing supervisor")
		path, err := thisPath()
		if err != nil {
			log.Fatalln("Failed to determine supervisor's path:", err)
		}
		err = install.Install(install.Options{
			SupervisorPath: path,
		})
		if err != nil {
			log.Fatalln("Failed to install supervisor:", err)
		}
		log.Println("Creating supervisor configuration")
		cfg := config.Config{
			MystHome:    *FlagMystHome,
			MystPath:    *FlagMystPath,
			OpenVPNPath: *FlagOpenVPNPath,
		}
		err = cfg.Write()
		if err != nil {
			log.Fatalln("Failed to create supervisor configuration:", err)
		}
	} else {
		log.Println("Running myst supervisor daemon")
		cfg, err := config.Read()
		if err != nil {
			log.Println("Failed to read supervisor configuration:", err)
		}
		supervisor := daemon.New(cfg)
		if err := supervisor.Start(); err != nil {
			log.Fatalln("Error running supervisor:", err)
		}
	}
}

func thisPath() (string, error) {
	thisExec, err := os.Executable()
	if err != nil {
		return "", err
	}
	thisPath, err := filepath.Abs(thisExec)
	if err != nil {
		return "", err
	}
	return thisPath, nil
}
