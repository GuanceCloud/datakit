// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"net"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type socketCaptureFeeder struct {
	pointsCh chan []*point.Point
}

func newSocketCaptureFeeder() *socketCaptureFeeder {
	return &socketCaptureFeeder{
		pointsCh: make(chan []*point.Point, 4),
	}
}

func (f *socketCaptureFeeder) Feed(_ point.Category, pts []*point.Point, _ ...dkio.FeedOption) error {
	f.pointsCh <- pts
	return nil
}

func (f *socketCaptureFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func waitSocketPoints(t *testing.T, f *socketCaptureFeeder) []*point.Point {
	t.Helper()

	select {
	case pts := <-f.pointsCh:
		return pts
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for socket log points")
		return nil
	}
}

// TestSocketLogger 测试 SocketLogger 基本功能
func TestSocketLogger(t *testing.T) {
	opts := []Option{
		WithSource("testing"),
		WithSockets([]string{"tcp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	assert.Equal(t, "testing", sk.cfg.source)
	assert.NotNil(t, sk.log)
	assert.NotNil(t, sk.tags)
}

// TestSocketLoggerWithOptions 测试带选项的 SocketLogger
func TestSocketLoggerWithOptions(t *testing.T) {
	opts := []Option{
		WithSource("test-source"),
		WithCharacterEncoding("utf-8"),
		WithExtraTags(map[string]string{"env": "test"}),
		WithSockets([]string{"tcp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	assert.Equal(t, "test-source", sk.cfg.source)
	assert.Equal(t, "utf-8", sk.cfg.characterEncoding)
	assert.Equal(t, "test", sk.tags["env"])
}

// TestSocketLoggerLifecycle 测试 SocketLogger 生命周期
func TestSocketLoggerLifecycle(t *testing.T) {
	opts := []Option{
		WithSource("testing"),
		WithSockets([]string{"tcp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	// 启动 SocketLogger
	go func() {
		sk.Start()
	}()

	// 等待启动
	time.Sleep(10 * time.Millisecond)

	// 停止 SocketLogger
	sk.Close()

	// 验证已停止
	assert.NotNil(t, sk.cancel)
}

// TestSocketLoggerWithTCP 测试 TCP 连接
func TestSocketLoggerWithTCP(t *testing.T) {
	opts := []Option{
		WithSource("testing"),
		WithSockets([]string{"tcp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	// 启动服务器
	go func() {
		sk.Start()
	}()

	// 等待服务器启动
	time.Sleep(50 * time.Millisecond)

	// 停止服务器
	sk.Close()
}

// TestSocketLoggerWithUDP 测试 UDP 连接
func TestSocketLoggerWithUDP(t *testing.T) {
	opts := []Option{
		WithSource("testing"),
		WithSockets([]string{"udp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	// 启动服务器
	go func() {
		sk.Start()
	}()

	// 等待服务器启动
	time.Sleep(50 * time.Millisecond)

	// 停止服务器
	sk.Close()
}

func TestSocketLoggerCollectorSourceIP(t *testing.T) {
	t.Run("tcp", func(t *testing.T) {
		feeder := newSocketCaptureFeeder()
		sk, err := NewSocketLogging(
			WithSource("testing"),
			WithSockets([]string{"tcp://127.0.0.1:0"}),
			WithFeeder(feeder),
		)
		require.NoError(t, err)
		require.NotNil(t, sk)
		t.Cleanup(sk.Close)

		srv, ok := sk.servers[0].(*tcpServer)
		require.True(t, ok)

		sk.Start()

		conn, err := net.Dial("tcp", srv.listener.Addr().String())
		require.NoError(t, err)
		_, err = conn.Write([]byte("tcp message\n"))
		require.NoError(t, err)
		require.NoError(t, conn.Close())

		pts := waitSocketPoints(t, feeder)
		require.Len(t, pts, 1)
		assert.Equal(t, "127.0.0.1", pts[0].GetTag(collectorSourceIPTag))
	})

	t.Run("udp", func(t *testing.T) {
		feeder := newSocketCaptureFeeder()
		sk, err := NewSocketLogging(
			WithSource("testing"),
			WithSockets([]string{"udp://127.0.0.1:0"}),
			WithFeeder(feeder),
		)
		require.NoError(t, err)
		require.NotNil(t, sk)
		t.Cleanup(sk.Close)

		srv, ok := sk.servers[0].(*udpServer)
		require.True(t, ok)

		sk.Start()

		conn, err := net.Dial("udp", srv.conn.LocalAddr().String())
		require.NoError(t, err)
		_, err = conn.Write([]byte("udp message\n"))
		require.NoError(t, err)
		require.NoError(t, conn.Close())

		pts := waitSocketPoints(t, feeder)
		require.Len(t, pts, 1)
		assert.Equal(t, "127.0.0.1", pts[0].GetTag(collectorSourceIPTag))
	})
}

// TestSocketLoggerSetup 测试 SocketLogger 设置
func TestSocketLoggerSetup(t *testing.T) {
	opts := []Option{
		WithSource("testing"),
		WithSockets([]string{"tcp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	// 测试设置
	err = sk.setup()
	assert.NoError(t, err)
}

// TestSocketLoggerMakeServer 测试创建服务器
func TestSocketLoggerMakeServer(t *testing.T) {
	opts := []Option{
		WithSource("testing"),
		WithSockets([]string{"tcp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	// 测试创建服务器
	err = sk.makeServer()
	assert.NoError(t, err)
}

// TestSocketLoggerFeed 测试消息处理
func TestSocketLoggerFeed(t *testing.T) {
	opts := []Option{
		WithSource("testing"),
		WithSockets([]string{"tcp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	// 测试消息处理
	pending := [][]byte{
		[]byte("test message 1"),
		[]byte("test message 2"),
	}

	sk.feed(pending)
}

// TestSocketLoggerClose 测试关闭
func TestSocketLoggerClose(t *testing.T) {
	opts := []Option{
		WithSource("testing"),
		WithSockets([]string{"tcp://127.0.0.1:0"}),
	}

	sk, err := NewSocketLogging(opts...)
	require.NoError(t, err)
	require.NotNil(t, sk)

	// 测试关闭
	sk.Close()
	// 注意：cancel 可能为 nil，这是正常的
}

func TestSocketLoggerMultilineValidationByEnabled(t *testing.T) {
	t.Run("disabled-skip-invalid-pattern", func(t *testing.T) {
		opts := []Option{
			WithSource("testing"),
			WithSockets([]string{"tcp://127.0.0.1:0"}),
			EnableMultiline(false),
			WithMultilinePattern(`(?invalid`),
		}

		sk, err := NewSocketLogging(opts...)
		require.NoError(t, err)
		require.NotNil(t, sk)
		sk.Close()
	})

	t.Run("enabled-validate-invalid-pattern", func(t *testing.T) {
		opts := []Option{
			WithSource("testing"),
			WithSockets([]string{"tcp://127.0.0.1:0"}),
			EnableMultiline(true),
			WithMultilinePattern(`(?invalid`),
		}

		sk, err := NewSocketLogging(opts...)
		require.Error(t, err)
		assert.Nil(t, sk)
	})
}
