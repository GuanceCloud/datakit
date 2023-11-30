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

func TestInput_parseBigData(t *testing.T) {
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
			name: "ok",
			fields: fields{
				mergedTags: map[string]string{"for": "bar"},
			},
			args: args{
				data: mockBigData,
				db:   0,
			},
			want: []string{
				"redis_bigkey,db_name=0,for=bar key=\"keySlice2999\",key_type=\"string\",message=\"big key  key: keySlice2999 key_type: string value_length: 47984\",status=\"unknown\",value_length=47984i",
				"redis_bigkey,db_name=0,for=bar key=\"keyZSet2999\",key_type=\"zset\",message=\"big key  key: keyZSet2999 key_type: zset value_length: 3001\",status=\"unknown\",value_length=3001i",
				"redis_bigkey,db_name=0,for=bar key=\"myhash\",key_type=\"hash\",message=\"big key  key: myhash key_type: hash value_length: 2\",status=\"unknown\",value_length=2i",
				"redis_bigkey,db_name=0,for=bar keys_sampled=3006i,message=\"big key  keys_sampled: 3006\",status=\"unknown\"",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				mergedTags: tt.fields.mergedTags,
			}
			got, err := ipt.parseBigData(tt.args.data, tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.parseBigData() error = %v, wantErr %v", err, tt.wantErr)
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

var mockBigData = "Warning: Using a password with '-a' or '-u' option on the command line interface may not be safe.\nWarning: AUTH failed\n\n# Scanning the entire keyspace to find biggest keys as well as\n# average sizes per key type.  You can use -i 0.1 to sleep 0.1 sec\n# per 100 SCAN commands (not usually needed).\n\n[00.00%] Biggest string found so far '\"keySlice2052\"' with 32832 bytes\n[00.00%] Biggest string found so far '\"keySlice2838\"' with 45408 bytes\n[00.00%] Biggest string found so far '\"keySlice2943\"' with 47088 bytes\n[01.70%] Biggest zset   found so far '\"keySlice0\"' with 1 members\n[06.49%] Biggest string found so far '\"keySlice2974\"' with 47584 bytes\n[09.55%] Biggest hash   found so far '\"myhash\"' with 2 fields\n[13.74%] Biggest string found so far '\"keySlice2987\"' with 47792 bytes\n[17.50%] Biggest zset   found so far '\"keyZSet2999\"' with 3001 members\n[26.95%] Biggest string found so far '\"keySlice2989\"' with 47824 bytes\n[29.01%] Biggest string found so far '\"keySlice2998\"' with 47968 bytes\n[52.76%] Biggest string found so far '\"keySlice2999\"' with 47984 bytes\n\n-------- summary -------\n\nSampled 3006 keys in the keyspace!\nTotal key length in bytes is 34927 (avg len 11.62)\n\nBiggest   hash found '\"myhash\"' has 2 fields\nBiggest string found '\"keySlice2999\"' has 47984 bytes\nBiggest   zset found '\"keyZSet2999\"' has 3001 members\n\n0 lists with 0 items (00.00% of keys, avg size 0.00)\n1 hashs with 2 fields (00.03% of keys, avg size 2.00)\n3002 strings with 71976011 bytes (99.87% of keys, avg size 23976.02)\n0 streams with 0 entries (00.00% of keys, avg size 0.00)\n0 sets with 0 members (00.00% of keys, avg size 0.00)\n3 zsets with 3033 members (00.10% of keys, avg size 1011.00)\n"
