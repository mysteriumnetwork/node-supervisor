// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package wireguard

import (
	"fmt"
	"sync"

	"github.com/mysteriumnetwork/node-supervisor/daemon/wireguard/wginterface"
)

// Monitor creates/deletes the wireguard interfaces and keeps track of them.
type Monitor struct {
	interfaces map[string]*wginterface.WgInterface
	sync.Mutex
}

func NewMonitor() *Monitor {
	return &Monitor{
		interfaces: make(map[string]*wginterface.WgInterface),
	}
}

func (m *Monitor) Up(requestedInterfaceName string, uid string) (string, error) {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.interfaces[requestedInterfaceName]; exists {
		return "", fmt.Errorf("interface %s already exists", requestedInterfaceName)
	}
	iface, err := wginterface.New(requestedInterfaceName, uid)
	if err != nil {
		return "", err
	}

	go iface.Listen()
	m.interfaces[iface.Name] = iface
	return iface.Name, err
}

func (m *Monitor) Down(interfaceName string) error {
	m.Lock()
	defer m.Unlock()

	iface, ok := m.interfaces[interfaceName]
	if !ok {
		return fmt.Errorf("interface %s not found", interfaceName)
	}

	iface.Down()
	delete(m.interfaces, interfaceName)
	return nil
}
