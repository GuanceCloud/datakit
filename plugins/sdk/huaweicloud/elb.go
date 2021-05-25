package huaweicloud

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud/elb"
)

const (
	elbListUrlV1  = "/v1.0/%s/elbaas/loadbalancers" //经典型
	elbListUrlV2  = "/v2/%s/elb/loadbalancers"      //共享型_企业项目
	elbListUrlV20 = "/v2.0/lbaas/loadbalancers"     //共享型
)

func (c *HWClient) ElbV1List(opt map[string]string) (res *elb.ListLoadbalancersV1, err error) {
	url := fmt.Sprintf(elbListUrlV1, c.projectid)
	resp, err := c.Request("GET", url, opt, nil)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}

	//c.logger.Debugf("resp %s", string(resp))
	err = json.Unmarshal(resp, &res)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}

	return res, nil
}

func (c *HWClient) ElbV2List(opt map[string]string) (res *elb.ListLoadbalancersV2, err error) {
	url := fmt.Sprintf(elbListUrlV2, c.projectid)
	resp, err := c.Request("GET", url, opt, nil)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}

	//c.logger.Debugf("resp %s", string(resp))
	err = json.Unmarshal(resp, &res)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}

	return res, nil
}

func (c *HWClient) ElbV20List(opt map[string]string) (res *elb.ListLoadbalancersV20, err error) {
	url := fmt.Sprintf(elbListUrlV20)
	opt[`project_id`] = c.projectid
	resp, err := c.Request("GET", url, opt, nil)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}

	//c.logger.Debugf("resp %s", string(resp))
	err = json.Unmarshal(resp, &res)
	if err != nil {
		c.logger.Errorf("%s", err)
		return nil, err
	}

	return res, nil
}
