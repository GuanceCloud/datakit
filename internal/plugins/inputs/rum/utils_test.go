// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"net"
	"testing"
	"time"

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

func TestQueue(t *testing.T) {
	q := NewQueue[string]()

	a, b, c, d := NewQueueNode("aaaa"), NewQueueNode("bbbb"), NewQueueNode("cccc"), NewQueueNode("dddd")

	q.Enqueue(a)
	q.Enqueue(b)
	q.Enqueue(c)
	q.Enqueue(d)
	q.dump() // d c b a

	assert.Equal(t, 4, q.Size())

	q.MoveToFront(q.RearNode()) // a d c b
	q.MoveToFront(q.RearNode()) // b a d c
	q.MoveToFront(q.RearNode()) // c b a d
	q.dump()                    // c b a d

	assert.Equal(t, q.RearNode(), d)
	assert.Equal(t, 4, q.Size())

	q.Remove(a) // c b d
	q.dump()    // c b d

	assert.Equal(t, 3, q.Size())
	assert.Equal(t, d, q.Dequeue()) // c b
	assert.Equal(t, b, q.Dequeue()) // c
	assert.Equal(t, c, q.Dequeue()) // empty
	q.Dequeue()                     // do nothing
	q.Dequeue()                     // do nothing
	q.dump()                        // empty

	assert.True(t, q.Empty(), "logic error, queue expected to be empty")

	e, f, g := NewQueueNode("eeee"), NewQueueNode("ffff"), NewQueueNode("gggg")
	q.Enqueue(e)
	assert.Equal(t, 1, q.Size())
	assert.Equal(t, e, q.FrontNode())
	assert.Equal(t, e, q.RearNode())

	q.Enqueue(f)
	q.Enqueue(g)
	q.dump()
	assert.Equal(t, 3, q.Size())
	assert.Equal(t, g, q.FrontNode())
	assert.Equal(t, e, q.RearNode())
}

func TestLruCDNCache(t *testing.T) {
	cache := newLruCDNCache(8)

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

	a := newCDNResolved("baidu.com", "", "百度云", time.Now())
	b := newCDNResolved("qiniu.com", "", "七牛云", time.Now())
	c := newCDNResolved("aliyun.com", "", "阿里云", time.Now())
	d := newCDNResolved("cloud.tencent.com", "", "腾讯云", time.Now())
	e := newCDNResolved("kingsoft.com", "", "金山云", time.Now())
	f := newCDNResolved("ucloud.com", "", "优克得", time.Now())
	g := newCDNResolved("huawei.com", "", "华为云", time.Now())
	h := newCDNResolved("wangsu.com", "", "网宿CDN", time.Now())

	cache.push(a)
	cache.push(b)
	cache.push(c)
	cache.push(d)
	cache.push(e)
	cache.push(f)
	cache.push(g)
	cache.push(h)
	cache.queue.dump()

	i := newCDNResolved("cdn.cnbj1.fds.api.mi-img.com", "", "小米CDN", time.Now())

	cache.push(i)
	cache.queue.dump()

	node := cache.get("qiniu.com")
	t.Logf("%+v\n", node.Data)
	assert.True(t, node != nil, "")
}

func TestIsPrivateIP(t *testing.T) {
	assert.True(t, isPrivateIP(net.ParseIP("10.200.14.195")), "10.200.14.195 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("127.0.0.1")), "127.0.0.1 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("192.168.100.1")), "192.168.100.1 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("172.16.2.14")), "172.16.2.14 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("172.17.2.14")), "172.17.2.14 is a private ip")
	assert.True(t, !isPrivateIP(net.ParseIP("8.8.8.8")), "8.8.8.8 is not a private ip")
}
