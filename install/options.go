// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package install

// Options for installation.
type Options struct {
	SupervisorPath string
}

func (o Options) valid() bool {
	return o.SupervisorPath != ""
}
