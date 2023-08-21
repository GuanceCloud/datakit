// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
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

	c.Dataway.Hostname = c.Hostname

	l.Infof("setup dataway with global host tags %q, election tags %q",
		c.GlobalHostTags, c.Election.Tags)

	if err := c.Dataway.Init(
		dataway.WithGlobalTags(c.GlobalHostTags, c.Election.Tags),
	); err != nil {
		return err
	}

	return nil
}
