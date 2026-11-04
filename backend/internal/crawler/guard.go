package crawler

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"syscall"
)

var blockedPrefixes = mustPrefixes(
	"0.0.0.0/8",       // "this" network
	"100.64.0.0/10",   // carrier-grade NAT
	"192.0.0.0/24",    // IETF protocol assignments
	"192.0.2.0/24",    // documentation
	"198.18.0.0/15",   // benchmarking
	"198.51.100.0/24", // documentation
	"203.0.113.0/24",  // documentation
	"240.0.0.0/4",     // reserved
	"64:ff9b::/96",    // NAT64
	"2001:db8::/32",   // documentation
)

func mustPrefixes(list ...string) []netip.Prefix {
	out := make([]netip.Prefix, len(list))
	for i, s := range list {
		out[i] = netip.MustParsePrefix(s)
	}
	return out
}

// IPAllowed reports whether a public-internet address may be contacted. Loopback, private, link-local (which
// includes the 169.254.169.254 cloud metadata address), multicast and reserved ranges are refused.
func IPAllowed(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !addr.IsValid() || addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() ||
		addr.IsInterfaceLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}
	for _, p := range blockedPrefixes {
		if p.Contains(addr) {
			return false
		}
	}
	return true
}

// dialControl runs after DNS resolution and just before connect, on the exact address about to be used, so a
// hostname that resolves (or later re-resolves) to an internal address cannot slip through.
func dialControl(_, address string, _ syscall.RawConn) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return ErrForbiddenURL
	}
	if port != "80" && port != "443" {
		return fmt.Errorf("%w: port %s", ErrForbiddenURL, port)
	}
	addr, err := netip.ParseAddr(host)
	if err != nil || !IPAllowed(addr) {
		return fmt.Errorf("%w: address", ErrForbiddenURL)
	}
	return nil
}

// CheckHost resolves a host up front so the dashboard can reject an obviously internal URL immediately.
// The dial-time check remains the authority.
func CheckHost(ctx context.Context, host string) error {
	if addr, err := netip.ParseAddr(host); err == nil {
		if !IPAllowed(addr) {
			return ErrForbiddenURL
		}
		return nil
	}
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return ErrUnreachable
	}
	for _, ip := range ips {
		if !IPAllowed(ip) {
			return ErrForbiddenURL
		}
	}
	return nil
}
