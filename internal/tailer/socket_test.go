// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
