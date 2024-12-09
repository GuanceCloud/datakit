// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tailer read socket logging
package tailer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/reader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
)

var socketGoroutine = datakit.G("socketLog")

type SocketLogger struct {
	opt *option

	servers []server
	tags    map[string]string

	cancel context.CancelFunc
	log    *logger.Logger
}

func NewSocketLogWithOptions(opts ...Option) (*SocketLogger, error) {
	c := getOption(opts...)
	sk := &SocketLogger{
		opt: c,
	}
	sk.tags = buildTags(sk.opt.extraTags)
	sk.log = logger.SLogger("socketLog/" + sk.opt.source)

	if err := sk.setup(); err != nil {
		sk.closeServers()
		return nil, err
	}

	return sk, nil
}

func (sk *SocketLogger) setup() error {
	if len(sk.opt.sockets) == 0 {
		sk.log.Warnf("logging sockets is empty")
		return nil
	}

	if _, err := multiline.New(sk.opt.multilinePatterns,
		multiline.WithMaxLifeDuration(sk.opt.maxMultilineLifeDuration)); err != nil {
		sk.log.Warn(err)
		return err
	}
	if err := sk.makeServer(); err != nil {
		sk.log.Warn(err)
		return err
	}

	return nil
}

func (sk *SocketLogger) makeServer() error {
	for _, socket := range sk.opt.sockets {
		u, err := url.Parse(socket)
		if err != nil {
			return fmt.Errorf("error socket config err=%w", err)
		}

		scheme := u.Scheme
		address := u.Host

		sk.log.Debugf("check logging socket Scheme=%s listenerAddr=%s", scheme, address)

		// default use TCP
		if scheme == "" {
			scheme = "tcp"
		}

		switch scheme {
		case "tcp", "tcp4", "tcp6":
			srv, err := newTCPServer(scheme, address, sk.opt)
			if err != nil {
				return fmt.Errorf("%s-socket listen port error: %w", scheme, err)
			}
			sk.servers = append(sk.servers, srv)

		case "udp", "udp4", "udp6":
			srv, err := newUDPServer(scheme, address)
			if err != nil {
				return fmt.Errorf("%s-socket listen port error: %w", scheme, err)
			}
			sk.servers = append(sk.servers, srv)

		default:
			return fmt.Errorf("socket config like this: socket=[tcp://127.0.0.1:9540] , and please check your logging.conf")
		}
	}

	return nil
}

func (sk *SocketLogger) Start() {
	if len(sk.servers) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	sk.cancel = cancel

	for _, srv := range sk.servers {
		func(s server) {
			socketGoroutine.Go(func(_ context.Context) error {
				if err := s.forwardMessage(ctx, sk.feed); err != nil {
					sk.log.Warn(err)
				}
				return nil
			})
		}(srv)
	}
}

func (sk *SocketLogger) feed(pending [][]byte) {
	pts := []*point.Point{}
	for _, cnt := range pending {
		if len(cnt) == 0 {
			continue
		}
		fields := map[string]interface{}{
			"message_length":      len(cnt),
			pipeline.FieldMessage: string(cnt),
			pipeline.FieldStatus:  pipeline.DefaultStatus,
		}

		pt := point.NewPointV2(
			sk.opt.source,
			append(point.NewTags(sk.tags), point.NewKVs(fields)...),
			point.DefaultLoggingOptions()...,
		)
		pts = append(pts, pt)
	}

	if len(pts) == 0 {
		return
	}

	if err := sk.opt.feeder.FeedV2(point.Logging, pts,
		dkio.WithInputName("socketLog/"+sk.opt.source),
		dkio.WithPipelineOption(&manager.Option{
			DisableAddStatusField: sk.opt.disableAddStatusField,
			IgnoreStatus:          sk.opt.ignoreStatus,
			ScriptMap:             map[string]string{sk.opt.source: sk.opt.pipeline},
		}),
	); err != nil {
		sk.log.Errorf("feed %d pts failed: %s, logging block-mode off, ignored", len(pts), err)
	}
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

func (sk *SocketLogger) Close() {
	if sk.cancel != nil {
		sk.cancel()
	}
	sk.closeServers()
	sk.log.Info("closed all")
}

func (sk *SocketLogger) closeServers() {
	for _, srv := range sk.servers {
		if srv != nil {
			if err := srv.close(); err != nil {
				sk.log.Warnf("closing connect fail %s", err)
			}
			continue
		}
	}
}

type server interface {
	forwardMessage(context.Context, func([][]byte)) error
	close() error
}

type tcpServer struct {
	listener net.Listener
	opt      *option
}

func newTCPServer(scheme, address string, opt *option) (*tcpServer, error) {
	listener, err := net.Listen(scheme, address)
	if err != nil {
		return nil, err
	}
	return &tcpServer{listener, opt}, nil
}

func (s *tcpServer) close() error {
	return s.listener.Close()
}

func (s *tcpServer) forwardMessage(ctx context.Context, feed func([][]byte)) error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return err
			}
			continue
		}

		socketLogConnect.WithLabelValues("tcp", "ok").Add(1)
		socketGoroutine.Go(func(_ context.Context) error {
			defer conn.Close() // nolint

			rd := reader.NewReader(conn)
			// must not error
			mult, _ := multiline.New(s.opt.multilinePatterns,
				multiline.WithMaxLifeDuration(s.opt.maxMultilineLifeDuration))

			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					// next
				}

				lines, _, err := rd.ReadLines()
				if err != nil {
					break
				}

				var pending [][]byte

				for _, line := range lines {
					if len(line) == 0 {
						continue
					}

					text, _ := mult.ProcessLine(multiline.TrimRightSpace(line))
					if len(text) == 0 {
						continue
					}
					pending = append(pending, text)
				}

				socketLogCount.WithLabelValues("tcp").Add(1)
				socketLogLength.WithLabelValues("tcp").Observe(float64(len(pending)))
				feed(pending)
			}

			if mult.BuffLength() > 0 {
				b := mult.Flush()
				socketLogCount.WithLabelValues("tcp").Add(1)
				socketLogLength.WithLabelValues("tcp").Observe(float64(1))
				feed([][]byte{b})
			}
			return nil
		})
	}
}

type udpServer struct {
	conn net.Conn
}

func newUDPServer(scheme, address string) (*udpServer, error) {
	udpAddr, err := net.ResolveUDPAddr(scheme, address)
	if err != nil {
		return nil, fmt.Errorf("resolve UDP addr error:%w", err)
	}

	conn, err := net.ListenUDP(scheme, udpAddr)
	if err != nil {
		return nil, err
	}

	return &udpServer{conn}, nil
}

func (s *udpServer) close() error {
	return s.conn.Close()
}

func (s *udpServer) forwardMessage(ctx context.Context, feed func([][]byte)) error {
	defer s.conn.Close() // nolint
	rd := reader.NewReader(s.conn, reader.DisablePreviousBlock())

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// next
		}

		lines, _, err := rd.ReadLines()
		if err != nil {
			if errors.Is(err, reader.ErrReadEmpty) {
				continue
			}
			return err
		}

		if len(lines) == 0 {
			continue
		}

		socketLogCount.WithLabelValues("udp").Add(1)
		socketLogLength.WithLabelValues("udp").Observe(float64(len(lines)))
		feed(lines)
	}
}
