// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func TestWithOptions(t *testing.T) {
	t.Run("with-service", func(t *testing.T) {
		cfg := defaultConfig()
		WithSource("testing-source")(cfg)
		WithService("testing-service")(cfg)

		res := map[string]string{"service": "testing-service"}
		assert.Equal(t, cfg.extraTags, res)
	})

	t.Run("with-default-service", func(t *testing.T) {
		cfg := defaultConfig()
		WithSource("testing-source")(cfg)
		WithService("")(cfg)

		res := map[string]string{"service": "testing-source"}
		assert.Equal(t, cfg.extraTags, res)
	})

	t.Run("with-default-service", func(t *testing.T) {
		cfg := defaultConfig()
		WithService("")(cfg)
		WithSource("testing-source")(cfg)

		res := map[string]string{"service": "default"}
		assert.Equal(t, cfg.extraTags, res)
	})

	t.Run("with-non-service", func(t *testing.T) {
		cfg := defaultConfig()
		WithSource("testing-source")(cfg)

		res := map[string]string{"service": "testing-source"}
		assert.Equal(t, cfg.extraTags, res)
	})
}

func TestBuildConfig(t *testing.T) {
	opts := []Option{
		WithSource("source"),
		WithService("service"),
		WithPipeline("pipeline.p"),
		WithStorageIndex("storageIndex"),
		WithCharacterEncoding("utf-8"),
		WithMaxMultilineLength(100),
		WithExtraTags(map[string]string{
			"key": "value",
			"abc": "123",
		}),
	}

	cfg := defaultConfig()
	cfg.source = "source"
	cfg.service = "service"
	cfg.pipeline = "pipeline.p"
	cfg.storageIndex = "storageIndex"
	cfg.characterEncoding = "utf-8"
	cfg.maxMultilineLength = 100
	cfg.extraTags = map[string]string{
		"service": "service",
		"key":     "value",
		"abc":     "123",
	}

	newCfg := buildConfig(opts)
	assert.Equal(t, cfg, newCfg)
}

func TestMergeOptions(t *testing.T) {
	opts1 := []Option{
		WithSource("source-1"),
		WithService("service-1"),
		WithPipeline("pipeline-1.p"),
		WithStorageIndex("storageIndex-1"),
		WithCharacterEncoding("utf-8"),
		WithMaxMultilineLength(100),
		WithExtraTags(map[string]string{
			"key": "value",
			"opt": "1",
		}),
	}
	opts2 := []Option{
		WithSource("source-2"),
		WithService("service-2"),
		WithPipeline("pipeline-2.p"),
		WithStorageIndex("storageIndex-2"),
		WithCharacterEncoding("utf-8"),
		WithMaxMultilineLength(200),
		WithExtraTags(map[string]string{
			"key": "value",
			"opt": "2",
		}),
		WithInsideFilepathFunc(nil),
	}

	newOpts := mergeOptions(opts1, opts2)

	expectedOpts := []Option{
		WithSockets(nil),
		WithIgnorePatterns(nil),
		WithSource("source-2"),
		WithService("service-2"),
		WithPipeline("pipeline-2.p"),
		WithStorageIndex("storageIndex-2"),
		WithCharacterEncoding("utf-8"),

		EnableDebugFields(false),
		EnableMultiline(false),
		WithMultilinePatterns(nil),
		WithMaxMultilineLength(200),

		WithMaxOpenFiles(500),
		WithIgnoreDeadLog(0),
		WithFileSizeThreshold(1000 * 1000 * 20),
		WithFromBeginning(false),
		WithRemoveAnsiEscapeCodes(false),

		WithExtraTags(map[string]string{
			"key": "value",
			"opt": "2",
		}),
		WithFieldWhitelist(nil),

		WithForwardFunc(nil),
		WithInsideFilepathFunc(nil),

		WithTextParserMode(0),
		WithFeeder(dkio.DefaultFeeder()),
	}

	expectedCfg := buildConfig(expectedOpts)
	actualCfg := buildConfig(newOpts)

	assert.Equal(t, expectedCfg, actualCfg)
}

// TestCheckConfig 测试配置验证功能
func TestCheckConfig(t *testing.T) {
	t.Run("valid-config", func(t *testing.T) {
		cfg := &config{
			source:             "test-source",
			characterEncoding:  "utf-8",
			maxOpenFiles:       100,
			fileSizeThreshold:  1024,
			maxMultilineLength: 1000,
			multilinePatterns:  []string{`^\d{4}-\d{2}-\d{2}`},
		}

		err := checkConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("invalid-character-encoding", func(t *testing.T) {
		cfg := &config{
			source:            "test-source",
			characterEncoding: "invalid-encoding",
		}

		err := checkConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character encoding")
	})

	t.Run("empty-source", func(t *testing.T) {
		cfg := &config{
			source: "",
		}

		err := checkConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source cannot be empty")
	})

	t.Run("invalid-max-open-files", func(t *testing.T) {
		cfg := &config{
			source:       "test-source",
			maxOpenFiles: -2,
		}

		err := checkConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maxOpenFiles must be >= -1")
	})

	t.Run("invalid-file-size-threshold", func(t *testing.T) {
		cfg := &config{
			source:            "test-source",
			fileSizeThreshold: -1,
		}

		err := checkConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fileSizeThreshold must be >= 0")
	})

	t.Run("invalid-multiline-length", func(t *testing.T) {
		cfg := &config{
			source:             "test-source",
			maxMultilineLength: -1,
		}

		err := checkConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maxMultilineLength must be >= 0")
	})
}

// TestOptionFunctions 测试各个选项函数
func TestOptionFunctions(t *testing.T) {
	t.Run("WithSockets", func(t *testing.T) {
		cfg := defaultConfig()
		sockets := []string{"tcp://localhost:8080", "udp://localhost:9090"}
		WithSockets(sockets)(cfg)
		assert.Equal(t, sockets, cfg.sockets)
	})

	t.Run("WithIgnorePatterns", func(t *testing.T) {
		cfg := defaultConfig()
		patterns := []string{"*.tmp", "*.log"}
		WithIgnorePatterns(patterns)(cfg)
		assert.Equal(t, patterns, cfg.ignorePatterns)
	})

	t.Run("WithPipeline", func(t *testing.T) {
		cfg := defaultConfig()
		pipeline := "test.p"
		WithPipeline(pipeline)(cfg)
		assert.Equal(t, pipeline, cfg.pipeline)
	})

	t.Run("WithStorageIndex", func(t *testing.T) {
		cfg := defaultConfig()
		index := "test-index"
		WithStorageIndex(index)(cfg)
		assert.Equal(t, index, cfg.storageIndex)
	})

	t.Run("WithCharacterEncoding", func(t *testing.T) {
		cfg := defaultConfig()
		encoding := "gbk"
		WithCharacterEncoding(encoding)(cfg)
		assert.Equal(t, encoding, cfg.characterEncoding)
	})

	t.Run("WithFromBeginning", func(t *testing.T) {
		cfg := defaultConfig()
		WithFromBeginning(true)(cfg)
		assert.True(t, cfg.fromBeginning)
	})

	t.Run("WithTextParserMode", func(t *testing.T) {
		cfg := defaultConfig()
		WithTextParserMode(DockerJSONLogMode)(cfg)
		assert.Equal(t, DockerJSONLogMode, cfg.mode)
	})

	t.Run("EnableDebugFields", func(t *testing.T) {
		cfg := defaultConfig()
		EnableDebugFields(true)(cfg)
		assert.True(t, cfg.enableDebugFields)
	})

	t.Run("WithFieldWhitelist", func(t *testing.T) {
		cfg := defaultConfig()
		whitelist := []string{"message", "timestamp"}
		WithFieldWhitelist(whitelist)(cfg)
		assert.Equal(t, whitelist, cfg.fieldWhitelist)
	})

	t.Run("EnableMultiline", func(t *testing.T) {
		cfg := defaultConfig()
		EnableMultiline(true)(cfg)
		assert.True(t, cfg.enableMultiline)
	})

	t.Run("WithMultilinePatterns", func(t *testing.T) {
		cfg := defaultConfig()
		patterns := []string{`^\d{4}-\d{2}-\d{2}`, `^ERROR`}
		WithMultilinePatterns(patterns)(cfg)
		assert.Equal(t, patterns, cfg.multilinePatterns)
	})

	t.Run("WithMaxMultilineLength", func(t *testing.T) {
		cfg := defaultConfig()
		length := int64(2048)
		WithMaxMultilineLength(length)(cfg)
		assert.Equal(t, length, cfg.maxMultilineLength)
	})

	t.Run("WithRemoveAnsiEscapeCodes", func(t *testing.T) {
		cfg := defaultConfig()
		WithRemoveAnsiEscapeCodes(true)(cfg)
		assert.True(t, cfg.removeAnsiEscapeCodes)
	})

	t.Run("WithMaxOpenFiles", func(t *testing.T) {
		cfg := defaultConfig()
		WithMaxOpenFiles(100)(cfg)
		assert.Equal(t, 100, cfg.maxOpenFiles)

		// 测试 -1 (无限制)
		WithMaxOpenFiles(-1)(cfg)
		assert.Equal(t, -1, cfg.maxOpenFiles)

		// 测试无效值（应该不改变）
		WithMaxOpenFiles(0)(cfg)
		assert.Equal(t, -1, cfg.maxOpenFiles)
	})

	t.Run("WithIgnoreDeadLog", func(t *testing.T) {
		cfg := defaultConfig()
		duration := 5 * time.Minute
		WithIgnoreDeadLog(duration)(cfg)
		assert.Equal(t, duration, cfg.ignoreDeadLog)

		// 测试零值（应该不改变）
		WithIgnoreDeadLog(0)(cfg)
		assert.Equal(t, duration, cfg.ignoreDeadLog)
	})

	t.Run("WithFileSizeThreshold", func(t *testing.T) {
		cfg := defaultConfig()
		threshold := int64(1024 * 1024) // 1MB
		WithFileSizeThreshold(threshold)(cfg)
		assert.Equal(t, threshold, cfg.fileSizeThreshold)

		// 测试零值（应该不改变）
		WithFileSizeThreshold(0)(cfg)
		assert.Equal(t, threshold, cfg.fileSizeThreshold)
	})

	t.Run("WithExtraTags", func(t *testing.T) {
		cfg := defaultConfig()
		tags := map[string]string{
			"env":     "production",
			"version": "1.0.0",
		}
		WithExtraTags(tags)(cfg)

		// 应该保留 service 标签
		assert.Equal(t, "default", cfg.extraTags["service"])
		assert.Equal(t, "production", cfg.extraTags["env"])
		assert.Equal(t, "1.0.0", cfg.extraTags["version"])
	})

	t.Run("AddTag", func(t *testing.T) {
		cfg := defaultConfig()
		AddTag("custom", "value")(cfg)
		assert.Equal(t, "value", cfg.extraTags["custom"])
		assert.Equal(t, "default", cfg.extraTags["service"])
	})

	t.Run("WithInsideFilepathFunc", func(t *testing.T) {
		cfg := defaultConfig()
		fn := func(path string) string {
			return "processed-" + path
		}
		WithInsideFilepathFunc(fn)(cfg)
		assert.NotNil(t, cfg.insideFilepathFunc)
		assert.Equal(t, "processed-test", cfg.insideFilepathFunc("test"))
	})

	t.Run("WithForwardFunc", func(t *testing.T) {
		cfg := defaultConfig()
		fn := func(filename, text string, fields map[string]interface{}) error {
			return nil
		}
		WithForwardFunc(fn)(cfg)
		assert.NotNil(t, cfg.forwardFunc)
	})

	t.Run("WithFeeder", func(t *testing.T) {
		cfg := defaultConfig()
		feeder := dkio.DefaultFeeder()
		WithFeeder(feeder)(cfg)
		assert.Equal(t, feeder, cfg.feeder)
	})
}

// TestMode 测试 Mode 类型
func TestMode(t *testing.T) {
	t.Run("String-methods", func(t *testing.T) {
		assert.Equal(t, "file", FileMode.String())
		assert.Equal(t, "docker-json", DockerJSONLogMode.String())
		assert.Equal(t, "cri-log", CriLogdMode.String())
		assert.Equal(t, "unknown", Mode(99).String())
	})
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	assert.Equal(t, "default", cfg.source)
	assert.Equal(t, "default", cfg.extraTags["service"])
	assert.Equal(t, int64(1000*1000*20), cfg.fileSizeThreshold) // 20MB
	assert.Equal(t, 500, cfg.maxOpenFiles)
	assert.NotNil(t, cfg.feeder)
}

// TestBuildConfigWithNilOptions 测试空选项
func TestBuildConfigWithNilOptions(t *testing.T) {
	cfg := buildConfig(nil)
	expected := defaultConfig()
	assert.Equal(t, expected, cfg)
}

// TestBuildConfigWithNilOption 测试包含 nil 的选项
func TestBuildConfigWithNilOption(t *testing.T) {
	opts := []Option{
		WithSource("test"),
		nil, // nil 选项应该被忽略
		WithService("test-service"),
	}

	cfg := buildConfig(opts)
	assert.Equal(t, "test", cfg.source)
	assert.Equal(t, "test-service", cfg.service)
}

// TestDeprecatedOptions 测试已废弃的选项
func TestDeprecatedOptions(t *testing.T) {
	t.Run("WithIgnoredStatuses", func(t *testing.T) {
		cfg := defaultConfig()
		statuses := []string{"debug", "info"}
		WithIgnoredStatuses(statuses)(cfg)
		assert.Equal(t, statuses, cfg.ignoredStatuses)
	})

	t.Run("WithDisableStatusField", func(t *testing.T) {
		cfg := defaultConfig()
		WithDisableStatusField(true)(cfg)
		assert.True(t, cfg.disableStatusField)
	})
}

// TestComplexScenarios 测试复杂场景
func TestComplexScenarios(t *testing.T) {
	t.Run("multiple-extra-tags", func(t *testing.T) {
		cfg := defaultConfig()

		// 添加多个标签
		AddTag("env", "production")(cfg)
		AddTag("version", "1.0.0")(cfg)
		AddTag("region", "us-west")(cfg)

		assert.Equal(t, "production", cfg.extraTags["env"])
		assert.Equal(t, "1.0.0", cfg.extraTags["version"])
		assert.Equal(t, "us-west", cfg.extraTags["region"])
		assert.Equal(t, "default", cfg.extraTags["service"])
	})

	t.Run("override-extra-tags", func(t *testing.T) {
		cfg := defaultConfig()

		// 先添加一些标签
		AddTag("env", "development")(cfg)
		AddTag("version", "0.1.0")(cfg)

		// 然后用 WithExtraTags 覆盖
		newTags := map[string]string{
			"env":     "production",
			"version": "2.0.0",
			"new":     "tag",
		}
		WithExtraTags(newTags)(cfg)

		// 应该保留 service 标签
		assert.Equal(t, "default", cfg.extraTags["service"])
		assert.Equal(t, "production", cfg.extraTags["env"])
		assert.Equal(t, "2.0.0", cfg.extraTags["version"])
		assert.Equal(t, "tag", cfg.extraTags["new"])
	})

	t.Run("service-auto-assignment", func(t *testing.T) {
		cfg := defaultConfig()

		// 设置 source 后，service 应该自动设置为相同的值
		WithSource("my-app")(cfg)
		assert.Equal(t, "my-app", cfg.source)
		assert.Equal(t, "my-app", cfg.service)
		assert.Equal(t, "my-app", cfg.extraTags["service"])
	})

	t.Run("empty-source-service-fallback", func(t *testing.T) {
		cfg := defaultConfig()

		// 先设置 service
		WithService("my-service")(cfg)
		assert.Equal(t, "my-service", cfg.service)
		assert.Equal(t, "my-service", cfg.extraTags["service"])

		// 然后设置空的 source，source 不会改变，service 也不会改变
		WithSource("")(cfg)
		assert.Equal(t, "default", cfg.source)     // 保持默认值
		assert.Equal(t, "my-service", cfg.service) // service 不会改变
		assert.Equal(t, "my-service", cfg.extraTags["service"])
	})
}
