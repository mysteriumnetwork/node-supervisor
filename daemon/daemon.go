// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package daemon

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/jackpal/gateway"
	"github.com/mysteriumnetwork/node-supervisor/config"
	"github.com/mysteriumnetwork/node-supervisor/daemon/transport"
	"github.com/mysteriumnetwork/node-supervisor/daemon/wireguard"
)

// Daemon - supervisor process.
type Daemon struct {
	cfg     *config.Config
	monitor *wireguard.Monitor
}

// New creates a new daemon.
func New(cfg *config.Config) Daemon {
	return Daemon{cfg: cfg, monitor: wireguard.NewMonitor()}
}

// Start supervisor daemon. Blocks.
func (d Daemon) Start() error {
	return transport.Start(d.dialog)
}

// dialog talks to the client via established connection.
func (d Daemon) dialog(conn io.ReadWriter) {
	scan := bufio.NewScanner(conn)
	answer := responder{conn}
	for scan.Scan() {
		line := scan.Bytes()
		cmd := strings.Split(string(line), " ")
		op := cmd[0]
		switch op {
		case CommandBye:
			answer.ok("bye")
			return
		case CommandPing:
			answer.ok("pong")
		case CommandRun:
			go func() {
				err := d.RunMyst()
				if err != nil {
					log.Printf("failed %s: %s", CommandRun, err)
					answer.err(err)
				} else {
					answer.ok()
				}
			}()
		case CommandWgUp:
			up, err := d.wgUp(cmd...)
			if err != nil {
				log.Printf("failed %s: %s", CommandWgUp, err)
				answer.err(err)
			} else {
				answer.ok(up)
			}
		case CommandWgDown:
			err := d.wgDown(cmd...)
			if err != nil {
				log.Printf("failed %s: %s", CommandWgDown, err)
				answer.err(err)
			} else {
				answer.ok()
			}
		case CommandAssignIP:
			err := d.assignIP(cmd...)
			if err != nil {
				log.Printf("failed %s: %s", CommandAssignIP, err)
				answer.err(err)
			} else {
				answer.ok()
			}
		case CommandExcludeRoute:
			err := d.excludeRoute(cmd...)
			if err != nil {
				log.Printf("failed %s: %s", CommandExcludeRoute, err)
				answer.err(err)
			} else {
				answer.ok()
			}
		case CommandDefaultRoute:
			err := d.defaultRoute(cmd...)
			if err != nil {
				log.Printf("failed %s: %s", CommandDefaultRoute, err)
				answer.err(err)
			} else {
				answer.ok()
			}
		case CommandKill:
			if err := d.KillMyst(); err != nil {
				log.Println("Could not kill myst:", err)
				answer.err(err)
			} else {
				answer.ok()
			}
		}
	}
}

func (d Daemon) wgUp(args ...string) (interfaceName string, err error) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	requestedInterfaceName := flags.String("iface", "", "Requested tunnel interface name")
	uid := flags.String("uid", "", "User ID."+
		" On POSIX systems, this is a decimal number representing the uid."+
		" On Windows, this is a security identifier (SID) in a string format.")
	if err := flags.Parse(args[1:]); err != nil {
		return "", err
	}
	if *requestedInterfaceName == "" {
		return "", errors.New("-iface is required")
	}
	if *uid == "" {
		return "", errors.New("-uid is required")
	}
	return d.monitor.Up(*requestedInterfaceName, *uid)
}

func (d Daemon) wgDown(args ...string) (err error) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	interfaceName := flags.String("iface", "", "")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *interfaceName == "" {
		return errors.New("-iface is required")
	}
	return d.monitor.Down(*interfaceName)
}

func (d Daemon) assignIP(args ...string) (err error) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	interfaceName := flags.String("iface", "", "")
	network := flags.String("net", "", "")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *interfaceName == "" {
		return errors.New("-iface is required")
	}
	if *network == "" {
		return errors.New("-net is required")
	}
	_, ipNet, err := net.ParseCIDR(*network)
	if err != nil {
		return fmt.Errorf("-net could not be parsed: %w", err)
	}
	output, err := exec.Command("sudo", "ifconfig", *interfaceName, *network, peerIP(*ipNet).String()).CombinedOutput()
	if err != nil {
		log.Println(output)
		return fmt.Errorf("ifconfig returned error: %w", err)
	}
	return nil
}

func (d Daemon) excludeRoute(args ...string) (err error) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	ip := flags.String("ip", "", "")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *ip == "" {
		return errors.New("-ip is required")
	}
	parsedIP := net.ParseIP(*ip)
	if parsedIP == nil {
		return fmt.Errorf("-ip could not be parsed: %w", err)
	}
	gw, err := gateway.DiscoverGateway()
	if err != nil {
		return err
	}
	output, err := exec.Command("route", "add", "-host", parsedIP.String(), gw.String()).CombinedOutput()
	if err != nil {
		log.Println(output)
		return fmt.Errorf("route add returned error: %w", err)
	}
	return nil
}

func (d Daemon) defaultRoute(args ...string) (err error) {
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	interfaceName := flags.String("iface", "", "")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *interfaceName == "" {
		return errors.New("-iface is required")
	}
	output, err := exec.Command("route", "add", "-net", "0.0.0.0/1", "-interface", *interfaceName).CombinedOutput()
	if err != nil {
		log.Println(output)
		return fmt.Errorf("route add returned error: %w", err)
	}
	output, err = exec.Command("route", "add", "-net", "128.0.0.0/1", "-interface", *interfaceName).CombinedOutput()
	if err != nil {
		log.Println(output)
		return fmt.Errorf("route add returned error: %w", err)
	}
	return nil
}

func peerIP(subnet net.IPNet) net.IP {
	lastOctetID := len(subnet.IP) - 1
	if subnet.IP[lastOctetID] == byte(1) {
		subnet.IP[lastOctetID] = byte(2)
	} else {
		subnet.IP[lastOctetID] = byte(1)
	}
	return subnet.IP
}
