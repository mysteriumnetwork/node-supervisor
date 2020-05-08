package wginterface

import (
	"fmt"
	"log"
	"net"

	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
)

type WgInterface struct {
	Name   string
	device *device.Device
	uapi   net.Listener
	stop   chan struct{}
}

func New(requestedInterfaceName string) (*WgInterface, error) {
	tunnel, interfaceName, err := createTunnel(requestedInterfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device %s: %w", requestedInterfaceName, err)
	}

	logger := device.NewLogger(device.LogLevelInfo, fmt.Sprintf("(%s) ", interfaceName))
	logger.Info.Println("Starting wireguard-go version", device.WireGuardGoVersion)

	wgDevice := device.NewDevice(tunnel, logger)
	logger.Info.Println("Device started")

	fileUAPI, err := ipc.UAPIOpen(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("UAPI listen error: %w", err)
	}

	uapi, err := ipc.UAPIListen(interfaceName, fileUAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UAPI socket: %w", err)
	}

	return &WgInterface{
		Name:   interfaceName,
		device: wgDevice,
		uapi:   uapi,
		stop:   make(chan struct{}),
	}, nil
}

func createTunnel(requestedInterfaceName string) (tunnel tun.Device, interfaceName string, err error) {
	tunnel, err = tun.CreateTUN(requestedInterfaceName, device.DefaultMTU)
	if err == nil {
		interfaceName = requestedInterfaceName
		realInterfaceName, err2 := tunnel.Name()
		if err2 == nil {
			interfaceName = realInterfaceName
		}
	}
	return tunnel, interfaceName, err
}

func (a *WgInterface) Listen() {
	for {
		conn, err := a.uapi.Accept()
		if err != nil {
			log.Println("Closing UAPI listener, err:", err)
			return
		}
		go a.device.IpcHandle(conn)
	}
}

func (a *WgInterface) Down() {
	_ = a.uapi.Close()
	a.device.Close()
}
