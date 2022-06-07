// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
	assert.Equal(t, outp.Fields, map[string]interface{}{"status": DefaultStatus})
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
		IgnoreStatus:          []string{DefaultStatus},
	})
	if err != nil {
		t.Fatal(err)
	}
	if drop != true {
		t.Fatal("!drop")
	}
}

func TestNewScript(t *testing.T) {
	for category := range datakit.CategoryDirName() {
		if s, err := NewScript("abc", "if true{}", DefaultScriptNS, category); err != nil {
			l.Error(err)
		} else if _, _, err := s.Run("d", nil, nil, "m", time.Time{}, nil); err != nil {
			l.Error(err)
		}
	}
	for category := range _allCategory {
		if s, err := NewScript("abc", "if true{}", DefaultScriptNS, category); err != nil {
			l.Error(err)
		} else if _, _, err := s.Run("d", nil, nil, "m", time.Time{}, nil); err != nil {
			l.Error(err)
		}
	}
	for category := range _allDeprecatedCategory {
		if s, err := NewScript("abc", "if true{}", DefaultScriptNS, category); err != nil {
			l.Error(err)
		} else if _, _, err := s.Run("d", nil, nil, "m", time.Time{}, nil); err != nil {
			l.Error(err)
		}
	}
	m1, m2 := datakit.CategoryList()
	for category := range m1 {
		if s, err := NewScript("abc", "if true{}", DefaultScriptNS, category); err != nil {
			l.Error(err)
		} else if _, _, err := s.Run("d", nil, nil, "m", time.Time{}, nil); err != nil {
			l.Error(err)
		}
	}
	for category := range m2 {
		if s, err := NewScript("abc", "if true{}", DefaultScriptNS, category); err != nil {
			l.Error(err)
		} else if _, _, err := s.Run("d", nil, nil, "m", time.Time{}, nil); err != nil {
			l.Error(err)
		}
	}

	if _, err := NewScript("abc", "if true{}", DefaultScriptNS, "-!-c-a-t-e-g-0-r-Y"); err == nil {
		l.Error("error == nil")
	}
}
