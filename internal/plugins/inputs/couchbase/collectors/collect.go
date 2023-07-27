// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package collectors used to collect metrics.
package collectors

import (
	"net/http"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/objects"
)

type MetricContext struct {
	ClusterName  string
	NodeHostname string
	BucketName   string
	Keyspace     string
	Source       string
	Target       string
}

type Option struct {
	TLSOpen        bool
	CacertFile     string
	CertFile       string
	KeyFile        string
	Scheme         string
	Host           string
	Port           int
	AdditionalPort int
	User           string
	Password       string
}

type Client struct {
	client  *http.Client
	Opt     *Option
	Pts     []*point.Point
	Ctx     *MetricContext
	Tags    map[string]string
	URLTags map[string]string

	config *objects.CollectorConfig
}

func (c *Client) GetPts() error {
	c.Pts = make([]*point.Point, 0)
	c.Ctx = &MetricContext{}

	if err := c.nodeCollect(); err != nil {
		return err
	}
	if err := c.bucketInfoCollect(); err != nil {
		return err
	}
	if err := c.taskCollect(); err != nil {
		return err
	}
	if err := c.queryCollect(); err != nil {
		return err
	}
	if err := c.indexCollect(); err != nil {
		return err
	}
	if err := c.searchCollect(); err != nil {
		return err
	}
	if err := c.cbasCollect(); err != nil {
		return err
	}
	if err := c.eventingCollect(); err != nil {
		return err
	}
	if err := c.perNodeBucketStatsCollect(); err != nil {
		return err
	}
	if err := c.bucketStatusCollect(); err != nil {
		return err
	}

	return nil
}

func (c *Client) SetClient(cli *http.Client) {
	c.client = cli
}

func (c *Client) request(url string) (*http.Response, error) {
	req, err := c.getReq(url)
	if err != nil {
		return nil, err
	}

	r, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) getReq(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)

	if c.Opt.User != "" {
		req.SetBasicAuth(c.Opt.User, c.Opt.Password)
	}

	return req, err
}
