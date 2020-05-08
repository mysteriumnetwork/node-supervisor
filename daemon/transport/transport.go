// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package transport

import "io"

// handlerFunc talks to a connected client.
type handlerFunc func(conn io.ReadWriter)
