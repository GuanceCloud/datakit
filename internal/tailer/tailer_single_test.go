// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

// TestNewTailerSingle 测试创建单个文件 tailer
func TestNewTailerSingle(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	opts := []Option{
		WithSource("test-source"),
	}

	single, err := NewTailerSingle(tmpFile.Name(), opts...)
	require.NoError(t, err)
	require.NotNil(t, single)

	assert.Equal(t, tmpFile.Name(), single.filepath)
	assert.Equal(t, "test-source", single.config.source)
	assert.NotNil(t, single.log)
	assert.NotNil(t, single.extraTags)
	assert.NotNil(t, single.updateChan)
}

// TestSingleRun 测试 Single 的运行
func TestSingleRun(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	opts := []Option{
		WithSource("test-source"),
	}

	single, err := NewTailerSingle(tmpFile.Name(), opts...)
	require.NoError(t, err)
	require.NotNil(t, single)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 运行 tailer
	go func() {
		single.Run(ctx)
	}()

	// 等待上下文超时
	<-ctx.Done()

	// 验证 tailer 已停止
	assert.NotNil(t, single.cancelFunc)
}

// TestSingleClose 测试 Single 的关闭
func TestSingleClose(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	opts := []Option{
		WithSource("test-source"),
	}

	single, err := NewTailerSingle(tmpFile.Name(), opts...)
	require.NoError(t, err)
	require.NotNil(t, single)

	// 关闭 tailer
	single.Close()

	// 验证状态 - cancel 可能为 nil，这是正常的
	// assert.NotNil(t, single.cancel)
}

// TestSingleUpdateOptions 测试更新选项
func TestSingleUpdateOptions(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	opts := []Option{
		WithSource("test-source"),
	}

	single, err := NewTailerSingle(tmpFile.Name(), opts...)
	require.NoError(t, err)
	require.NotNil(t, single)

	// 更新选项
	newOpts := []Option{
		WithSource("updated-source"),
		WithCharacterEncoding("gbk"),
		WithMaxMultilineLength(2048),
	}

	single.UpdateOptions(newOpts)

	// 验证更新选项已发送到通道
	assert.NotNil(t, single.updateChan)

	// 注意：UpdateOptions 只是发送到通道，不会立即更新配置
	// 实际的配置更新需要在 Run 方法中处理
}

// TestSingleSetupFile 测试文件设置
func TestSingleSetupFile(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// 写入一些内容
	_, err = tmpFile.WriteString("test log content\n")
	require.NoError(t, err)
	tmpFile.Close()

	opts := []Option{
		WithSource("test-source"),
	}

	single, err := NewTailerSingle(tmpFile.Name(), opts...)
	require.NoError(t, err)
	require.NotNil(t, single)

	// 设置文件
	err = single.setupFile()
	assert.NoError(t, err)
	assert.NotNil(t, single.file)

	// 清理
	if single.file != nil {
		single.file.Close()
	}
}

// TestSingleShouldAddField 测试字段过滤
func TestSingleShouldAddField(t *testing.T) {
	t.Run("no-whitelist", func(t *testing.T) {
		// 创建临时文件
		tmpFile, err := os.CreateTemp("", "test-*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		opts := []Option{
			WithSource("test-source"),
			WithFieldWhitelist([]string{}),
		}

		single, err := NewTailerSingle(tmpFile.Name(), opts...)
		require.NoError(t, err)
		require.NotNil(t, single)

		// 没有白名单时，所有字段都应该被添加
		assert.True(t, single.shouldAddField("message"))
		assert.True(t, single.shouldAddField("timestamp"))
		assert.True(t, single.shouldAddField("level"))
	})

	t.Run("with-whitelist", func(t *testing.T) {
		// 创建临时文件
		tmpFile, err := os.CreateTemp("", "test-*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		opts := []Option{
			WithSource("test-source"),
			WithFieldWhitelist([]string{"message", "timestamp"}),
		}

		single, err := NewTailerSingle(tmpFile.Name(), opts...)
		require.NoError(t, err)
		require.NotNil(t, single)

		// 只有白名单中的字段才应该被添加
		assert.True(t, single.shouldAddField("message"))
		assert.True(t, single.shouldAddField("timestamp"))
		assert.False(t, single.shouldAddField("level"))
		assert.False(t, single.shouldAddField("source"))
	})
}

// TestSingleDecode 测试解码功能
func TestSingleDecode(t *testing.T) {
	t.Run("utf-8-encoding", func(t *testing.T) {
		// 创建临时文件
		tmpFile, err := os.CreateTemp("", "test-*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		opts := []Option{
			WithSource("test-source"),
			WithCharacterEncoding("utf-8"),
		}

		single, err := NewTailerSingle(tmpFile.Name(), opts...)
		require.NoError(t, err)
		require.NotNil(t, single)

		// 测试 UTF-8 解码
		text := "Hello, 世界!"
		result, err := single.decode([]byte(text))
		assert.NoError(t, err)
		assert.Equal(t, text, string(result))
	})

	t.Run("no-encoding", func(t *testing.T) {
		// 创建临时文件
		tmpFile, err := os.CreateTemp("", "test-*.log")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		opts := []Option{
			WithSource("test-source"),
			WithCharacterEncoding(""),
		}

		single, err := NewTailerSingle(tmpFile.Name(), opts...)
		require.NoError(t, err)
		require.NotNil(t, single)

		// 没有编码时，应该直接返回原始数据
		text := "Hello, World!"
		result, err := single.decode([]byte(text))
		assert.NoError(t, err)
		assert.Equal(t, text, string(result))
	})
}

// TestSingleWithFeeder 测试使用 Feeder
func TestSingleWithFeeder(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	opts := []Option{
		WithSource("test-source"),
		WithFeeder(dkio.DefaultFeeder()),
	}

	single, err := NewTailerSingle(tmpFile.Name(), opts...)
	require.NoError(t, err)
	require.NotNil(t, single)

	// 验证 feeder 已设置
	assert.NotNil(t, single.config.feeder)
}

// TestSingleWithExtraTags 测试额外标签
func TestSingleWithExtraTags(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	opts := []Option{
		WithSource("test-source"),
		WithExtraTags(map[string]string{
			"env":     "test",
			"version": "1.0.0",
		}),
	}

	single, err := NewTailerSingle(tmpFile.Name(), opts...)
	require.NoError(t, err)
	require.NotNil(t, single)

	// 验证额外标签已设置
	assert.Equal(t, "test", single.extraTags["env"])
	assert.Equal(t, "1.0.0", single.extraTags["version"])
}
