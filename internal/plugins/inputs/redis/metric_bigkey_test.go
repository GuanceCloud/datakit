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

		{
			name: "v6.0.8",
			fields: fields{
				mergedTags: map[string]string{"for": "bar"},
			},
			args: args{
				data: mockBigData6_0_8,
				db:   0,
			},
			want: []string{
				"redis_bigkey,db_name=0,for=bar key=\"wo18127.0.0.1:6379\",key_type=\"string\",message=\"big key  key: wo18127.0.0.1:6379 key_type: string value_length: 288\",status=\"unknown\",value_length=288i",
				"redis_bigkey,db_name=0,for=bar keys_sampled=6i,message=\"big key  keys_sampled: 6\",status=\"unknown\"",
			},
			wantErr: false,
		},

		{
			name: "v7.0",
			fields: fields{
				mergedTags: map[string]string{"for": "bar"},
			},
			args: args{
				data: mockBigData7_0,
				db:   0,
			},
			want: []string{
				"redis_bigkey,db_name=0,for=bar key=\"wo19\",key_type=\"string\",message=\"big key  key: wo19 key_type: string value_length: 358\",status=\"unknown\",value_length=358i",
				"redis_bigkey,db_name=0,for=bar keys_sampled=20i,message=\"big key  keys_sampled: 20\",status=\"unknown\"",
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

var mockBigData = `
Warning: Using a password with '-a' or '-u' option on the command line interface may not be safe.
Warning: AUTH failed

# Scanning the entire keyspace to find biggest keys as well as
# average sizes per key type.  You can use -i 0.1 to sleep 0.1 sec
# per 100 SCAN commands (not usually needed).

[00.00%] Biggest string found so far '"keySlice2052"' with 32832 bytes
[00.00%] Biggest string found so far '"keySlice2838"' with 45408 bytes
[00.00%] Biggest string found so far '"keySlice2943"' with 47088 bytes
[01.70%] Biggest zset   found so far '"keySlice0"' with 1 members
[06.49%] Biggest string found so far '"keySlice2974"' with 47584 bytes
[09.55%] Biggest hash   found so far '"myhash"' with 2 fields
[13.74%] Biggest string found so far '"keySlice2987"' with 47792 bytes
[17.50%] Biggest zset   found so far '"keyZSet2999"' with 3001 members
[26.95%] Biggest string found so far '"keySlice2989"' with 47824 bytes
[29.01%] Biggest string found so far '"keySlice2998"' with 47968 bytes
[52.76%] Biggest string found so far '"keySlice2999"' with 47984 bytes

-------- summary -------

Sampled 3006 keys in the keyspace!
Total key length in bytes is 34927 (avg len 11.62)

Biggest   hash found '"myhash"' has 2 fields
Biggest string found '"keySlice2999"' has 47984 bytes
Biggest   zset found '"keyZSet2999"' has 3001 members

0 lists with 0 items (00.00% of keys, avg size 0.00)
1 hashs with 2 fields (00.03% of keys, avg size 2.00)
3002 strings with 71976011 bytes (99.87% of keys, avg size 23976.02)
0 streams with 0 entries (00.00% of keys, avg size 0.00)
0 sets with 0 members (00.00% of keys, avg size 0.00)
3 zsets with 3033 members (00.10% of keys, avg size 1011.00)
`

var mockBigData6_0_8 = `
Warning: Using a password with '-a' or '-u' option on the command line interface may not be safe.

# Scanning the entire keyspace to find biggest keys as well as
# average sizes per key type.  You can use -i 0.1 to sleep 0.1 sec
# per 100 SCAN commands (not usually needed).

[00.00%] Biggest string found so far '"wo15127.0.0.1:6379"' with 240 bytes
[00.00%] Biggest string found so far '"wo17127.0.0.1:6379"' with 272 bytes
[00.00%] Biggest string found so far '"wo18127.0.0.1:6379"' with 288 bytes

-------- summary -------

Sampled 6 keys in the keyspace!
Total key length in bytes is 105 (avg len 17.50)

Biggest string found '"wo18127.0.0.1:6379"' has 288 bytes

0 lists with 0 items (00.00% of keys, avg size 0.00)
0 hashs with 0 fields (00.00% of keys, avg size 0.00)
6 strings with 992 bytes (100.00% of keys, avg size 165.33)
0 streams with 0 entries (00.00% of keys, avg size 0.00)
0 sets with 0 members (00.00% of keys, avg size 0.00)
0 zsets with 0 members (00.00% of keys, avg size 0.00)
`

var mockBigData7_0 = `
Warning: Using a password with '-a' or '-u' option on the command line interface may not be safe.

# Scanning the entire keyspace to find biggest keys as well as
# average sizes per key type.  You can use -i 0.1 to sleep 0.1 sec
# per 100 SCAN commands (not usually needed).

[00.00%] Biggest string found so far '"wo12"' with 246 bytes
[00.00%] Biggest string found so far '"wo19"' with 358 bytes

-------- summary -------

Sampled 20 keys in the keyspace!
Total key length in bytes is 70 (avg len 3.50)

Biggest string found '"wo19"' has 358 bytes

0 lists with 0 items (00.00% of keys, avg size 0.00)
0 hashs with 0 fields (00.00% of keys, avg size 0.00)
20 strings with 4119 bytes (100.00% of keys, avg size 205.95)
0 streams with 0 entries (00.00% of keys, avg size 0.00)
0 sets with 0 members (00.00% of keys, avg size 0.00)
0 zsets with 0 members (00.00% of keys, avg size 0.00)
`
