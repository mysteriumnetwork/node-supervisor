// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// +build mage

package main

import "github.com/magefile/mage/sh"

const BIN = "build/myst_supervisor"

// Build builds the myst supervisor.
func Build() error {
	return sh.Run("go", "build", "-o", BIN, "cmd/myst_supervisor.go")
}
