// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInput_parseHotData(t *testing.T) {
	type fields struct {
		mergedTags map[string]string
	}
	type args struct {
		data string
		db   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "keyname no quotation mark",
			fields: fields{
				mergedTags: map[string]string{"for": "bar"},
			},
			args: args{
				data: mockHotData01,
				db:   0,
			},
			want: []string{
				"redis_hotkey,db_name=0,for=bar key=\"/data/p/net.p\",key_count=5i,message=\"hot key  key: /data/p/net.p key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice15\",key_count=5i,message=\"hot key  key: keySlice15 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice3\",key_count=5i,message=\"hot key  key: keySlice3 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice30\",key_count=5i,message=\"hot key  key: keySlice30 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice35\",key_count=5i,message=\"hot key  key: keySlice35 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice46\",key_count=6i,message=\"hot key  key: keySlice46 key_count: 6\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice48\",key_count=5i,message=\"hot key  key: keySlice48 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice53\",key_count=5i,message=\"hot key  key: keySlice53 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice56\",key_count=5i,message=\"hot key  key: keySlice56 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice57\",key_count=5i,message=\"hot key  key: keySlice57 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice65\",key_count=5i,message=\"hot key  key: keySlice65 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice70\",key_count=5i,message=\"hot key  key: keySlice70 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice78\",key_count=5i,message=\"hot key  key: keySlice78 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice8\",key_count=5i,message=\"hot key  key: keySlice8 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice84\",key_count=5i,message=\"hot key  key: keySlice84 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice91\",key_count=5i,message=\"hot key  key: keySlice91 key_count: 5\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar keys_sampled=102i,message=\"hot key  keys_sampled: 102\",status=\"unknown\"",
			},
			wantErr: false,
		},

		{
			name: "keyname have quotation mark",
			fields: fields{
				mergedTags: map[string]string{"for": "bar"},
			},
			args: args{
				data: mockHotData02,
				db:   0,
			},
			want: []string{
				"redis_hotkey,db_name=0,for=bar key=\"keySlice0\",key_count=6i,message=\"hot key  key: keySlice0 key_count: 6\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2108\",key_count=4i,message=\"hot key  key: keySlice2108 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2182\",key_count=4i,message=\"hot key  key: keySlice2182 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2218\",key_count=4i,message=\"hot key  key: keySlice2218 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2275\",key_count=4i,message=\"hot key  key: keySlice2275 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2328\",key_count=4i,message=\"hot key  key: keySlice2328 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2340\",key_count=4i,message=\"hot key  key: keySlice2340 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2462\",key_count=4i,message=\"hot key  key: keySlice2462 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2669\",key_count=4i,message=\"hot key  key: keySlice2669 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2691\",key_count=4i,message=\"hot key  key: keySlice2691 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2791\",key_count=4i,message=\"hot key  key: keySlice2791 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2796\",key_count=4i,message=\"hot key  key: keySlice2796 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2803\",key_count=4i,message=\"hot key  key: keySlice2803 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2838\",key_count=4i,message=\"hot key  key: keySlice2838 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice2943\",key_count=4i,message=\"hot key  key: keySlice2943 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar key=\"keySlice33\",key_count=4i,message=\"hot key  key: keySlice33 key_count: 4\",status=\"unknown\"",
				"redis_hotkey,db_name=0,for=bar keys_sampled=3006i,message=\"hot key  keys_sampled: 3006\",status=\"unknown\"",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				mergedTags: tt.fields.mergedTags,
			}
			got, err := ipt.parseHotData(tt.args.data, tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.parseHotData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			gotStr := []string{}
			for _, v := range got {
				s := v.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				gotStr = append(gotStr, s)
			}
			sort.Strings(gotStr)

			assert.Equal(t, gotStr, tt.want)
		})
	}
}

var mockHotData01 = "\n# Scanning the entire keyspace to find hot keys as well as\n# average sizes per key type.  You can use -i 0.1 to sleep 0.1 sec\n# per 100 SCAN commands (not usually needed).\n\n[00.00%] Hot key 'keySlice53' found so far with counter 5\n[00.00%] Hot key 'keySlice3' found so far with counter 5\n[00.00%] Hot key 'keySlice7' found so far with counter 4\n[00.00%] Hot key 'keySlice87' found so far with counter 4\n[00.00%] Hot key 'keySlice28' found so far with counter 4\n[00.00%] Hot key 'keySlice12' found so far with counter 4\n[00.00%] Hot key '/data/p/net.p' found so far with counter 5\n[00.00%] Hot key 'keySlice85' found so far with counter 4\n[00.00%] Hot key 'keySlice2' found so far with counter 4\n[00.00%] Hot key 'keySlice52' found so far with counter 4\n[09.80%] Hot key 'keySlice23' found so far with counter 4\n[09.80%] Hot key 'keySlice91' found so far with counter 5\n[09.80%] Hot key 'keySlice8' found so far with counter 5\n[09.80%] Hot key '/data/c/net' found so far with counter 4\n[09.80%] Hot key 'keySlice14' found so far with counter 4\n[09.80%] Hot key 'keySlice48' found so far with counter 5\n[09.80%] Hot key 'keySlice57' found so far with counter 5\n[19.61%] Hot key 'keySlice84' found so far with counter 5\n[29.41%] Hot key 'keySlice56' found so far with counter 5\n[29.41%] Hot key 'keySlice78' found so far with counter 5\n[29.41%] Hot key 'keySlice65' found so far with counter 5\n[39.22%] Hot key 'keySlice30' found so far with counter 5\n[39.22%] Hot key 'keySlice15' found so far with counter 5\n[39.22%] Hot key 'keySlice35' found so far with counter 5\n[39.22%] Hot key 'keySlice70' found so far with counter 5\n[50.00%] Hot key 'keySlice17' found so far with counter 5\n[89.22%] Hot key 'keySlice46' found so far with counter 6\n\n-------- summary -------\n\nSampled 102 keys in the keyspace!\nhot key found with counter: 6\tkeyname: keySlice46\nhot key found with counter: 5\tkeyname: keySlice53\nhot key found with counter: 5\tkeyname: keySlice3\nhot key found with counter: 5\tkeyname: /data/p/net.p\nhot key found with counter: 5\tkeyname: keySlice91\nhot key found with counter: 5\tkeyname: keySlice8\nhot key found with counter: 5\tkeyname: keySlice48\nhot key found with counter: 5\tkeyname: keySlice57\nhot key found with counter: 5\tkeyname: keySlice84\nhot key found with counter: 5\tkeyname: keySlice56\nhot key found with counter: 5\tkeyname: keySlice78\nhot key found with counter: 5\tkeyname: keySlice65\nhot key found with counter: 5\tkeyname: keySlice30\nhot key found with counter: 5\tkeyname: keySlice15\nhot key found with counter: 5\tkeyname: keySlice35\nhot key found with counter: 5\tkeyname: keySlice70\n"

var mockHotData02 = "Warning: Using a password with '-a' or '-u' option on the command line interface may not be safe.\nWarning: AUTH failed\n\n# Scanning the entire keyspace to find hot keys as well as\n# average sizes per key type.  You can use -i 0.1 to sleep 0.1 sec\n# per 100 SCAN commands (not usually needed).\n\n[00.00%] Hot key '\"keySlice2052\"' found so far with counter 3\n[00.00%] Hot key '\"keySlice542\"' found so far with counter 2\n[00.00%] Hot key '\"keySlice465\"' found so far with counter 2\n[00.00%] Hot key '\"keySlice2838\"' found so far with counter 4\n[00.00%] Hot key '\"keySlice671\"' found so far with counter 2\n[00.00%] Hot key '\"keySlice1074\"' found so far with counter 2\n[00.00%] Hot key '\"keySlice670\"' found so far with counter 2\n[00.00%] Hot key '\"keySlice2669\"' found so far with counter 4\n[00.00%] Hot key '\"keySlice2340\"' found so far with counter 4\n[00.00%] Hot key '\"keySlice2943\"' found so far with counter 4\n[00.33%] Hot key '\"keySlice2462\"' found so far with counter 4\n[00.33%] Hot key '\"keySlice2275\"' found so far with counter 4\n[00.33%] Hot key '\"keySlice409\"' found so far with counter 2\n[00.33%] Hot key '\"keySlice1805\"' found so far with counter 3\n[00.33%] Hot key '\"keySlice1445\"' found so far with counter 3\n[00.33%] Hot key '\"keySlice33\"' found so far with counter 4\n[00.33%] Hot key '\"keySlice2108\"' found so far with counter 4\n[00.33%] Hot key '\"keySlice1888\"' found so far with counter 3\n[00.33%] Hot key '\"keySlice1723\"' found so far with counter 3\n[00.33%] Hot key '\"keySlice2803\"' found so far with counter 4\n[00.67%] Hot key '\"keySlice2182\"' found so far with counter 4\n[00.67%] Hot key '\"keySlice1480\"' found so far with counter 3\n[00.67%] Hot key '\"keySlice2328\"' found so far with counter 4\n[00.67%] Hot key '\"keySlice2796\"' found so far with counter 4\n[00.67%] Hot key '\"keySlice2791\"' found so far with counter 4\n[01.03%] Hot key '\"keySlice2218\"' found so far with counter 4\n[01.03%] Hot key '\"keySlice2691\"' found so far with counter 4\n[01.03%] Hot key '\"keySlice2514\"' found so far with counter 4\n[01.70%] Hot key '\"keySlice0\"' found so far with counter 6\n\n-------- summary -------\n\nSampled 3006 keys in the keyspace!\nhot key found with counter: 6\tkeyname: \"keySlice0\"\nhot key found with counter: 4\tkeyname: \"keySlice2838\"\nhot key found with counter: 4\tkeyname: \"keySlice2669\"\nhot key found with counter: 4\tkeyname: \"keySlice2340\"\nhot key found with counter: 4\tkeyname: \"keySlice2943\"\nhot key found with counter: 4\tkeyname: \"keySlice2462\"\nhot key found with counter: 4\tkeyname: \"keySlice2275\"\nhot key found with counter: 4\tkeyname: \"keySlice33\"\nhot key found with counter: 4\tkeyname: \"keySlice2108\"\nhot key found with counter: 4\tkeyname: \"keySlice2803\"\nhot key found with counter: 4\tkeyname: \"keySlice2182\"\nhot key found with counter: 4\tkeyname: \"keySlice2328\"\nhot key found with counter: 4\tkeyname: \"keySlice2796\"\nhot key found with counter: 4\tkeyname: \"keySlice2791\"\nhot key found with counter: 4\tkeyname: \"keySlice2218\"\nhot key found with counter: 4\tkeyname: \"keySlice2691\"\n"
