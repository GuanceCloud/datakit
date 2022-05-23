package script

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestScript(t *testing.T) {
	s, err := NewScript("abc.p", "if true {}", GitRepoScriptNS, datakit.Logging)
	if err != nil {
		t.Fatal(err)
	}
	if ng := s.Engine(); ng == nil {
		t.Fatalf("no engine")
	}
	outp, drop, err := s.Run("ng", nil, nil, "msg", time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, outp.Fields, map[string]interface{}{"status": DefaultPipelineStatus})
	assert.Equal(t, outp.Tags, map[string]string{})
	assert.Equal(t, "abc.p", s.Name())
	assert.Equal(t, datakit.Logging, s.Category())
	assert.Equal(t, s.NS(), GitRepoScriptNS)

	t.Log(drop)
	t.Log(outp)

	outp, _, err = s.Run("ng", nil, nil, "msg", time.Now(), &Option{
		DisableAddStatusField: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(outp.Fields) != 0 {
		t.Fatal(outp.Fields)
	}

	_, drop, err = s.Run("ng", nil, nil, "msg", time.Now(), &Option{
		DisableAddStatusField: false,
		IgnoreStatus:          []string{DefaultPipelineStatus},
	})
	if err != nil {
		t.Fatal(err)
	}
	if drop != true {
		t.Fatal("!drop")
	}
}
