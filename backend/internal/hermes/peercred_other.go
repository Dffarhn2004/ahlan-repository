//go:build !linux

package hermes

import "net"

// peerUID is unavailable off Linux. With AllowedUIDs set the bridge then rejects every caller (fail closed).
func peerUID(net.Conn) (uint32, bool) { return 0, false }
