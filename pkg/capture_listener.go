package gont

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"go.uber.org/zap"
)

var _ io.Writer = (*captureListener)(nil)

type captureListener struct {
	listener net.Listener

	Conns chan net.Conn

	connsLock sync.RWMutex
	conns     map[net.Conn]any
	logger    *zap.Logger
}

func newCaptureListener(addr string) (*captureListener, error) {
	var network, address string
	if addrParts := strings.SplitN(addr, ":", 2); len(addrParts) == 1 {
		network = "tcp"
		address = addrParts[0]
	} else {
		network = addrParts[0]
		address = addrParts[1]
	}

	lst, err := net.Listen(network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	cs := &captureListener{
		listener: lst,
		Conns:    make(chan net.Conn),
		conns:    map[net.Conn]any{},
		logger: zap.L().Named("capture.socket").With(
			zap.String("addr", addr),
		),
	}

	go cs.listen()

	return cs, nil
}

func (cs *captureListener) Close() error {
	return cs.listener.Close()
}

func (cs *captureListener) Write(b []byte) (n int, err error) {
	cs.connsLock.RLock()
	defer cs.connsLock.RUnlock()

	for c := range cs.conns {
		if n, err = c.Write(b); err != nil {
			logger := cs.logger.With(zap.String("remote", c.RemoteAddr().String()))
			if errors.Is(err, net.ErrClosed) {
				logger.Debug("Connection closed")
			} else {
				logger.Error("Failed to write to connection. Closing...", zap.Error(err))

				if err := c.Close(); err != nil {
					logger.Error("Failed to close connection")
				}
			}

			cs.connsLock.RUnlock()
			cs.removeConn(c)
			cs.connsLock.RLock()
		}
	}

	return n, nil
}

func (cs *captureListener) listen() {
	for {
		c, err := cs.listener.Accept()
		if err != nil {
			continue
		}

		cs.logger.Debug("New connection",
			zap.String("remote", c.RemoteAddr().String()))

		select {
		case cs.Conns <- c:
		default:
		}

		cs.addConn(c)
	}
}

func (cs *captureListener) addConn(c net.Conn) {
	cs.connsLock.Lock()
	cs.conns[c] = nil
	cs.connsLock.Unlock()
}

func (cs *captureListener) removeConn(c net.Conn) {
	cs.connsLock.Lock()
	delete(cs.conns, c)
	cs.connsLock.Unlock()
}
