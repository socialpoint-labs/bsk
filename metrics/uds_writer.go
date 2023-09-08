package metrics

import (
	"errors"
	"net"
	"syscall"
)

type UDSWriter struct {
	address string
	conn    net.Conn
}

func (u *UDSWriter) Write(p []byte) (n int, err error) {
	if u.conn == nil {
		if u.conn, err = connect(u.address); err != nil {
			return 0, err
		}
	}

	n, err = u.conn.Write(p)

	if err != nil && errors.Is(err, syscall.ECONNRESET) {
		_ = u.conn.Close()
		u.conn = nil
	}

	return n, err
}

func connect(address string) (net.Conn, error) {
	conn, err := net.Dial("unixgram", address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewUDSWriter(address string) *UDSWriter {
	return &UDSWriter{address: address}
}
