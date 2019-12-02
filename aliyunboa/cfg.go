package aliyunboa

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"

	"github.com/influxdata/toml"
)

const (
	aliyunboaConfigSample = `
[[boa]]
  ## Aliyun Region (required)
  ## See: https://www.alibabacloud.com/help/zh/doc-detail/40654.htm
  region_id = 'cn-hangzhou'
  
  ## Aliyun Credentials (required)
  access_key_id = ''
  access_key_secret = ''

  account_interval = "24h"
  bill_interval = "1h"
  order_interval = "1h"
`
)

var (
	Cfg AliyunBoaCfg
)

type (
	Boa struct {
		AccessKeyID     string   `toml:"access_key_id"`
		AccessKeySecret string   `toml:"access_key_secret"`
		RegionID        string   `toml:"region_id"`
		AccountInterval Duration `toml:"account_interval"`
		BiilInterval    Duration `toml:"bill_interval"`
		OrdertInterval  Duration `toml:"order_interval"`
	}

	AliyunBoaCfg struct {
		Boas []*Boa `toml:"boa"`
	}
)

func (c *AliyunBoaCfg) SampleConfig() string {
	return aliyunboaConfigSample
}

func (c *AliyunBoaCfg) FilePath(root string) string {
	d := filepath.Join(root, "aliyunboa")
	return filepath.Join(d, "aliyunboa.conf")
}

func (c *AliyunBoaCfg) ToTelegraf(f string) (string, error) {
	return "", nil
}

func (c *AliyunBoaCfg) Load(f string) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if err = toml.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

type (
	Duration struct {
		time.Duration
	}
)

func (d *Duration) UnmarshalTOML(b []byte) error {

	var err error
	b = bytes.Trim(b, `'`)

	d.Duration, err = time.ParseDuration(string(b))
	if err == nil {
		return nil
	}

	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		d.Duration, err = time.ParseDuration(uq)
		if err != nil {
			return err
		}
	}

	sI, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		d.Duration = time.Second * time.Duration(sI)
		return nil
	}

	sF, err := strconv.ParseFloat(string(b), 64)
	if err == nil {
		d.Duration = time.Second * time.Duration(sF)
		return nil
	}

	return nil
}
