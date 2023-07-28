package util

import (
	"math"
	"net"

	"golang.org/x/sys/unix"
)

// round iFloat to iDecimalPlaces decimal points.
func RoundToDecimal(iFloat float64, iDecimalPlaces int) float64 {
	var multiplier float64 = 10
	for i := 1; i < iDecimalPlaces; i++ {
		multiplier *= 10
	}
	return math.Round(iFloat*multiplier) / multiplier
}

// UDPAddr -> RawSockaddr conversion.
func NetAddrToSockaddr(addr *net.UDPAddr) unix.Sockaddr {
	if ip := addr.IP.To4(); ip != nil {
		return &unix.SockaddrInet4{
			Port: addr.Port,
			Addr: [4]byte(ip),
		}
	}

	if ip := addr.IP.To16(); ip != nil {
		return &unix.SockaddrInet6{
			Addr: [16]byte(ip),
			Port: addr.Port,
		}
	}

	return nil
}

// RawSockaddr -> UDPAddr conversion.
func SockaddrToUDPAddr(sockAddr unix.Sockaddr) *net.UDPAddr {
	switch typedSocketAddr := sockAddr.(type) {
	case *unix.SockaddrInet4:
		return &net.UDPAddr{
			IP:   typedSocketAddr.Addr[:],
			Port: typedSocketAddr.Port,
		}

	case *unix.SockaddrInet6:
		zone := ""
		ifi, err := net.InterfaceByIndex(int(typedSocketAddr.ZoneId))
		if err == nil && typedSocketAddr.ZoneId != 0 {
			zone = ifi.Name
		}
		return &net.UDPAddr{
			IP:   typedSocketAddr.Addr[:],
			Port: typedSocketAddr.Port,
			Zone: zone,
		}
	}

	return nil
}
