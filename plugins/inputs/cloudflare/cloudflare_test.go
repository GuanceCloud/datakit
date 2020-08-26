package cloudflare

import (
	"testing"
)

const (
	email  = "a_ha2008@qq.com"
	zoneID = "0355d2a9ca65054d3d7f6013d8b237f4"
	apiKey = "d44e173bf1a3cd849c868a9662ae56d32470a"
)

func TestMain(t *testing.T) {

	var c = Cloudflare{
		Email:    email,
		ZoneID:   zoneID,
		APIKey:   apiKey,
		Interval: "24h",
	}

	c.laodCfg()

	data, err := c.getMetrics()
	if err != nil {
		panic(err)
	}

	t.Logf("len: %d\n", len(data))
	t.Logf("%s\n", data)

}
