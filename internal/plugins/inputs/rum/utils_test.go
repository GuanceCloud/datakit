// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"net"
	"testing"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/stretchr/testify/assert"
)

func TestIsDomainName(t *testing.T) {
	assert.False(t, isDomainName("127.0.0.1"))
	assert.False(t, isDomainName("172.16.1.10"))
	assert.False(t, isDomainName("114.114.114.114"))
	assert.False(t, isDomainName("[::1]"))
	assert.True(t, isDomainName("localhost"))
	assert.True(t, isDomainName("xxxx.xxxxxxxxxxxxxxxxxxx"))
	assert.True(t, isDomainName("cdn.cnbj1.fds.api.mi-img.com"))
}

func TestLruCDNCache(t *testing.T) {
	cache := expirable.NewLRU[string, *cdnResolved](8, nil, cdnCacheTTL)

	domains := []string{
		"cdn.cnbj1.fds.api.mi-img.com",
		"res.vmallres.com",
		"f7.baidu.com",
		"shopstatic.vivo.com.cn",
		"msecfs.opposhop.cn",
		"img11.360buyimg.com",
		"m.ykimg.com",
		"userblink.csdnimg.cn",
		"resource.ksyun.com",
	}

	for _, domain := range domains {
		cname, cdnName, err := lookupCDNName(domain)
		if err != nil {
			t.Log(err)
			continue
		}
		t.Logf("cdn name for domain %s: cname: %s, cdn: %s", domain, cname, cdnName)
	}

	a := newCDNResolved("baidu.com", "", "百度云")
	b := newCDNResolved("qiniu.com", "", "七牛云")
	c := newCDNResolved("aliyun.com", "", "阿里云")
	d := newCDNResolved("cloud.tencent.com", "", "腾讯云")
	e := newCDNResolved("kingsoft.com", "", "金山云")
	f := newCDNResolved("ucloud.com", "", "优克得")
	g := newCDNResolved("huawei.com", "", "华为云")
	h := newCDNResolved("wangsu.com", "", "网宿CDN")

	cache.Add(a.domain, a)
	cache.Add(b.domain, b)
	cache.Add(c.domain, c)
	cache.Add(d.domain, d)
	cache.Add(e.domain, e)
	cache.Add(f.domain, f)
	cache.Add(g.domain, g)
	cache.Add(h.domain, h)

	for _, resolved := range cache.Values() {
		if resolved != nil {
			t.Logf("%+#v", *resolved)
		}
	}

	i := newCDNResolved("cdn.cnbj1.fds.api.mi-img.com", "", "小米CDN")

	cache.Add(i.domain, i)
	for _, resolved := range cache.Values() {
		if resolved != nil {
			t.Logf("%+#v", *resolved)
		}
	}

	node, ok := cache.Get("qiniu.com")
	t.Logf("%+v\n", node)
	assert.True(t, ok, "")
}

func TestIsPrivateIP(t *testing.T) {
	assert.True(t, isPrivateIP(net.ParseIP("10.200.14.195")), "10.200.14.195 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("127.0.0.1")), "127.0.0.1 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("192.168.100.1")), "192.168.100.1 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("172.16.2.14")), "172.16.2.14 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("172.17.2.14")), "172.17.2.14 is a private ip")
	assert.True(t, !isPrivateIP(net.ParseIP("8.8.8.8")), "8.8.8.8 is not a private ip")
}
