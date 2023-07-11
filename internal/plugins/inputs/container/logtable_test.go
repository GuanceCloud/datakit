// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogTableFindDifferences(t *testing.T) {
	testcases := []struct {
		inTable *logTable
		inIDs   []string
		out     []string
	}{
		{
			inTable: &logTable{
				table: map[string]map[string]chan interface{}{
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
				table: map[string]map[string]chan interface{}{
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
				table: map[string]map[string]chan interface{}{
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
				table: map[string]map[string]chan interface{}{},
			},
			inIDs: []string{
				"id-01",
			},
			out: nil,
		},
		{
			inTable: &logTable{
				table: map[string]map[string]chan interface{}{},
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
			table: map[string]map[string]chan interface{}{
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
			},
		}

		out := "{id:id-01,paths:[/var/log/01/1,/var/log/01/2]}, {id:id-02,paths:[/var/log/02/1,/var/log/02/2]}, {id:id-03,paths:[/var/log/03/1,/var/log/03/2]}"

		assert.Equal(t, out, in.String())
	})
}

func TestLogTableRemoveID(t *testing.T) {
	t.Run("logtable-remove-path", func(t *testing.T) {
		in := &logTable{
			table: map[string]map[string]chan interface{}{
				"id-01": {
					"/var/log/01/1": nil,
				},
				"id-02": {
					"/var/log/02/1": nil,
				},
			},
		}

		out := &logTable{
			table: map[string]map[string]chan interface{}{
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
			table: map[string]map[string]chan interface{}{
				"id-01": {
					"/var/log/01/1": nil,
					"/var/log/01/2": nil,
				},
			},
		}

		out := &logTable{
			table: map[string]map[string]chan interface{}{
				"id-01": {
					"/var/log/01/1": nil,
				},
			},
		}

		in.removePathFromTable("id-01", "/var/log/01/2")
		assert.Equal(t, out, in)
	})
}
