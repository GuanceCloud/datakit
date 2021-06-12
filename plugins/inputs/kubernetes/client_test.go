package kubernetes

import (
	"fmt"
	// "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
	"time"
)

type pki struct {
	path string
}

func NewPKI(path string) *pki {
	return &pki{path: path}
}

func (p *pki) CACertPath() string {
	return path.Join(p.path, "cacert.pem")
}

func (p *pki) ClientCertPath() string {
	return path.Join(p.path, "clientcert.pem")
}

func (p *pki) ClientKeyPath() string {
	return path.Join(p.path, "clientkey.pem")
}

func TestCreateConfigByKubePath(t *testing.T) {
	testsCase := []struct {
		input  string
		expect string
		isErr  bool
	}{
		{
			input:  "/Users/liushaobo/.kube/config",
			expect: "https://172.16.2.41:6443",
		},
		{
			input:  "/Users/liushaobo/.kube",
			expect: "",
			isErr:  true,
		},
	}

	for i, test := range testsCase {
		errmsg := fmt.Sprintf("case%d", i)
		c, err := createConfigByKubePath(test.input)
		if test.isErr {
			assert.Error(t, err, errmsg)
			continue
		}
		got := c.Host
		assert.NoError(t, err, errmsg)
		assert.Equal(t, test.expect, got, errmsg)

		_, err = newClient(c, 5*time.Second)
		if test.isErr {
			assert.Error(t, err, errmsg)
			continue
		}
		assert.NoError(t, err, errmsg)
	}
}

// func TestCreateConfigByCert(t *testing.T) {
// 	var pki = NewPKI("./pki")

// 	testsCase := []struct {
// 		input     string
// 		tlsConfig *tls.ClientConfig
// 		expect    string
// 		isErr     bool
// 	}{
// 		{
// 			input: "https://172.16.2.41:6443",
// 			tlsConfig: &tls.ClientConfig{
// 				TLSCA:   pki.CACertPath(),
// 				TLSCert: pki.ClientCertPath(),
// 				TLSKey:  pki.ClientKeyPath(),
// 			},
// 			expect: "https://172.16.2.41:6443",
// 		},
// 		{
// 			input:     "https://172.16.2.41:6443",
// 			tlsConfig: &tls.ClientConfig{},
// 			expect:    "",
// 			isErr:     true,
// 		},
// 	}

// 	for i, test := range testsCase {
// 		errmsg := fmt.Sprintf("case%d", i)
// 		c := createConfigByCert(test.input, test.tlsConfig)
// 		got := c.Host
// 		assert.Equal(t, test.expect, got, errmsg)

// 		_, err := newClient(c, 5*time.Second)
// 		if test.isErr {
// 			assert.Error(t, err, errmsg)
// 			continue
// 		}
// 		assert.NoError(t, err, errmsg)
// 	}
// }

func TestCreateConfigByBearToken(t *testing.T) {
	var pki = NewPKI("./pki")

	testsCase := []struct {
		input       string
		bearerToken string
		caFile      string
		expect      string
		isErr       bool
	}{
		{
			input:       "https://172.16.2.41:6443",
			bearerToken: "eyJhbGciOiJSUzI1NiIsImtpZCI6InFWNzd1LTNDNEdEd0FlTjdPQzF1NXBGVnYxU2JrTlVJQ3RUUnZlbXRGZ1EifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4ta3F4NzUiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjM5ZmQxOTQ4LTY5YTAtNDZlZi1hZjc3LWYxYzUwMmFmZDdiMiIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.f4oPuQ0fuY1jZI7o7CeGr-FtfQbnYlxzphtZAeKo31HjQAG5ynl4rYLRt1PK7lpCoMiMrAw5xDSMlG2DN9bTF3OYQJbfC4Mq3olPGxHHjxoTSotrfGrMK779NZ_JzRw6OQ9mKEgG8vadFpd4nGRi4KuD-7w8ysOzm_j6Z78eVTxhKrOuU11a6WEUh_LGnJSNLjAdN8xKqim90qcWy5jvdYl2s9N2tRPvkSJ22xwJ9Icts0HHZfvAywG7Rb69WyN13ct37N1_bICwjVrWuONyXOgNSiV7JvUFI2ZFpKfpDrDhpGRwwmVCR5a8BjP0S1kNjjckK9ma4ubYyvLIDS86Xw",
			caFile:      pki.CACertPath(),
			expect:      "https://172.16.2.41:6443",
		},
	}

	for i, test := range testsCase {
		errmsg := fmt.Sprintf("case%d", i)
		c, err := createConfigByToken(test.input, test.bearerToken, test.caFile, false)
		if test.isErr {
			assert.Error(t, err, errmsg)
			continue
		}

		assert.NoError(t, err, errmsg)
		cli, err := newClient(c, 5*time.Second)
		if test.isErr {
			assert.Error(t, err, errmsg)
			continue
		}
		assert.NoError(t, err, errmsg)

		// 通过 ServerVersion 方法来获取版本号
		versionInfo, err := cli.ServerVersion()
		if err != nil {
			assert.Error(t, err, errmsg)
			continue
		}

		t.Log("version ==>", versionInfo.String())
	}
}
