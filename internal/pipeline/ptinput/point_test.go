// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ptinput

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/platypus/pkg/ast"
)

func TestPt(t *testing.T) {
	pt := NewPlPoint(point.Logging, "t", nil, nil, time.Now())
	if _, _, err := pt.Get("a"); err == nil {
		t.Fatal("err == nil")
	}

	if _, _, ok := pt.GetWithIsTag("a"); ok {
		t.Fatal("ok")
	}

	if err := pt.Set("a", 1, ast.Int); err != nil {
		t.Fatal(err)
	}

	if err := pt.Set("a1", []any{1}, ast.List); err != nil {
		t.Fatal(err)
	}

	if err := pt.Set("xx2", []any{1}, ast.List); err != nil {
		t.Fatal(err)
	}

	if err := pt.Set("xx2", 1.2, ast.Float); err != nil {
		t.Fatal(err)
	}

	if _, _, err := pt.Get("xx2"); err != nil {
		t.Fatal(err)
	}

	if err := pt.RenameKey("xx2", "xxb"); err != nil {
		t.Fatal(err)
	}

	if err := pt.SetTag("a", 1., ast.Float); err != nil {
		t.Fatal(err)
	}

	if err := pt.Set("a", 1, ast.Int); err != nil {
		t.Fatal(err)
	}

	if _, ok := pt.Fields()["a"]; ok {
		t.Fatal("a in fields")
	}

	if err := pt.RenameKey("a", "b"); err != nil {
		t.Fatal(err)
	}

	if pt.PtTime().UnixNano() == 0 {
		t.Fatal("time == 0")
	}

	pt.GetAggBuckets()
	pt.SetAggBuckets(nil)

	pt.Set("time", 1, ast.Int)
	pt.KeyTime2Time()
	if pt.PtTime().UnixNano() != 1 {
		t.Fatal("time != 1")
	}

	pt.MarkDrop(true)
	if !pt.Dropped() {
		t.Fatal("!dropped")
	}

	dpt, err := pt.DkPoint()
	if err != nil {
		t.Fatal(err)
	}

	pt, err = WrapDeprecatedPoint(point.Logging, dpt)
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := pt.Get("b"); err != nil {
		t.Fatal(err.Error())
	}

	if _, ok := pt.Tags()["b"]; !ok {
		t.Fatal("b not in tags")
	}

	if _, istag, ok := pt.GetWithIsTag("b"); !ok || !istag {
		t.Fatal("not tag")
	}

	if err := pt.Set("b", []any{}, ast.List); err != nil {
		t.Fatal(err)
	}

	if _, istag, ok := pt.GetWithIsTag("a1"); !ok || istag {
		t.Fatal("is tag")
	}

	if _, ok := pt.Fields()["xxb"]; !ok {
		t.Fatal("xxb not in field")
	}

	if pt.GetPtName() != "t" {
		t.Fatal("name != \"t\"")
	}

	pt.SetPtName("t2")
	if pt.GetPtName() != "t2" {
		t.Fatal("name != \"t2\"")
	}
}
