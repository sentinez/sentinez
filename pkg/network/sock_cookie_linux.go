// Copyright 2026 Sentinéz Labs.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//go:build linux

package network

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

var errNoSyscallConn = errors.New("so_cookie: conn has no file descriptor")

// sockCookie returns the kernel socket cookie (SO_COOKIE, Linux 4.12+) of
// conn: unique per network namespace and the same value eBPF programs read
// with bpf_get_socket_cookie.
func sockCookie(conn net.Conn) (uint64, error) {
	sc, ok := conn.(syscall.Conn)
	if !ok {
		return 0, errNoSyscallConn
	}

	raw, err := sc.SyscallConn()
	if err != nil {
		return 0, fmt.Errorf("so_cookie: %w", err)
	}

	var res struct {
		cookie uint64
		err    error
	}

	err = raw.Control(func(fd uintptr) {
		res.cookie, res.err = unix.GetsockoptUint64(
			int(fd), unix.SOL_SOCKET, unix.SO_COOKIE)
	})
	if err != nil {
		return 0, fmt.Errorf("so_cookie: %w", err)
	}

	if res.err != nil {
		return 0, fmt.Errorf("so_cookie: %w", res.err)
	}

	return res.cookie, nil
}
