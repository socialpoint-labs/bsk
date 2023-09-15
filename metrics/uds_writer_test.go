package metrics_test

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/metrics"
)

func TestUDSWriter_Write(t *testing.T) {
	const payload = "foo,bar,baz"

	t.Parallel()
	a := assert.New(t)

	conn := arrangeListeningConnection(t)
	defer conn.close()
	w := metrics.NewUDSWriter(conn.addr)

	// assert we can write to the connection
	n, err := w.Write([]byte(payload))
	a.Equal(11, n)
	a.NoError(err)
	conn.assertRead(payload)

	// close the connection and make writes fail
	conn.close()
	n, err = w.Write([]byte(payload))
	a.Equal(0, n)
	a.Error(err)
	// one more time so that it encounters the conn set to nil
	n, err = w.Write([]byte(payload))
	a.Equal(0, n)
	a.Error(err)

	// reopen and assert we can write again
	conn.open()
	n, err = w.Write([]byte(payload))
	a.Equal(11, n)
	a.NoError(err)
	conn.assertRead(payload)
}

type listeningConnection struct {
	t    testing.TB
	addr string
	conn *net.UnixConn
}

func (lc *listeningConnection) open() {
	lc.conn = arrangeListeningUnixConnection(lc.t, lc.addr)
}

func (lc *listeningConnection) close() {
	_ = lc.conn.Close()
	lc.conn = nil

	_ = os.Remove(lc.addr)
}

func (lc *listeningConnection) assertRead(s string) {
	a := assert.New(lc.t)

	bb, err := lc.read()
	a.NoError(err)
	a.Equal(s, string(bb))
}

func (lc *listeningConnection) read() ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := lc.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

func arrangeListeningConnection(t testing.TB) *listeningConnection {
	addr := arrangeUnixAddr(t)
	conn := arrangeListeningUnixConnection(t, addr)

	return &listeningConnection{
		t:    t,
		addr: addr,
		conn: conn,
	}
}

func arrangeUnixAddr(t testing.TB) string {
	d, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(d); err != nil {
			t.Error(err)
		}
	})
	return filepath.Join(d, "sock")
}

func arrangeListeningUnixConnection(t testing.TB, addr string) *net.UnixConn {
	la, err := net.ResolveUnixAddr("unixgram", addr)
	if err != nil {
		t.Fatal(err)
	}
	c, err := net.ListenUnixgram("unixgram", la)
	if err != nil {
		t.Fatal(err)
	}

	return c
}
