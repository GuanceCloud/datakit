// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func (c *Config) SetupDataway() error {
	if c.Dataway == nil {
		return fmt.Errorf("dataway config is empty")
	}

	// 如果 env 已传入了 dataway 配置, 则不再追加老的 dataway 配置,
	// 避免俩边配置了同样的 dataway, 造成数据混乱
	if c.Dataway.DeprecatedURL != "" && len(c.Dataway.URLs) == 0 {
		c.Dataway.URLs = []string{c.Dataway.DeprecatedURL}
	}

	if len(c.Dataway.URLs) > 0 && c.Dataway.URLs[0] == datakit.DatawayDisableURL {
		c.RunMode = datakit.ModeDev
		return nil
	} else {
		c.RunMode = datakit.ModeNormal
	}

	dataway.ExtraHeaders = map[string]string{
		"X-Datakit-Info": fmt.Sprintf("%s; %s", c.Hostname, datakit.Version),
	}

	c.Dataway.Hostname = c.Hostname

	// NOTE: this should not happen, the installer will rewrite datakit.conf
	// to move top-level sinker config to dataway.
	if c.SinkersDeprecated != nil && len(c.SinkersDeprecated.Arr) > 0 {
		c.Dataway.Sinkers = c.SinkersDeprecated.Arr
	}

	if err := c.Dataway.Init(); err != nil {
		return err
	}

	return nil
}
