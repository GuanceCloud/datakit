// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
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
	plpt := &ptinput.Point{}
	plpt = ptinput.InitPt(plpt, "ng", nil, nil, time.Now())
	err := s.Run(plpt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, plpt.Fields, map[string]interface{}{"status": DefaultStatus})
	assert.Equal(t, plpt.Tags, map[string]string{})
	assert.Equal(t, "abc.p", s.Name())
	assert.Equal(t, datakit.Logging, s.Category())
	assert.Equal(t, s.NS(), GitRepoScriptNS)

	//nolint:dogsled
	plpt = ptinput.InitPt(plpt, "ng", nil, nil, time.Now())
	err = s.Run(plpt, nil, &Option{DisableAddStatusField: true})
	if err != nil {
		t.Fatal(err)
	}

	if len(plpt.Fields) != 0 {
		t.Fatal(plpt.Fields)
	}

	//nolint:dogsled
	plpt = ptinput.InitPt(plpt, "ng", nil, nil, time.Now())
	err = s.Run(plpt, nil, &Option{
		DisableAddStatusField: false,
		IgnoreStatus:          []string{DefaultStatus},
	})
	if err != nil {
		t.Fatal(err)
	}
	if plpt.Drop != true {
		t.Fatal("!drop")
	}
}

func TestNewScript(t *testing.T) {
	for category := range datakit.CategoryDirName() {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if err := ret["abc"].Run(ptinput.InitPt(&ptinput.Point{}, "d", nil, nil, time.Time{}), nil, nil); err != nil {
			t.Error(err)
		}
	}
	for category := range _allCategory {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if err := ret["abc"].Run(ptinput.InitPt(&ptinput.Point{}, "d", nil, nil, time.Time{}), nil, nil); err != nil {
			t.Error(err)
		}
	}
	for category := range _allDeprecatedCategory {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if err := ret["abc"].Run(ptinput.InitPt(&ptinput.Point{}, "d", nil, nil, time.Time{}), nil, nil); err != nil {
			t.Error(err)
		}
	}
	m1, m2 := datakit.CategoryList()
	for category := range m1 {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if err := ret["abc"].Run(ptinput.InitPt(&ptinput.Point{}, "d", nil, nil, time.Time{}), nil, nil); err != nil {
			t.Error(err)
		}
	}
	for category := range m2 {
		if ret, retErr := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, category); len(retErr) > 0 {
			t.Error(retErr)
		} else if err := ret["abc"].Run(ptinput.InitPt(&ptinput.Point{}, "d", nil, nil, time.Time{}), nil, nil); err != nil {
			t.Error(err)
		}
	}

	if _, err := NewScripts(map[string]string{"abc": "if true{}"}, nil, DefaultScriptNS, "-!-c-a-t-e-g-0-r-Y"); err == nil {
		t.Error("error == nil")
	}
}

func TestDrop(t *testing.T) {
	ret, retErr := NewScripts(map[string]string{"abc.p": "add_key(a, \"a\"); add_key(status, \"debug\"); drop(); add_key(b, \"b\")"},
		nil, GitRepoScriptNS, datakit.Logging)
	if len(retErr) > 0 {
		t.Fatal(retErr)
	}

	s := ret["abc.p"]
	t.Log(s.FilePath())

	plpt := &ptinput.Point{}
	plpt = ptinput.InitPt(plpt, "ng", nil, nil, time.Now())
	if err := s.Run(plpt, nil, nil); err != nil {
		t.Fatal(err)
	}

	if plpt.Drop != true {
		t.Error("drop != true")
	}
}
