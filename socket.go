package camellia

import (
	"net"
	"strconv"
	"syscall"

	"github.com/arianxx/camellia-io/internal"
)

var inBuf = [1024]byte{}

// Socket includes the corresponding information of a socket.
type Socket struct {
	loop          *EventLoop
	fd            int
	network, addr string
	port          int
	sa            syscall.Sockaddr
	in, out       []byte
	closedCount   int
}

// NewSocket creates a new non-blocking socket.
func NewSocket(network, addr string, loop *EventLoop) (*Socket, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	if err = syscall.SetNonblock(fd, true); err != nil {
		return nil, err
	}
	sa, err := getSockAddr(network, addr)
	var portStr string
	addr, portStr, err = net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	var port int
	port, _ = strconv.Atoi(portStr)

	return &Socket{
		loop, fd, network, addr, port, sa,
		[]byte{}, []byte{}, 0,
	}, nil
}

// Close closes a socket.
func (s *Socket) Close() error {
	s.closedCount = 2
	return syscall.Close(s.fd)
}

// Shutdown stops one side of a socket.
func (s *Socket) Shutdown(how int) error {
	s.closedCount += 1
	if s.closedCount == 2 {
		return s.Close()
	}

	return syscall.Shutdown(s.fd, how)
}

// Listener is the decorators of the Socket to listen on the specific address.
type Listener struct {
	*Socket
}

// NewListener creates a new non-blocking listener.
func NewListener(network, addr string, loop *EventLoop) (*Listener, error) {
	sock, err := NewSocket(network, addr, loop)
	if err != nil {
		return nil, err
	}

	return &Listener{sock}, nil
}

// BindAndListen makes the listener bind and listen on the system.
func (l *Listener) BindAndListen() error {
	var err error
	err = syscall.Bind(l.fd, l.sa)
	if err != nil {
		_ = l.Close()
		return err
	}
	err = syscall.Listen(l.fd, 1024)
	if err != nil {
		_ = l.Close()
		return err
	}
	return nil
}

// RegisterAccept register the accept event in the eventloop.
func (l *Listener) RegisterAccept() error {
	return l.loop.Register(l.fd, internal.EV_READABLE, l.acceptEvent, nil)
}

func (l *Listener) acceptEvent(el *EventLoop, _ interface{}) Action {
	nfd, sa, err := syscall.Accept(l.fd)
	if err != nil {
		return CONTINUE
	}
	if err = syscall.SetNonblock(nfd, true); err != nil {
		_ = syscall.Close(nfd)
		return CONTINUE
	}

	c, err := NewConn(nfd, sa, el)
	if err != nil {
		return CONTINUE
	}

	_ = el.Register(c.fd, internal.EV_READABLE, c.readEvent, nil)
	el.SetTriggerDataPtr([]string{c.network, c.addr, strconv.Itoa(c.port)})
	return TRIGGER_OPEN_EVENT
}

// Conn is the decorators of the Socket to process a specific connection.
type Conn struct {
	*Socket
}

// NewConn creates a new Conn.
func NewConn(fd int, sa syscall.Sockaddr, loop *EventLoop) (*Conn, error) {
	conn := &Conn{&Socket{fd: fd, sa: sa, loop: loop}}
	conn.in, conn.out = []byte{}, []byte{}
	var err error
	conn.network, conn.addr, conn.port, err = resolveSockaddrInfo(sa)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Read reads the buffered data from Conn.
func (c *Conn) Read() []byte {
	res := c.in
	c.in = []byte{}
	return res
}

// Write writes some data to the Conn that will be sent later.
func (c *Conn) Write(d []byte) {
	c.out = append(c.out, d...)
}

func (c *Conn) readEvent(el *EventLoop, _ interface{}) Action {
	var (
		n      int
		err    error
		action Action
	)

	n, err = syscall.Read(c.fd, inBuf[:])
	if err == syscall.EINTR || err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
		action = CONTINUE
	} else if n <= 0 {
		_ = c.Shutdown(syscall.SHUT_RD)
		action = SHUTDOWN_RD
	} else {
		c.in = append(c.in, inBuf[:n]...)
		el.SetTriggerDataPtr(c)
		action = TRIGGER_DATA_EVENT
	}

	if c.closedCount == 0 {
		_ = el.Register(c.fd, internal.EV_WRITABLE, c.writeEvent, nil)
	}
	return action
}

func (c *Conn) writeEvent(el *EventLoop, _ interface{}) Action {
	var action = CONTINUE

	if len(c.out) != 0 {
		n, err := syscall.Write(c.fd, c.out)
		if err != nil {
			_ = c.Shutdown(syscall.SHUT_WR)
			action = SHUTDOWN_WR
		} else {
			c.out = c.out[n:]
		}
	}

	if c.closedCount == 0 {
		_ = el.Register(c.fd, internal.EV_READABLE, c.readEvent, nil)
	}
	return action
}

func getSockAddr(network, addr string) (syscall.Sockaddr, error) {
	// Only ipv4 address resolution is now implemented.
	switch network {
	case "tcp4":
		ip, port, err := parseIpv4Addr(addr)
		if err != nil {
			return nil, err
		}
		sa := &syscall.SockaddrInet4{Port: port}
		copy(sa.Addr[:], ip[:4])
		return sa, nil
	}

	return nil, &UnknownNetworkError{network, nil}
}

func resolveSockaddrInfo(sa syscall.Sockaddr) (network, addr string, port int, err error) {
	switch v := sa.(type) {
	case *syscall.SockaddrInet4:
		return "tcp4", net.IP(v.Addr[:]).String(), v.Port, nil
	}
	return "", "", -1, &UnknownNetworkError{}
}

func parseIpv4Addr(addr string) (ip net.IP, port int, err error) {
	ipStr, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	ip = net.ParseIP(ipStr).To4()
	port, err = strconv.Atoi(portStr)
	if err != nil {
		return
	}
	return
}
