// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build darwin || netbsd || freebsd || openbsd || dragonfly
// +build darwin netbsd freebsd openbsd dragonfly

package ws

import (
	"errors"
	"net"
	"sync"
	"syscall"
)

type epoll struct {
	fd          int
	ts          syscall.Timespec
	connections map[int]net.Conn
	changes     []syscall.Kevent_t
	lock        *sync.RWMutex
}

func MkEpoll() (*epoll, error) {
	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	_, err = syscall.Kevent(fd, []syscall.Kevent_t{
		{
			Ident:  0,
			Filter: syscall.EVFILT_USER,
			Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
		},
	}, nil, nil)

	if err != nil {
		return nil, err
	}

	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		ts:          syscall.NsecToTimespec(1e9),
		connections: make(map[int]net.Conn),
	}, nil
}

func (e *epoll) Add(conn net.Conn) error {
	fd := websocketFD(conn)

	e.lock.Lock()
	defer e.lock.Unlock()

	e.changes = append(e.changes,
		syscall.Kevent_t{
			Ident:  uint64(fd),
			Flags:  syscall.EV_ADD | syscall.EV_EOF,
			Filter: syscall.EVFILT_READ,
		})
	e.connections[fd] = conn

	return nil
}

func (e *epoll) Remove(conn net.Conn) error {
	fd := websocketFD(conn)

	e.lock.Lock()
	defer e.lock.Unlock()

	if len(e.changes) <= 1 {
		e.changes = nil
	} else {
		changes := make([]syscall.Kevent_t, 0, len(e.changes)-1)
		ident := uint64(fd)
		for _, ke := range e.changes {
			if ke.Ident != ident {
				changes = append(changes, ke)
			}
		}

		e.changes = changes
	}

	delete(e.connections, fd)
	return nil
}

func (e *epoll) Wait(count int) ([]net.Conn, error) {
	events := make([]syscall.Kevent_t, count)

	e.lock.RLock()
	changes := e.changes
	e.lock.RUnlock()

retry:

	n, err := syscall.Kevent(e.fd, changes, events, &e.ts)
	if err != nil {
		if errors.Is(err, syscall.EINTR) {
			goto retry
		}

		l.Error("Wait() error: %s", err.Error())
		return nil, err
	}

	connections := make([]net.Conn, 0, n)
	e.lock.RLock()

	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Ident)]
		if (events[i].Flags & syscall.EV_EOF) == syscall.EV_EOF {
			if err := conn.Close(); err != nil {
				l.Warnf("Close: %s, ignored", err)
			}
		}
		connections = append(connections, conn)
	}

	e.lock.RUnlock()

	return connections, nil
}

func (e *epoll) Close() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.connections = nil
	e.changes = nil
	l.Debugf("epoll closed")
	return syscall.Close(e.fd)
}
