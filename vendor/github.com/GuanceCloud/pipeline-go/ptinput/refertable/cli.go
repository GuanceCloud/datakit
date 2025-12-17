// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package refertable

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

var DefaultRetryConfig = RetryConfig{
	MaxRetry:       3,
	InitialDelay:   time.Second,
	MaxDelay:       time.Second * 15,
	RetryCondition: nil,
}

type RetryConfig struct {
	MaxRetry       int                                       // 最大重试次数
	InitialDelay   time.Duration                             // 初始重试间隔
	MaxDelay       time.Duration                             // 最大重试间隔
	RetryCondition func(err error, resp *http.Response) bool // 自定义重试条件
}

func DoRequestWithRetry(
	ctx context.Context,
	method string,
	url string,
	body []byte,
	headers map[string]string,
	retryConfig *RetryConfig,
	transport *http.Transport,
) (*http.Response, error) {
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Minute,
	}

	if retryConfig.RetryCondition == nil {
		retryConfig.RetryCondition = func(err error, resp *http.Response) bool {
			if err != nil {
				return true
			}
			// HTTP状态码：5xx（服务器错误）、429（限流）、408（请求超时）
			return resp.StatusCode >= 500 || resp.StatusCode == 429 || resp.StatusCode == 408
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 5. 初始化指数退避延迟
	delay := retryConfig.InitialDelay

	// 6. 重试循环
	for attempt := 0; attempt <= retryConfig.MaxRetry; attempt++ {
		// 检查上下文是否取消/超时
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request canceled or timed out: %w", ctx.Err())
		}

		// 发起请求（每次重试克隆请求，避免body被消费）
		cloneReq := req.Clone(ctx)
		resp, err := client.Do(cloneReq)

		// 请求成功且无需重试，直接返回
		if err == nil && !retryConfig.RetryCondition(err, resp) {
			return resp, nil
		}

		// 关闭响应体（避免内存/连接泄漏）
		if resp != nil {
			_ = resp.Body.Close()
		}

		// 最后一次重试失败，返回最终错误
		if attempt == retryConfig.MaxRetry {
			return nil, fmt.Errorf("the last retry failed: %w", err)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}

		// 指数退避：延迟翻倍，不超过最大延迟
		delay *= 2
		if delay > retryConfig.MaxDelay {
			delay = retryConfig.MaxDelay
		}
	}

	return nil, fmt.Errorf("failed to make request after %d attempts", retryConfig.MaxRetry)
}
