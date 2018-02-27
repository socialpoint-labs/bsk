package netutil

import (
	"net"
)

// FreeTCPAddr returns an available TCP port
func FreeTCPAddr() *net.TCPAddr {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	if err := l.Close(); err != nil {
		panic(err)
	}

	return l.Addr().(*net.TCPAddr)
}

// FreeUDPAddr returns an available UDP port
func FreeUDPAddr() *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	if err := l.Close(); err != nil {
		panic(err)
	}

	return l.LocalAddr().(*net.UDPAddr)
}
