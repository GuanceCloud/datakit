// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tailer read socket logging
package tailer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	plmanager "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
)

const (
	// 32Mb is max Logging Point field len.
	maxReadBufferLen = 1024 * 1024 * 32
)

var g = datakit.G("tailer")

type server struct {
	addr string       // udpConns's or tcpListeners's  key
	lis  net.Listener // tcp listener
	conn net.Conn     // udp.Conn
}

type socketLogger struct {
	// 存放连接，释放连接使用
	tcpListeners    map[string]net.Listener
	udpConns        map[string]net.Conn
	socketBufferLen int // read buffer lens
	ignorePatterns  []string
	tags            map[string]string
	// 配置
	opt  *Option
	stop chan struct{}

	servers []*server
	feeder  dkio.Feeder
	ptCache chan *point.Point
}

func NewWithOpt(opt *Option, ignorePatterns ...[]string) (sl *socketLogger, err error) {
	sl = &socketLogger{
		tcpListeners:    make(map[string]net.Listener),
		udpConns:        make(map[string]net.Conn),
		socketBufferLen: maxReadSize,
		ignorePatterns: func() []string {
			if len(ignorePatterns) > 0 {
				return ignorePatterns[0]
			}
			return nil
		}(),
		opt:     opt,
		stop:    make(chan struct{}, 1),
		ptCache: make(chan *point.Point, 100),
	}
	if err := sl.opt.Init(); err != nil {
		return nil, err
	}
	sl.tags = buildTags(opt.GlobalTags)

	l = logger.SLogger("socketLog")
	return sl, nil
}

func buildTags(globalTags map[string]string) map[string]string {
	tags := make(map[string]string)
	for k, v := range globalTags {
		tags[k] = v
	}
	if _, ok := tags["log_source"]; !ok {
		tags["log_source"] = "socket"
	}
	return tags
}

func (sl *socketLogger) Start() {
	if len(sl.opt.Sockets) == 0 {
		sl.opt.log.Warnf("logging sockets is empty")
		return
	}

	if sl.feeder == nil {
		sl.feeder = dkio.DefaultFeeder()
	}

	for _, socket := range sl.opt.Sockets {
		s, err := mkServer(socket)
		if err != nil {
			l.Error(err)
			return
		}
		sl.servers = append(sl.servers, s)
	}
	// 配置无误之后 开始accept.
	sl.toReceive()
	sl.feedV2()
}

func mkServer(socket string) (s *server, err error) {
	s = &server{addr: socket}
	socketURL, err := url.Parse(socket)
	if err != nil {
		return nil, fmt.Errorf("error socket config err=%w", err)
	}

	network := socketURL.Scheme
	listenAddr := socketURL.Host

	l.Debugf("check logging socket Scheme=%s listenerAddr=%s", network, listenAddr)

	switch network {
	case "", "tcp", "tcp4", "tcp6": // default use TCP
		listener, err := net.Listen(network, listenAddr)
		if err != nil {
			return nil, fmt.Errorf("socket listen port error:%w", err)
		}
		s.lis = listener

	case "udp", "udp4", "udp6":
		udpAddr, err := net.ResolveUDPAddr(network, listenAddr)
		if err != nil {
			return nil, fmt.Errorf("resolve UDP addr error:%w", err)
		}
		conn, err := net.ListenUDP(network, udpAddr)
		if err != nil {
			return nil, fmt.Errorf(" net.ListenUDP error:%w", err)
		}
		s.conn = conn

	default:
		return nil, fmt.Errorf("socket config like this: socket=[tcp://127.0.0.1:9540] , and please check your logging.conf")
	}

	return s, err
}

// toReceive: 根据listen或udp.conn 开始接收数据.
func (sl *socketLogger) toReceive() {
	if sl.servers == nil || len(sl.servers) == 0 {
		return
	}
	for _, s := range sl.servers {
		if s.lis != nil {
			sl.tcpListeners[s.addr] = s.lis
			l.Infof("TCP port:%s start to accept", s.addr)

			func(lis net.Listener) {
				g.Go(func(ctx context.Context) error {
					sl.accept(lis)
					return nil
				})
			}(s.lis)
		}
		if s.conn != nil {
			sl.udpConns[s.addr] = s.conn
			l.Infof("UDP port:%s start to accept", s.addr)

			func(conn net.Conn) {
				g.Go(func(ctx context.Context) error {
					sl.doSocketUDP(conn)
					return nil
				})
			}(s.conn)
		}
	}
}

func (sl *socketLogger) accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				l.Infof("tcp conn is close")
				return
			}
			sl.opt.log.Warnf("Error accepting:%s", err.Error())
			continue
		}
		socketLogConnect.WithLabelValues("tcp", "ok").Add(1)
		g.Go(func(ctx context.Context) error {
			sl.doSocketV2(conn)
			return nil
		})
	}
}

func (sl *socketLogger) doSocketUDP(conn net.Conn) {
	data := make([]byte, sl.socketBufferLen)
	for {
		n, err := conn.Read(data)
		// see:$GOROOT/src/io/io.go:83
		if err != nil && n == 0 {
			l.Errorf("err not nil err=%v", err)
			return
		}
		socketLogConnect.WithLabelValues("udp", "ok").Add(1)
		l.Debugf("data len =%d", n)
		pipDate := bytes.Split(data[:n], []byte{'\n'})
		for _, s := range pipDate {
			if len(s) > 0 {
				sl.makeAndFeedPoint("udp", s)
			}
		}
	}
}

func (sl *socketLogger) doSocketV2(conn net.Conn) {
	defer conn.Close() //nolint
	reader := bufio.NewReaderSize(conn, sl.socketBufferLen)
	cacheBts := make([]byte, 0)
	for {
		bts, full, err := reader.ReadLine()
		if err != nil {
			l.Warnf("readline err=%v", err)
			socketLogConnect.WithLabelValues("tcp", "eof").Add(1)
			break
		}
		if full {
			cacheBts = append(cacheBts, bts...)
			if len(cacheBts) >= maxReadBufferLen {
				sl.makeAndFeedPoint("tcp", cacheBts)
				cacheBts = cacheBts[:0]
			}
			continue
		}
		if len(bts) == 0 {
			continue
		}
		l.Debugf("readline len=%d", len(bts))
		if len(cacheBts) > 0 {
			bts = append(cacheBts, bts...)
			cacheBts = cacheBts[:0]
		}
		sl.makeAndFeedPoint("tcp", bts)
	}
}

func (sl *socketLogger) makeAndFeedPoint(network string, line []byte) {
	fieles := map[string]interface{}{pipeline.FieldMessage: string(line), pipeline.FieldStatus: pipeline.DefaultStatus}

	pt := point.NewPointV2(sl.opt.Source,
		append(point.NewTags(sl.tags), point.NewKVs(fieles)...),
		point.WithTime(time.Now()))
	socketLogCount.WithLabelValues(network).Add(1)
	socketLogLength.WithLabelValues(network).Observe(float64(len(line)))
	sl.ptCache <- pt
}

func (sl *socketLogger) feedV2() {
	var ioOpt *dkio.Option
	if sl.opt.Pipeline != "" {
		ioOpt = &dkio.Option{
			PlOption: &plmanager.Option{
				DisableAddStatusField: sl.opt.DisableAddStatusField,
				IgnoreStatus:          sl.ignorePatterns,
				ScriptMap:             map[string]string{sl.opt.Source: sl.opt.Pipeline},
			},
		}
	}
	ticker := time.NewTicker(time.Second * 5)
	pts := make([]*point.Point, 0)
	sendAndReset := func() {
		if err := sl.feeder.Feed("socklogging/"+sl.opt.InputName, point.Logging, pts, ioOpt); err != nil {
			l.Error(err)
		}
		// 发送成功与否，重置数组，否则内存会一直涨。
		pts = pts[:0]
	}

	for {
		select {
		case <-ticker.C:
			if len(pts) != 0 {
				sendAndReset()
			}
		case pt := <-sl.ptCache:
			pts = append(pts, pt)
			if len(pts) >= 100 {
				sendAndReset()
			}
		case <-sl.stop:
			return
		}
	}
}

func (sl *socketLogger) Close() {
	sl.stop <- struct{}{}
	for _, listener := range sl.tcpListeners {
		err := listener.Close()
		sl.opt.log.Infof("close tcp port err=%v", err)
	}
	for _, listener := range sl.udpConns {
		err := listener.Close()
		sl.opt.log.Infof("close udp port err=%v", err)
	}
}
