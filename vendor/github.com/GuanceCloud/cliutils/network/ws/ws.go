// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ws wraps websocket implements among UNIX-like(Linux & macOS) platform
package ws

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var (
	l             = logger.DefaultSLogger("ws")
	CommonChanCap = 128
)

type Server struct {
	Path string
	Bind string

	MsgHandler func(*Server, net.Conn, []byte, ws.OpCode) error // server msg handler
	AddCli     func(w http.ResponseWriter, r *http.Request)

	uptime time.Time

	exit *cliutils.Sem
	wg   *sync.WaitGroup

	epoller *epoll
}

func NewServer(bind, path string) (s *Server, err error) {
	s = &Server{
		Path: path,
		Bind: bind,

		uptime: time.Now(),

		exit: cliutils.NewSem(),
		wg:   &sync.WaitGroup{},
	}

	s.epoller, err = MkEpoll()
	if err != nil {
		l.Error("MkEpoll() error: %s", err.Error())
		return
	}

	return
}

func (s *Server) AddConnection(conn net.Conn) error {
	if err := s.epoller.Add(conn); err != nil {
		l.Errorf("epoll.Add() error: %s", err.Error())

		if err := conn.Close(); err != nil {
			l.Warnf("Close: %s, ignored", err)
		}
		return err
	}

	return nil
}

func SendMsgToClient(msg []byte, conn net.Conn) error {
	return wsutil.WriteServerText(conn, msg)
}

func (s *Server) Stop() {
	s.exit.Close()

	l.Debug("wait...")
	s.wg.Wait()

	l.Debug("wait done")
}

func (s *Server) Start() {
	l = logger.SLogger("ws")

	// remove resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	l.Debugf("rLimit cur: %d, max: %d", rLimit.Cur, rLimit.Max)

	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		fmt.Printf("warn: Setrlimit %+#v failed: %s\n", rLimit, err.Error())
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.startEpoll()
	}()

	srv := &http.Server{
		Addr: s.Bind,
	}

	if s.AddCli == nil {
		l.Fatal("AddCli not set")
	}

	http.HandleFunc(s.Path, s.AddCli)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			l.Info(err)
		}
	}()

	<-s.exit.Wait()
	if err := srv.Shutdown(context.TODO()); err != nil {
		l.Errorf("srv.Shutdown: %s", err.Error())
	}

	l.Info("websocket server stopped.")
}

func (s *Server) startEpoll() {
	for {
		select {
		case <-s.exit.Wait():
			l.Debug("epoll exit.")
			if err := s.epoller.Close(); err != nil {
				l.Warnf("Close: %s, ignored", err)
			}
			return

		default:

			connections, err := s.epoller.Wait(100)
			if err != nil {
				// You should check the epoll_wait return value,
				// then if it's -1 compare errno to EINTR and,
				// if so, retry the system call.
				// This is usually done with continue in a loop.
				continue
			}

			for _, conn := range connections {
				if conn == nil {
					break
				}

				if data, opcode, err := wsutil.ReadClientData(conn); err != nil {
					l.Debugf("ReadClientData: %s", err.Error())

					if err := s.epoller.Remove(conn); err != nil {
						l.Errorf("Failed to remove %v", err)
					}

					l.Debugf("close cli %s", conn.RemoteAddr().String())
					if err := conn.Close(); err != nil {
						l.Warnf("Close: %s, ignored", err)
					}
				} else if s.MsgHandler != nil {
					if err := s.MsgHandler(s, conn, data, opcode); err != nil {
						l.Error("s.handler() error: %s", err.Error())
					}
				}
			}
		}
	}
}
