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
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/encoding"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/reader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
)

var socketGoroutine = datakit.G("socketLog")

type SocketLogger struct {
	cfg      *config // 配置信息
	feedName string  // 数据源名称

	servers []server          // 服务器列表
	tags    map[string]string // 标签映射

	cancel context.CancelFunc // 取消函数
	log    *logger.Logger     // 日志记录器
}

func NewSocketLogging(opts ...Option) (*SocketLogger, error) {
	sk := &SocketLogger{
		cfg: buildConfig(opts),
	}

	sk.feedName = dkio.FeedSource("socketLog", sk.cfg.source)
	if sk.cfg.storageIndex != "" {
		sk.feedName = dkio.FeedSource(sk.feedName, sk.cfg.storageIndex)
	}

	sk.tags = buildTags(sk.cfg.extraTags)
	sk.log = logger.SLogger("socketLog/" + sk.cfg.source)

	if err := sk.setup(); err != nil {
		sk.closeServers()
		return nil, err
	}

	return sk, nil
}

func (sk *SocketLogger) setup() error {
	if len(sk.cfg.sockets) == 0 {
		sk.log.Warnf("logging sockets is empty")
		return nil
	}

	if _, err := multiline.New(sk.cfg.multilinePatterns, multiline.WithMaxLength(int(sk.cfg.maxMultilineLength))); err != nil {
		sk.log.Warn(err)
		return err
	}

	if sk.cfg.characterEncoding != "" {
		if _, err := encoding.NewDecoder(sk.cfg.characterEncoding); err != nil {
			sk.log.Warnf("newdecoder err: %s", err)
			return err
		}
	}

	if err := sk.makeServer(); err != nil {
		sk.log.Warn(err)
		return err
	}

	return nil
}

func (sk *SocketLogger) makeServer() error {
	for _, socket := range sk.cfg.sockets {
		u, err := url.Parse(socket)
		if err != nil {
			return fmt.Errorf("error socket config err=%w", err)
		}

		scheme := u.Scheme
		address := u.Host

		sk.log.Debugf("check logging socket Scheme=%s listenerAddr=%s", scheme, address)

		// default use TCP
		if scheme == "" {
			scheme = "tcp" // nolint:goconst
		}

		switch scheme {
		case "tcp", "tcp4", "tcp6":
			srv, err := newTCPServer(scheme, address, sk.cfg)
			if err != nil {
				return fmt.Errorf("%s-socket listen port error: %w", scheme, err)
			}
			sk.servers = append(sk.servers, srv)

		case "udp", "udp4", "udp6":
			srv, err := newUDPServer(scheme, address, sk.cfg)
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
			"message_length":       len(cnt),
			constants.FieldMessage: string(cnt),
			constants.FieldStatus:  pipeline.DefaultStatus,
		}

		pt := point.NewPoint(
			sk.cfg.source,
			append(point.NewTags(sk.tags), point.NewKVs(fields)...),
			point.DefaultLoggingOptions()...,
		)
		pts = append(pts, pt)
	}

	if len(pts) == 0 {
		return
	}

	if err := sk.cfg.feeder.Feed(point.Logging, pts,
		dkio.WithSource(sk.feedName),
		dkio.WithStorageIndex(sk.cfg.storageIndex),
		dkio.WithPipelineOption(&lang.LogOption{
			ScriptMap: map[string]string{sk.cfg.source: sk.cfg.pipeline},
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
	cfg      *config
}

func newTCPServer(scheme, address string, cfg *config) (*tcpServer, error) {
	listener, err := net.Listen(scheme, address)
	if err != nil {
		return nil, err
	}
	return &tcpServer{listener, cfg}, nil
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

		socketConnectCounter.WithLabelValues("tcp", "ok").Inc()
		socketGoroutine.Go(func(_ context.Context) error {
			defer conn.Close() // nolint

			rd := reader.NewReader(conn)
			// must not error
			mult, _ := multiline.New(s.cfg.multilinePatterns, multiline.WithMaxLength(int(s.cfg.maxMultilineLength)))

			var decoder *encoding.Decoder
			if s.cfg.characterEncoding != "" {
				// must not error
				decoder, _ = encoding.NewDecoder(s.cfg.characterEncoding)
			}

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

					text, err := decodingBytes(decoder, line)
					if err != nil {
						decodeErrorCounter.WithLabelValues(s.cfg.source, s.cfg.characterEncoding, err.Error()).Inc()
					}

					text = removeAnsiEscapeCodes(text, s.cfg.removeAnsiEscapeCodes)
					if mult != nil {
						text, _ = mult.ProcessLine(multiline.TrimRightSpace(text))
					}
					if len(text) == 0 {
						continue
					}
					pending = append(pending, text)
				}

				socketMessageCounter.WithLabelValues("tcp").Inc()
				socketLengthSummary.WithLabelValues("tcp").Observe(float64(len(pending)))
				feed(pending)
			}

			if mult != nil && mult.BuffLength() > 0 {
				b := mult.Flush()
				socketMessageCounter.WithLabelValues("tcp").Inc()
				socketLengthSummary.WithLabelValues("tcp").Observe(float64(1))
				feed([][]byte{b})
			}
			return nil
		})
	}
}

type udpServer struct {
	conn net.Conn
	cfg  *config
}

func newUDPServer(scheme, address string, cfg *config) (*udpServer, error) {
	udpAddr, err := net.ResolveUDPAddr(scheme, address)
	if err != nil {
		return nil, fmt.Errorf("resolve UDP addr error:%w", err)
	}

	conn, err := net.ListenUDP(scheme, udpAddr)
	if err != nil {
		return nil, err
	}

	return &udpServer{conn, cfg}, nil
}

func (s *udpServer) close() error {
	return s.conn.Close()
}

func (s *udpServer) forwardMessage(ctx context.Context, feed func([][]byte)) error {
	defer s.conn.Close() // nolint

	rd := reader.NewReader(s.conn, reader.DisablePreviousBlock())
	var decoder *encoding.Decoder
	if s.cfg.characterEncoding != "" {
		// must not error
		decoder, _ = encoding.NewDecoder(s.cfg.characterEncoding)
	}

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

		var pending [][]byte
		for _, line := range lines {
			if len(lines) == 0 {
				continue
			}

			text, err := decodingBytes(decoder, line)
			if err != nil {
				decodeErrorCounter.WithLabelValues(s.cfg.source, s.cfg.characterEncoding, err.Error()).Inc()
			}

			text = removeAnsiEscapeCodes(text, s.cfg.removeAnsiEscapeCodes)
			pending = append(pending, text)
		}

		socketMessageCounter.WithLabelValues("udp").Inc()
		socketLengthSummary.WithLabelValues("udp").Observe(float64(len(lines)))
		feed(pending)
	}
}

func decodingBytes(decoder *encoding.Decoder, text []byte) ([]byte, error) {
	if decoder == nil {
		return text, nil
	}
	return decoder.Bytes(text)
}
