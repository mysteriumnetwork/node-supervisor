// Copyright (c) 2020 BlockDev AG
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package daemon

import (
	"fmt"
	"io"
	"log"
	"strings"
)

type responder struct {
	io.Writer
}

func (r *responder) ok(result ...string) {
	args := []string{"ok"}
	args = append(args, result...)
	r.message(strings.Join(args, ": "))
}

func (r *responder) err(result ...error) {
	args := []string{"error"}
	for _, err := range result {
		args = append(args, err.Error())
	}
	r.message(strings.Join(args, ": "))
}

func (r *responder) message(msg string) {
	if _, err := fmt.Fprintln(r, msg); err != nil {
		log.Printf("Could not send message: %q error: %s\n", msg, err)
	}
}
