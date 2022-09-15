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
	ret, retErr := NewScripts(map[string]string{"abc.p": "if true {}"}, nil, GitRepoScriptNS, datakit.Logging)

	if len(retErr) > 0 {
		t.Fatal(retErr)
	}

	s := ret["abc.p"]
	t.Log(s.FilePath())

	if ng := s.Engine(); ng == nil {
		t.Fatalf("no engine")
	}
	_, tags, f, _, drop, err := s.Run("ng", nil, nil, time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, f, map[string]interface{}{"status": DefaultStatus})
	assert.Equal(t, tags, map[string]string{})
	assert.Equal(t, "abc.p", s.Name())
	assert.Equal(t, datakit.Logging, s.Category())
	assert.Equal(t, s.NS(), GitRepoScriptNS)

	t.Log(drop)

	//nolint:dogsled
	_, _, f, _, _, err = s.Run("ng", nil, nil, time.Now(), &Option{
		DisableAddStatusField: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(f) != 0 {
		t.Fatal(f)
	}

	//nolint:dogsled
	_, _, _, _, drop, err = s.Run("ng", nil, nil, time.Now(), &Option{
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
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if _, _, _, _, _, err := ret["abc"].Run("d", nil, nil, time.Time{}, nil); err != nil {
			t.Error(err)
		}
	}
	for category := range _allCategory {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if _, _, _, _, _, err := ret["abc"].Run("d", nil, nil, time.Time{}, nil); err != nil {
			t.Error(err)
		}
	}
	for category := range _allDeprecatedCategory {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if _, _, _, _, _, err := ret["abc"].Run("d", nil, nil, time.Time{}, nil); err != nil {
			t.Error(err)
		}
	}
	m1, m2 := datakit.CategoryList()
	for category := range m1 {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if _, _, _, _, _, err := ret["abc"].Run("d", nil, nil, time.Time{}, nil); err != nil {
			t.Error(err)
		}
	}
	for category := range m2 {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if _, _, _, _, _, err := ret["abc"].Run("d", nil, nil, time.Time{}, nil); err != nil {
			t.Error(err)
		}
	}

	if _, err := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, "-!-c-a-t-e-g-0-r-Y"); err == nil {
		t.Error("error == nil")
	}
}
