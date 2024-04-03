package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	v1 := "k8s_url"
	t.Setenv(EnvPrefix+EnvK8sURL, v1)
	v2 := "k8s_token"
	t.Setenv(EnvPrefix+EnvBearerToken, v2)
	v3 := "k8s_token_path"
	t.Setenv(EnvPrefix+EnvBearerTokenPath, v3)
	v4 := "netfilter"
	t.Setenv(EnvPrefix+EnvNetlogNetFilter, v4)

	fl := Flag{}
	readEnv(&fl)

	assert.Equal(t, v1, fl.K8sInfo.URL)
	assert.Equal(t, v2, fl.K8sInfo.BearerToken)
	assert.Equal(t, v3, fl.K8sInfo.BearerTokenPath)
	assert.Equal(t, v4, fl.BPFNetLog.NetFilter)

	t.Log(fl)
}
