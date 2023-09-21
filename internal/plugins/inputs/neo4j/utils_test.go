// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package neo4j

import (
	"reflect"
	"testing"
)

//   - 1. no additional tags
//     name: "dbms_vm_pause_time_total", prefix: "dbms_vm_pause_time_total",
//     dbms_vm_pause_time_total -> dbms_vm_pause_time_total
//
//   - 2. one additional tag have "_", with "total" suffix
//     name: "dbms_vm_gc_time_total", prefix: "dbms_vm_gc_time_", tag: []string{"gc"}, suffix: "total",
//     dbms_vm_gc_time_g1_old_generation_total->dbms_vm_gc_time_total. tags: gc->g1_old_generation
//
//   - 3. one additional tag have "_", without suffix
//     name: "dbms_vm_memory_pool", prefix: "dbms_vm_memory_pool_", tag: []string{"gc"}, suffix: "",
//     dbms_vm_memory_pool_g1_eden_space->dbms_vm_memory_pool. tags: pool->g1_eden_space
//
//   - 4. with some additional tags, with any suffix
//     name: "database_pool_used_heap", prefix: "database_", tag: []string{"db", "", "pool", "database"}, suffix: "_used_heap",
//     database_system_pool_transaction_system_used_heap -> database_pool_used_heap. tags: db->system, pool->transaction, database->system
func Test_formatFieldName(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    map[string]string
		want1   map[string]string
		wantErr bool
	}{
		{
			name:  "1. no additional tags",
			key:   "dbms_vm_pause_time_total",
			want:  map[string]string{},
			want1: map[string]string{},
		},
		{
			name:  "2. one additional tag have _, with total suffix",
			key:   "dbms_vm_gc_time_g1_old_generation_total",
			want:  map[string]string{"dbms_vm_gc_time_g1_old_generation_total": "dbms_vm_gc_time_total"},
			want1: map[string]string{"gc": "g1_old_generation"},
		},
		{
			name:  "3. one additional tag have _, without suffix",
			key:   "dbms_vm_memory_pool_g1_eden_space",
			want:  map[string]string{"dbms_vm_memory_pool_g1_eden_space": "dbms_vm_memory_pool"},
			want1: map[string]string{"pool": "g1_eden_space"},
		},
		{
			name:  "4.1 with some additional tags, with any suffix",
			key:   "database_system_pool_transaction_system_used_heap",
			want:  map[string]string{"database_system_pool_transaction_system_used_heap": "database_pool_used_heap"},
			want1: map[string]string{"db": "system", "pool": "transaction", "database": "system"},
		},
		{
			name:  "4.2 with some additional tags, with nil suffix",
			key:   "system_pool_transaction_system_used_heap",
			want:  map[string]string{"system_pool_transaction_system_used_heap": "pool_used_heap"},
			want1: map[string]string{"db": "system", "pool": "transaction", "database": "system"},
		},
		{
			name:  "4.3 with one additional tag, <bolt> is tag",
			key:   "dbms_pool_bolt_total_size",
			want:  map[string]string{"dbms_pool_bolt_total_size": "dbms_pool_total_size"},
			want1: map[string]string{"pool": "bolt"},
		},
		{
			name:    "unexpected name, not real error",
			key:     "dbms_pool_bolt_total_size_xyz",
			want:    map[string]string{},
			want1:   map[string]string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := formatFieldName(tt.key)
			if err != nil && !tt.wantErr {
				t.Errorf("formatFieldName() case = %s, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formatFieldName() case = %s, got = %v, want %v", tt.name, got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("formatFieldName() case = %s, got1 = %v, want %v", tt.name, got1, tt.want1)
			}
		})
	}
}
