// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogConfigs(t *testing.T) {
	cases := []struct {
		in                     string
		out                    logConfigs
		parseFail, contentFail bool
	}{
		{
			in: "[{\"disable\":true}]",
			out: logConfigs{
				&logConfig{
					Disable: true,
				},
			},
		},
		// fail
		{
			in:        "[{]",
			parseFail: true,
		},
		{
			in: "[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\"}]",
			out: logConfigs{
				&logConfig{
					Disable:  false,
					Source:   "testing-source",
					Service:  "testing-service",
					Pipeline: "test.p",
				},
			},
		},
		{
			in: "[{\"tags\":{\"some_tag\":\"some_value\"}}]",
			out: logConfigs{
				&logConfig{
					Tags: map[string]string{"some_tag": "some_value"},
				},
			},
		},
		{
			in: "[{\"multiline_match\":\"^\\\\[[0-9]{4}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline:         `^\[[0-9]{4}`,
					MultilinePatterns: []string{`^\[[0-9]{4}`},
				},
			},
		},
		{
			in: "[{\"multiline_match\":\"^\\\\[[0-9]{4}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline:         "^\\[[0-9]{4}",
					MultilinePatterns: []string{"^\\[[0-9]{4}"},
				},
			},
		},
		// fail，multiline_match 需要 4 条斜线转义等于 1 根
		{
			in:        "[{\"multiline_match\":\"^\\d{4}-\\d{2}\"}]",
			parseFail: true,
		},
		{
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline:         "^\\d{4}-\\d{2}",
					MultilinePatterns: []string{"^\\d{4}-\\d{2}"},
				},
			},
		},
		{
			// 等于上一条测试
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline:         `^\d{4}-\d{2}`,
					MultilinePatterns: []string{`^\d{4}-\d{2}`},
				},
			},
		},
		// 解析通过，内容错误
		{
			in: "[{\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline:         `^\\d{4}-\\d{2}`,
					MultilinePatterns: []string{`^\\d{4}-\\d{2}`},
				},
			},
			contentFail: true,
		},
		{
			// 8 条斜线等于实际 2 条
			in: "[{\"multiline_match\":\"^\\\\\\\\d{4}-\\\\\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline:         "^\\\\d{4}-\\\\d{2}",
					MultilinePatterns: []string{"^\\\\d{4}-\\\\d{2}"},
				},
			},
		},
		{
			// 等同上一条测试
			in: "[{\"multiline_match\":\"^\\\\\\\\d{4}-\\\\\\\\d{2}\"}]",
			out: logConfigs{
				&logConfig{
					Multiline:         `^\\d{4}-\\d{2}`,
					MultilinePatterns: []string{`^\\d{4}-\\d{2}`},
				},
			},
		},
		// 解析通过，内容错误
		{
			in: "[{\"tags\":{\"some_tag\":\"some_value\"}}]",
			out: logConfigs{
				&logConfig{
					Tags: map[string]string{"some_tag": "some_value_11"},
				},
			},
			contentFail: true,
		},
		{
			in: "[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\", \"tags\":{\"some_tag\":\"some_value\"}}]",
			out: logConfigs{
				&logConfig{
					Disable:           false,
					Source:            "testing-source",
					Service:           "testing-service",
					Pipeline:          "test.p",
					Multiline:         "^\\d{4}-\\d{2}",
					MultilinePatterns: []string{"^\\d{4}-\\d{2}"},
					Tags:              map[string]string{"some_tag": "some_value"},
				},
			},
		},
		// many config
		{
			in: "[{\"disable\":false}, {\"disable\":true}]",
			out: logConfigs{
				&logConfig{
					Disable: false,
				},
				&logConfig{
					Disable: true,
				},
			},
		},
	}

	for idx, tc := range cases {
		info := &logInstance{
			configStr: tc.in,
		}

		err := info.parseLogConfigs()
		if tc.parseFail && assert.Error(t, err) {
			t.Logf("[%d][OK   ] %s\n", idx, err)
			continue
		}
		if !assert.NoError(t, err) {
			t.Logf("[%d][ERROR] %s\n", idx, err)
			continue
		}

		res := info.configs

		if tc.contentFail {
			assert.Equal(t, tc.contentFail, assert.NotEqual(t, tc.out, res))
		} else {
			assert.Equal(t, !tc.contentFail, assert.Equal(t, tc.out, res))
		}

		t.Logf("[%d][OK   ] %v\n", idx, tc)
	}
}

func TestJoinHostFilepath(t *testing.T) {
	cases := []struct {
		inHostDir, inInsideDir, inPath string
		out                            string
	}{
		{
			inHostDir:   "/var/lib/kubelet/pods/ABCDEFG012344567/volumes/kubernetes.io~empty-dir/<volume-name>/",
			inInsideDir: "/tmp/log",
			inPath:      "/tmp/log/nginx-log/a.log",
			out:         "/var/lib/kubelet/pods/ABCDEFG012344567/volumes/kubernetes.io~empty-dir/<volume-name>/nginx-log/a.log",
		},
	}

	for _, tc := range cases {
		res := joinHostFilepath(tc.inHostDir, tc.inInsideDir, tc.inPath)
		assert.Equal(t, tc.out, res)
	}
}

func TestJoinInsideFilepath(t *testing.T) {
	cases := []struct {
		inHostDir, inInsideDir, inPath string
		out                            string
	}{
		{
			inHostDir:   "/var/lib/kubelet/pods/ABCDEFG012344567/volumes/kubernetes.io~empty-dir/<volume-name>/",
			inInsideDir: "/tmp/log",
			inPath:      "/var/lib/kubelet/pods/ABCDEFG012344567/volumes/kubernetes.io~empty-dir/<volume-name>/nginx-log/a.log",
			out:         "/tmp/log/nginx-log/a.log",
		},
	}

	for _, tc := range cases {
		res := joinInsideFilepath(tc.inHostDir, tc.inInsideDir, tc.inPath)
		assert.Equal(t, tc.out, res)
	}
}

func TestLogTableFindDifferences(t *testing.T) {
	testcases := []struct {
		inTable *logTable
		inIDs   []string
		out     []string
	}{
		{
			inTable: &logTable{
				table: map[string]map[string]func(){
					"id-01": nil,
					"id-02": nil,
				},
			},
			inIDs: []string{
				"id-01",
				"id-02",
			},
			out: nil,
		},
		{
			inTable: &logTable{
				table: map[string]map[string]func(){
					"id-01": nil,
					"id-02": nil,
				},
			},
			inIDs: []string{
				"id-01",
			},
			out: []string{
				"id-02",
			},
		},
		{
			inTable: &logTable{
				table: map[string]map[string]func(){
					"id-01": nil,
					"id-02": nil,
				},
			},
			inIDs: []string{
				"id-01",
				"id-03",
				"id-04",
			},
			out: []string{
				"id-02",
			},
		},
		{
			inTable: &logTable{
				table: map[string]map[string]func(){},
			},
			inIDs: []string{
				"id-01",
			},
			out: nil,
		},
		{
			inTable: &logTable{
				table: map[string]map[string]func(){},
			},
			inIDs: []string{},
			out:   nil,
		},
	}

	for _, tc := range testcases {
		res := tc.inTable.findDifferences(tc.inIDs)
		assert.Equal(t, tc.out, res)
	}
}

func TestLogTableString(t *testing.T) {
	t.Run("logtable-string", func(t *testing.T) {
		in := &logTable{
			table: map[string]map[string]func(){
				"id-01": {
					"/var/log/01/1": nil,
					"/var/log/01/2": nil,
				},
				"id-03": {
					"/var/log/03/1": nil,
					"/var/log/03/2": nil,
				},
				"id-02": {
					"/var/log/02/2": nil,
					"/var/log/02/1": nil,
				},
				"1234567890123": {
					"/var/log/01/1": nil,
				},
			},
		}

		out := "{id:123456789012,paths:[/var/log/01/1]}, {id:id-01,paths:[/var/log/01/1,/var/log/01/2]}, {id:id-02,paths:[/var/log/02/1,/var/log/02/2]}, {id:id-03,paths:[/var/log/03/1,/var/log/03/2]}"

		assert.Equal(t, out, in.String())
	})
}

func TestLogTableRemoveID(t *testing.T) {
	t.Run("logtable-remove-path", func(t *testing.T) {
		in := &logTable{
			table: map[string]map[string]func(){
				"id-01": {
					"/var/log/01/1": nil,
				},
				"id-02": {
					"/var/log/02/1": nil,
				},
			},
		}

		out := &logTable{
			table: map[string]map[string]func(){
				"id-02": {
					"/var/log/02/1": nil,
				},
			},
		}

		in.removeFromTable("id-01")
		assert.Equal(t, out, in)
	})
}

func TestLogTableRemovePath(t *testing.T) {
	t.Run("logtable-remove-path", func(t *testing.T) {
		in := &logTable{
			table: map[string]map[string]func(){
				"id-01": {
					"/var/log/01/1": nil,
					"/var/log/01/2": nil,
				},
			},
		}

		out := &logTable{
			table: map[string]map[string]func(){
				"id-01": {
					"/var/log/01/1": nil,
				},
			},
		}

		in.removePathFromTable("id-01", "/var/log/01/2")
		assert.Equal(t, out, in)
	})
}
