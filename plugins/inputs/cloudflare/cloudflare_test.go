package cloudflare

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {

	var c = Cloudflare{
		Email:    os.Getenv("CF_EMAIL"),
		ZoneID:   os.Getenv("CF_ZoneID"),
		APIKey:   os.Getenv("CF_APIKEY"),
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
