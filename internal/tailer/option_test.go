// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
