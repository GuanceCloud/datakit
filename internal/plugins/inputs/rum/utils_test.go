// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"archive/zip"
	"net"
	"os"
	"path/filepath"
	T "testing"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDomainName(t *T.T) {
	assert.False(t, isDomainName("127.0.0.1"))
	assert.False(t, isDomainName("172.16.1.10"))
	assert.False(t, isDomainName("114.114.114.114"))
	assert.False(t, isDomainName("[::1]"))
	assert.True(t, isDomainName("localhost"))
	assert.True(t, isDomainName("xxxx.xxxxxxxxxxxxxxxxxxx"))
	assert.True(t, isDomainName("cdn.cnbj1.fds.api.mi-img.com"))
}

func TestLruCDNCache(t *T.T) {
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

func TestIsPrivateIP(t *T.T) {
	assert.True(t, isPrivateIP(net.ParseIP("10.200.14.195")), "10.200.14.195 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("127.0.0.1")), "127.0.0.1 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("192.168.100.1")), "192.168.100.1 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("172.16.2.14")), "172.16.2.14 is a private ip")
	assert.True(t, isPrivateIP(net.ParseIP("172.17.2.14")), "172.17.2.14 is a private ip")
	assert.True(t, !isPrivateIP(net.ParseIP("8.8.8.8")), "8.8.8.8 is not a private ip")
}

func Test_loadZipFile(t *T.T) {
	t.Run(`ignore`, func(t *T.T) {
		zipEntryName := "../../some-source.map"
		zipFileName := filepath.Join(t.TempDir(), "some.zip")

		newZipFile, err := os.Create(zipFileName)
		require.NoError(t, err)
		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)

		content := "some sourcemap binary data"
		fileHeader := &zip.FileHeader{
			Name: zipEntryName,
		}
		fileHeader.Method = zip.Store // no compression

		zipEntry, err := zipWriter.CreateHeader(fileHeader)
		require.NoError(t, err)

		_, err = zipEntry.Write([]byte(content))
		require.NoError(t, err)

		zipWriter.Close() // Close is required, or zip.OpenReader() error: zip: not a valid zip file

		res, err := loadZipFile(zipFileName, 1024, ".map")
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run(`basic`, func(t *T.T) {
		zipEntryName := "some-source.map"
		zipFileName := filepath.Join(t.TempDir(), "some.zip")

		newZipFile, err := os.Create(zipFileName)
		require.NoError(t, err)
		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)

		fileHeader := &zip.FileHeader{
			Name: zipEntryName,
		}
		fileHeader.Method = zip.Store // no compression

		zipEntry, err := zipWriter.CreateHeader(fileHeader)
		require.NoError(t, err)

		content, err := os.ReadFile("testdata/mapfile.json")
		require.NoError(t, err)

		_, err = zipEntry.Write(content)
		require.NoError(t, err)

		zipWriter.Close() // Close is required, or zip.OpenReader() error: zip: not a valid zip file

		res, err := loadZipFile(zipFileName, 1<<20, ".map")
		require.NoError(t, err)
		assert.Len(t, res, 1)

		t.Logf("source map: %+#v", res[zipEntryName])
	})

	t.Run(`mix`, func(t *T.T) {
		entries := []string{
			"some-source.map",
			"../../abc/../some-source-1.map",
			"../../some-source-2.map",
			"../..//some-source-3.map",
			"/etc/crontab/some-source-4.map",
			"abc/../some-source-5.map", // it's ok: filename is some-source-5.map
		}

		zipFileName := filepath.Join(t.TempDir(), "some.zip")

		newZipFile, err := os.Create(zipFileName)
		require.NoError(t, err)
		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)
		content, err := os.ReadFile("testdata/mapfile.json")
		require.NoError(t, err)

		for _, ent := range entries {
			fileHeader := &zip.FileHeader{
				Name: ent,
			}
			fileHeader.Method = zip.Store // no compression

			zipEntry, err := zipWriter.CreateHeader(fileHeader)
			require.NoError(t, err)

			_, err = zipEntry.Write(content)
			require.NoError(t, err)
		}

		zipWriter.Close() // Close is required, or zip.OpenReader() error: zip: not a valid zip file

		res, err := loadZipFile(zipFileName, 1<<20, ".map")
		require.NoError(t, err)
		assert.Len(t, res, 2)

		for k := range res {
			t.Logf("source map: %+#v", res[k])
		}
	})
}
