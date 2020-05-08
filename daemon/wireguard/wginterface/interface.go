// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package wginterface

import (
	"net"

	"golang.zx2c4.com/wireguard/device"
)

type WgInterface struct {
	Name   string
	device *device.Device
	uapi   net.Listener
}
