//go:build linux
// +build linux

package protodec

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func TestPrefix(t *testing.T) {
	t.Run("POST", func(t *testing.T) {
		data := []byte("POST / HTTP/1.1\r\n\r\n")
		v, ok := httpMethod(data)
		if !ok {
			t.Fatal("not req start")
		}
		assert.Equal(t, "POST", v)
	})

	t.Run("GET", func(t *testing.T) {
		data := []byte("GET / HTTP/1.1\r\n\r\n")
		v, ok := httpMethod(data)
		if !ok {
			t.Fatal("not req start")
		}
		assert.Equal(t, "GET", v)
	})

	t.Run("HTTP", func(t *testing.T) {
		data := []byte("HTTP/1.1 200 OK\r\n\r\n")
		v, code, ok := httpProtoVersion(data)
		if !ok {
			t.Fatal("not resp start")
		}
		assert.Equal(t, "1.1", v)
		assert.Equal(t, 200, code)
	})
}

func TestMysql(t *testing.T) {
	m := &mysqlInfo{}
	t.Run("DetectMysql", func(t *testing.T) {
		mysqlReqMsg := "\x17\x00\x00\x00\x03\x43\x52\x45\x41\x54\x45\x20\x44\x41\x54\x41\x42\x41\x53\x45\x20" +
			"\x74\x65\x73\x74\x64\x62"
		seq, err := detectMysql([]byte(mysqlReqMsg), len(mysqlReqMsg))
		expectedSeq := 0
		if err != nil {
			t.Fatal("not client req create")
		}
		assert.Equal(t, expectedSeq, int(seq))
	})

	t.Run("Client", func(t *testing.T) {
		mysqlReqMsg := "\x03\x43\x52\x45\x41\x54\x45\x20\x44\x41\x54\x41\x42\x41\x53\x45\x20" +
			"\x74\x65\x73\x74\x64\x62"
		expected := "CREATE DATABASE testdb"
		msg, stmtId, err := m.isClientMsg([]byte(mysqlReqMsg), comm.FnSysRecvfrom)
		if err != nil {
			t.Fatal("not client req")
		}
		assert.Equal(t, expected, msg)
		assert.Equal(t, 0, stmtId)
	})

	t.Run("InsertPreStmt", func(t *testing.T) {
		insertReqMsg := "\x16\x49\x4e\x53\x45\x52\x54\x20\x49\x4e\x54\x4f\x20" +
			"\x63\x69\x74\x69\x65\x73\x28\x6e\x61\x6d\x65\x2c\x20\x70\x6f\x70\x75\x6c\x61\x74\x69" +
			"\x6f\x6e\x29\x20\x56\x41\x4c\x55\x45\x53\x28\x3f\x2c\x20\x3f\x29\x3b"
		expected := "INSERT INTO cities(name, population) VALUES(?, ?);"
		msg, stmtId, err := m.isClientMsg([]byte(insertReqMsg), comm.FnSysRecvfrom)
		if err != nil {
			t.Fatal("not client req")
		}
		assert.Equal(t, expected, msg)
		assert.Equal(t, -1, stmtId)
	})

	t.Run("InsertPreStmtServer", func(t *testing.T) {
		insertResMsg := "\x00\x01\x00\x00\x00\x00\x00\x02\x00\x00\x00\x00"
		msg, code, err := m.isServerMsg([]byte(insertResMsg), comm.FnSysSendto)
		if err != nil {
			t.Fatal("not server req")
		}
		t.Log(msg)
		assert.Equal(t, 0, code)
	})

	t.Run("Server", func(t *testing.T) {
		mysqlRespMsg := "\xff\xef\x03\x23\x48\x59\x30\x30\x30\x43\x61\x6e\x27\x74\x20" +
			"\x63\x72\x65\x61\x74\x65\x20\x64\x61\x74\x61\x62\x61\x73\x65\x20\x27\x74\x65" +
			"\x73\x74\x64\x62\x27\x3b\x20\x64\x61\x74\x61\x62\x61\x73\x65\x20\x65\x78\x69" +
			"\x73\x74\x73"
		expectedMsg := "Error Code: 1007, Error Msg: Can't create database 'testdb'; database exists"
		expectedCode := 1007
		msg, code, err := m.isServerMsg([]byte(mysqlRespMsg), 0)
		if err != nil {
			t.Fatal("not server req")
		}
		assert.Equal(t, expectedCode, code)
		assert.Equal(t, expectedMsg, msg)
	})
}

func TestRedis(t *testing.T) {
	t.Run("TestReadLength", func(t *testing.T) {
		msg := "\x2a\x33\x0d\x0a\x24\x33\x0d\x0a\x73\x65\x74\x0d\x0a\x24" +
			"\x34\x0d\x0a\x6b\x65\x79\x30\x0d\x0a\x24\x36\x0d\x0a\x76\x61" +
			"\x6c\x75\x65\x30\x0d\x0a"

		payload, length, err := readLength([]byte(msg[1:]))
		if err != nil {
			t.Fatal("read length error")
		}

		assert.Equal(t, []byte(msg[4:]), payload)
		assert.Equal(t, 3, length)
	})

	t.Run("TestDecodeBulkString", func(t *testing.T) {
		msg := "\x24\x33\x0d\x0a\x73\x65\x74\x0d\x0a\x24" +
			"\x34\x0d\x0a\x6b\x65\x79\x30\x0d\x0a\x24\x36\x0d\x0a\x76\x61" +
			"\x6c\x75\x65\x30\x0d\x0a"

		expectedPayload := "\x24\x34\x0d\x0a\x6b\x65\x79\x30\x0d\x0a\x24\x36" +
			"\x0d\x0a\x76\x61\x6c\x75\x65\x30\x0d\x0a"
		expectedCommand := "\x73\x65\x74"
		r := &redisInfo{}
		payload, command, err := r.decodeBulkString([]byte(msg))
		if err != nil {
			t.Fatal("read msg error")
		}

		assert.Equal(t, []byte(expectedPayload), payload)
		assert.Equal(t, []byte(expectedCommand), command)
	})

	t.Run("TestParseRequest", func(t *testing.T) {
		msg := "\x2a\x33\x0d\x0a\x24\x33\x0d\x0a\x73\x65\x74\x0d\x0a\x24" +
			"\x34\x0d\x0a\x6b\x65\x79\x30\x0d\x0a\x24\x36\x0d\x0a\x76\x61" +
			"\x6c\x75\x65\x30\x0d\x0a"

		expectedPayload := "\x24\x33\x0d\x0a\x73\x65\x74\x0d\x0a\x24" +
			"\x34\x0d\x0a\x6b\x65\x79\x30\x0d\x0a\x24\x36\x0d\x0a\x76\x61" +
			"\x6c\x75\x65\x30\x0d\x0a"
		expectedCmd := "SET"
		r := &redisInfo{}
		err := r.parseRequest([]byte(msg))
		if err != nil {
			t.Fatalf("parse request err %v", err)
		}

		assert.Equal(t, []byte(expectedPayload), r.payload)
		assert.Equal(t, expectedCmd, r.cmd)
	})

	t.Run("TestParseRequest_Req", func(t *testing.T) {
		msg := "\x2b\x4f\x4b\x0d\x0a"

		r := &redisInfo{}
		err := r.parseRequest([]byte(msg))
		if err == nil {
			t.Fatalf("should have error")
		}
	})
}

func TestDecodeRedis(t *testing.T) {
	type rq struct {
		resource string
		expected string
	}

	cases := []rq{
		{
			resource: "+OK\r\n",
			expected: "+OK",
		},
	}

	r := &redisInfo{}
	for _, c := range cases {
		t.Run(c.resource, func(t *testing.T) {
			output := []byte{}
			_, err := r.decodeRespType(&output, []byte(c.resource))
			if err != nil {
				t.Fatalf("error occur %v", err)
			}
			assert.Equal(t, c.expected, string(output))
		})
	}
}

func TestRedisParseResponse(t *testing.T) {
	type rq struct {
		resource string
		expected string
	}

	cases := []rq{
		{
			resource: "*-1\r\n",
			expected: "",
		},
		{
			resource: "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n",
			expected: "",
		},
		{
			resource: "$0\r\n\r\n",
			expected: "",
		},
		{
			resource: "!9\r\nabcdefghi\r\n",
			expected: "!abcdefghi",
		},
		{
			resource: "*3\r\n$3\r\nset\r\n$4\r\nkey3\r\n$6\r\nvalue3\r\n",
			expected: "",
		},
	}

	r := &redisInfo{}
	for _, c := range cases {
		t.Run(c.resource, func(t *testing.T) {
			err := r.parseResponse([]byte(c.resource))
			if err != nil {
				t.Fatalf("error occur %v", err)
			}
			assert.Equal(t, c.expected, string(r.payload))
		})
	}
}

func TestRedisStringify(t *testing.T) {
	type cases struct {
		resource string
		expected string
	}

	encode := func(resource string) []byte {
		args := strings.Split(resource, " ")
		n := len(args)

		var output []byte
		header := fmt.Sprintf("*%d\r\n", n)
		output = append(output, []byte(header)...)

		for _, arg := range args {
			toAppend := fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
			output = append(output, []byte(toAppend)...)
		}

		return output
	}

	tc := []cases{
		{
			resource: "GET KEY ",
			expected: "GET KEY",
		},
		{
			resource: "AUTH my-password",
			expected: "AUTH ?",
		},
		{
			resource: "AUTH default my-password",
			expected: "AUTH ?",
		},
		{
			resource: "HELLO 3 AUTH default my-password SETNAME client",
			expected: "HELLO 3 AUTH ?",
		},
		{
			resource: "APPEND key value",
			expected: "APPEND key ?",
		},
		{
			resource: "GETSET key value",
			expected: "GETSET key ?",
		},
		{
			resource: "LPUSHX key value",
			expected: "LPUSHX key ?",
		},
		{
			resource: "GEOADD Sicily 13.583333 37.316667 'Agrigento'",
			expected: "GEOADD Sicily 13.583333 37.316667 ?",
		},
		{
			resource: "GEORADIUSBYMEMBER Sicily Agrigento 100 km",
			expected: "GEORADIUSBYMEMBER Sicily ? 100 km",
		},
		{
			resource: "RPUSHX mylist 'World'",
			expected: "RPUSHX mylist ?",
		},
		{
			resource: "SET mykey 'Hello'",
			expected: "SET mykey ?",
		},
		{
			resource: "SETNX mykey 'Hello'",
			expected: "SETNX mykey ?",
		},
		{
			resource: "SISMEMBER myset 'one'",
			expected: "SISMEMBER myset ?",
		},
		{
			resource: "ZRANK myzset 'three'",
			expected: "ZRANK myzset ?",
		},
		{
			resource: "ZRANK myzset 'three' WITHSCORE",
			expected: "ZRANK myzset ? WITHSCORE",
		},
		{
			resource: "HSETNX myhash field 'World'",
			expected: "HSETNX myhash field ?",
		},
		{
			resource: "LREM mylist -2 'hello'",
			expected: "LREM mylist -2 ?",
		},
		{
			resource: "LINSERT mylist BEFORE 'World' 'There'",
			expected: "LINSERT mylist BEFORE 'World' ?",
		},
		{
			resource: "GEOHASH Sicily Palermo Catania",
			expected: "GEOHASH Sicily ?",
		},
		{
			resource: "HSET myhash field1 'Hello'",
			expected: "HSET myhash field1 ?",
		},
		{
			resource: "MSET key1 'Hello' key2 'World'",
			expected: "MSET key1 ? key2 ?",
		},
		{
			resource: "BITFIELD key GET type offset SET type offset value INCRBY type",
			expected: "BITFIELD key GET type offset SET type offset ? INCRBY type",
		},
	}

	for _, c := range tc {
		t.Run(c.resource, func(t *testing.T) {
			cmd := encode(c.resource)
			r := &redisInfo{}
			err := r.parseRequest(cmd)
			if err != nil {
				t.Fatalf("parse cmd error %v", err)
			}
			output := r.stringify()
			assert.Equal(t, c.expected, string(output))
		})
	}
}
