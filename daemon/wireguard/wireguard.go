package wireguard

import (
	"fmt"
	"sync"

	"github.com/mysteriumnetwork/node-supervisor/daemon/wireguard/wginterface"
)

type Monitor struct {
	interfaces map[string]*wginterface.WgInterface
	sync.Mutex
}

func NewMonitor() *Monitor {
	return &Monitor{
		interfaces: make(map[string]*wginterface.WgInterface),
	}
}

func (m *Monitor) Up(requestedInterfaceName string) (string, error) {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.interfaces[requestedInterfaceName]; exists {
		return "", fmt.Errorf("interface %s already exists", requestedInterfaceName)
	}
	iface, err := wginterface.New(requestedInterfaceName)
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
