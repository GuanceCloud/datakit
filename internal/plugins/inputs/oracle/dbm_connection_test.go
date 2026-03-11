// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"database/sql"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestBuildConnectionPoints(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
		cdbName: "TESTCDB",
	}

	tests := []struct {
		name     string
		rows     []dbmConnectionRowDB
		wantLen  int
		validate func(t *testing.T, pts []*point.Point)
	}{
		{
			name:    "empty rows",
			rows:    []dbmConnectionRowDB{},
			wantLen: 0,
		},
		{
			name: "single row with all fields",
			rows: []dbmConnectionRowDB{
				{
					UserName:        sql.NullString{String: "testuser", Valid: true},
					Status:          "ACTIVE",
					PdbName:         sql.NullString{String: "TESTPDB", Valid: true},
					ConnectionCount: 10,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				fields := pt.Fields()

				assert.Equal(t, "testhost:1521", tags.GetTag("server"))
				assert.Equal(t, "TESTCDB", tags.GetTag("cdb_name"))
				assert.Equal(t, "TESTPDB", tags.GetTag("pdb_name"))
				assert.Equal(t, "testuser", tags.GetTag("username"))
				assert.Equal(t, "ACTIVE", tags.GetTag("connection_status"))

				if connCount := fields.Get("connection_count"); connCount != nil {
					assert.Equal(t, int64(10), connCount.Raw())
				}
			},
		},
		{
			name: "row with NULL pdb_name",
			rows: []dbmConnectionRowDB{
				{
					UserName:        sql.NullString{String: "testuser", Valid: true},
					Status:          "INACTIVE",
					PdbName:         sql.NullString{Valid: false},
					ConnectionCount: 5,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				// pdb_name is always set via getPdbName, which returns CDB$ROOT or cdbName when empty
				// Since isMultitenant is false by default, it returns cdbName
				assert.Equal(t, "TESTCDB", tags.GetTag("pdb_name"))
			},
		},
		{
			name: "row with NULL username",
			rows: []dbmConnectionRowDB{
				{
					UserName:        sql.NullString{Valid: false},
					Status:          "ACTIVE",
					PdbName:         sql.NullString{String: "TESTPDB", Valid: true},
					ConnectionCount: 3,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				// username should not be set when NULL
				assert.Empty(t, tags.GetTag("username"))
			},
		},
		{
			name: "filter zero connection_count",
			rows: []dbmConnectionRowDB{
				{
					UserName:        sql.NullString{String: "testuser", Valid: true},
					Status:          "ACTIVE",
					PdbName:         sql.NullString{String: "TESTPDB", Valid: true},
					ConnectionCount: 0,
				},
				{
					UserName:        sql.NullString{String: "testuser2", Valid: true},
					Status:          "ACTIVE",
					PdbName:         sql.NullString{String: "TESTPDB", Valid: true},
					ConnectionCount: 5,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				assert.Equal(t, "testuser2", tags.GetTag("username"))
			},
		},
		{
			name: "filter negative connection_count",
			rows: []dbmConnectionRowDB{
				{
					UserName:        sql.NullString{String: "testuser", Valid: true},
					Status:          "ACTIVE",
					PdbName:         sql.NullString{String: "TESTPDB", Valid: true},
					ConnectionCount: -1,
				},
			},
			wantLen: 0,
		},
		{
			name: "multiple rows",
			rows: []dbmConnectionRowDB{
				{
					UserName:        sql.NullString{String: "user1", Valid: true},
					Status:          "ACTIVE",
					PdbName:         sql.NullString{String: "PDB1", Valid: true},
					ConnectionCount: 10,
				},
				{
					UserName:        sql.NullString{String: "user2", Valid: true},
					Status:          "INACTIVE",
					PdbName:         sql.NullString{String: "PDB1", Valid: true},
					ConnectionCount: 5,
				},
				{
					UserName:        sql.NullString{String: "user1", Valid: true},
					Status:          "ACTIVE",
					PdbName:         sql.NullString{String: "PDB2", Valid: true},
					ConnectionCount: 3,
				},
			},
			wantLen: 3,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptsTime := time.Now()
			pts := ipt.buildConnectionPoints(tt.rows, ptsTime)
			assert.Len(t, pts, tt.wantLen)
			if tt.validate != nil {
				tt.validate(t, pts)
			}
		})
	}
}

func TestBuildConnectionPoints_EmptyCDBName(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
		cdbName: "",
	}

	rows := []dbmConnectionRowDB{
		{
			UserName:        sql.NullString{String: "testuser", Valid: true},
			Status:          "ACTIVE",
			PdbName:         sql.NullString{String: "TESTPDB", Valid: true},
			ConnectionCount: 10,
		},
	}

	pts := ipt.buildConnectionPoints(rows, time.Now())
	assert.Len(t, pts, 1)
	pt := pts[0]
	tags := pt.Tags()
	// cdb_name should be empty string when not set
	assert.Equal(t, "", tags.GetTag("cdb_name"))
}

func TestBuildConnectionPoints_GetPdbName(t *testing.T) {
	tests := []struct {
		name           string
		ipt            *Input
		pdbName        string
		expectedPdbTag string
	}{
		{
			name: "valid pdb_name",
			ipt: &Input{
				cdbName: "TESTCDB",
			},
			pdbName:        "TESTPDB",
			expectedPdbTag: "TESTPDB",
		},
		{
			name: "empty pdb_name with multitenant",
			ipt: &Input{
				cdbName:       "TESTCDB",
				isMultitenant: true,
			},
			pdbName:        "",
			expectedPdbTag: "CDB$ROOT",
		},
		{
			name: "empty pdb_name without multitenant",
			ipt: &Input{
				cdbName:       "TESTCDB",
				isMultitenant: false,
			},
			pdbName:        "",
			expectedPdbTag: "TESTCDB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := []dbmConnectionRowDB{
				{
					UserName:        sql.NullString{String: "testuser", Valid: true},
					Status:          "ACTIVE",
					PdbName:         sql.NullString{String: tt.pdbName, Valid: tt.pdbName != ""},
					ConnectionCount: 10,
				},
			}

			pts := tt.ipt.buildConnectionPoints(rows, time.Now())
			if len(pts) > 0 {
				pt := pts[0]
				tags := pt.Tags()
				// getPdbName is always called and pdb_name tag is always set
				// When pdbName is empty, getPdbName returns CDB$ROOT (if multitenant) or cdbName (if not multitenant)
				assert.Equal(t, tt.expectedPdbTag, tags.GetTag("pdb_name"))
			}
		})
	}
}
