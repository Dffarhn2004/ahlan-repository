//go:build linux

package hermes

import (
	"net"
	"syscall"
)

// peerUID returns the UID of the process on the other end of a Unix socket connection (SO_PEERCRED).
func peerUID(conn net.Conn) (uint32, bool) {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return 0, false
	}
	raw, err := unixConn.SyscallConn()
	if err != nil {
		return 0, false
	}
	var uid uint32
	var credErr error
	if err := raw.Control(func(fd uintptr) {
		cred, e := syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if e != nil {
			credErr = e
			return
		}
		uid = cred.Uid
	}); err != nil || credErr != nil {
		return 0, false
	}
	return uid, true
}
