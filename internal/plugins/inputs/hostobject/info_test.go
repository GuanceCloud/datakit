// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNetInfo(t *testing.T) {
	ifs, err := interfaces()
	if err != nil {
		l.Errorf("fail to get interfaces, %s", err)
	}
	var infos []*NetInfo

	// netVIfaces := map[string]bool{}
	netVIfaces, _ := NetIgnoreIfaces()

	for _, it := range ifs {
		if _, ok := netVIfaces[it.Name]; ok {
			continue
		}
		i := &NetInfo{
			Index:        it.Index,
			MTU:          it.MTU,
			Name:         it.Name,
			HardwareAddr: it.HardwareAddr,
			Flags:        it.Flags,
		}
		for _, ad := range it.Addrs {
			ip, _, _ := net.ParseCIDR(ad.Addr)
			if ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				i.IP4 = ad.Addr
				i.IP4All = append(i.IP4All, ad.Addr)
			} else if ip.To16() != nil {
				i.IP6 = ad.Addr
				i.IP6All = append(i.IP6All, ad.Addr)
			}
		}
		infos = append(infos, i)
	}
	assert.NotEmpty(t, infos, "infos should not be empty")
}

func createTempFile(t *testing.T, content []byte) string {
	t.Helper()

	tempFile, err := ioutil.TempFile("", "testfile*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	return tempFile.Name()
}

func TestGetConfigFile(t *testing.T) {
	testCases := []struct {
		name    string
		content []byte
		isValid bool
	}{
		{"ValidTextFile", []byte("This is a text file."), true},
		{"LargeFile", make([]byte, 5*1024), false}, // 大于 4KB 的文件
		{"NonTextFile", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}, false}, // 非文本文件
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := createTempFile(t, tc.content)
			defer os.Remove(filePath)

			ipt := &Input{
				ConfigPath: []string{filePath},
			}

			files := ipt.getConfigFile()

			require.Equal(t, tc.isValid, len(files) == 1)

			if tc.isValid {
				if content, ok := files[filePath]; !ok || content != string(tc.content) {
					t.Errorf("failed to read content from %s", filePath)
				}
			}
		})
	}
}
