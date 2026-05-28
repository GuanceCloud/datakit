//go:build linux
// +build linux

package protodec

import (
	"errors"
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

func TestHTTPDecPipeExportsRequestResponseFields(t *testing.T) {
	req := []byte("GET /api/v1/users?token=secret HTTP/1.1\r\n" +
		"Host: example.com\r\n" +
		"traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01\r\n\r\n")
	resp := []byte("HTTP/1.1 201 Created\r\nContent-Length: 0\r\n\r\n")

	dec := &httpDecPipe{}
	dec.Decode(comm.NICDEgress, &comm.NetwrkData{
		Payload:     req,
		CaptureSize: len(req),
		TCPSeq:      100,
		TS:          1000,
		TSTail:      1100,
	}, 123456789, nil)
	dec.Decode(comm.NICDIngress, &comm.NetwrkData{
		Payload:     resp,
		CaptureSize: len(resp),
		TCPSeq:      200,
		TS:          1500,
		TSTail:      1700,
	}, 0, nil)

	pts := dec.Export(true)
	if len(pts) != 1 {
		t.Fatalf("exported HTTP records = %d, want 1", len(pts))
	}
	got := pts[0]
	assert.Equal(t, ProtoHTTP, got.L7Proto)
	assert.Equal(t, comm.DOut, got.Direction)
	assert.Equal(t, int64(600), got.Duration)
	assert.Equal(t, int64(600), got.Cost)
	assert.Equal(t, int64(123456789), got.Time)
	assert.Equal(t, int64(len(resp)), got.KVs.Get(comm.FieldBytesRead).GetI())
	assert.Equal(t, int64(len(req)), got.KVs.Get(comm.FieldBytesWritten).GetI())
	assert.Equal(t, "GET", got.KVs.Get(comm.FieldHTTPMethod).GetS())
	assert.Equal(t, "/api/v1/users", got.KVs.Get(comm.FieldHTTPRoute).GetS())
	assert.Equal(t, "example.com", got.KVs.Get(comm.FieldHTTPHost).GetS())
	assert.Equal(t, "1.1", got.KVs.Get(comm.FieldHTTPVersion).GetS())
	assert.Equal(t, "201", got.KVs.Get(comm.FieldHTTPStatusCode).GetS())
	assert.Equal(t, "ok", got.KVs.Get(comm.FieldStatus).GetS())
	assert.Equal(t, "GET /api/v1/users", got.KVs.Get(comm.FieldResource).GetS())
	assert.True(t, got.Meta.SampledSpan)
	assert.True(t, got.Meta.SpanHexEnc)
	assert.NotZero(t, got.Meta.TraceID)
	assert.NotZero(t, got.Meta.ParentSpanID)
	assert.Equal(t, uint32(100), got.Meta.ReqTCPSeq)
	assert.Equal(t, uint32(200), got.Meta.RespTCPSeq)
}

func TestSubProtoSetKeepsDetectorOrder(t *testing.T) {
	Init()
	pset := SubProtoSet(ProtoRedis, ProtoHTTP, ProtoMySQL)
	assert.Equal(t, []L7Protocol{ProtoRedis, ProtoHTTP, ProtoMySQL}, pset.protoOrder)
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

	t.Run("TestReadLengthTooLarge", func(t *testing.T) {
		msg := fmt.Sprintf("%d\r\n", uint64(maxInt)+1)
		if _, _, err := readLength([]byte(msg)); err == nil {
			t.Fatal("expected oversized RESP length to be rejected")
		}
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

func TestMysqlHeaderDecoder(t *testing.T) {
	insertReqMsg := "\x1e\x00\x00\x00\x16\x49\x4e\x53\x45\x52\x54\x20\x49\x4e\x54\x4f\x20" +
		"\x63\x69\x74\x69\x65\x73\x28\x6e\x61\x6d\x65\x2c\x20\x70\x6f\x70\x75\x6c\x61\x74\x69" +
		"\x6f\x6e\x29\x20\x56\x41\x4c\x55\x45\x53\x28\x3f\x2c\x20\x3f\x29\x3b"

	hd := &headerDecoder{}
	offset, err := hd.decode([]byte(insertReqMsg), 0)
	if err != nil {
		t.Fatalf("error occurred %v", err)
	}
	expectedOffset := 4
	assert.Equal(t, expectedOffset, offset)
}

func TestMysqlPendingBufferLimit(t *testing.T) {
	m := &mysqlInfo{}
	_, err := m.parse([]byte("123456"), 0, 5)
	if !errors.Is(err, errMysqlPendingBufferLimit) {
		t.Fatalf("expected pending buffer limit error, got %v", err)
	}
	if m.reader == nil || m.reader.Len() != 0 {
		t.Fatal("expected mysql pending reader to be reset")
	}
}

func TestMysqlPendingBufferLimitEnv(t *testing.T) {
	t.Setenv(mysqlPendingBufferLimitEnv, "bad")
	if got := mysqlPendingBufferLimit(); got != defaultMysqlPendingBufferLimit {
		t.Fatalf("invalid env limit = %d, want %d", got, defaultMysqlPendingBufferLimit)
	}

	t.Setenv(mysqlPendingBufferLimitEnv, "99999999")
	if got := mysqlPendingBufferLimit(); got != maxMysqlPendingBufferLimit {
		t.Fatalf("clamped env limit = %d, want %d", got, maxMysqlPendingBufferLimit)
	}
}

func TestTrimComment(t *testing.T) {
	type cases struct {
		resource        string
		expectedCommand string
		expectedComment string
	}

	tc := []cases{
		{
			resource:        "sELECT 1",
			expectedCommand: "SELECT",
			expectedComment: "",
		},
		{
			resource:        "/* comment */ SELECT 1 FROM TABLE",
			expectedCommand: "SELECT",
			expectedComment: "/* comment */ ",
		},
		{
			resource:        "/* i am comment */ /*i am comment*/SelecT 1",
			expectedCommand: "SELECT",
			expectedComment: "/* i am comment */ /*i am comment*/",
		},
		{
			resource:        "/* spanID: 1234567 */ SELECT * from TABLE",
			expectedCommand: "SELECT",
			expectedComment: "/* spanID: 1234567 */ ",
		},
		{
			resource:        "\x00\x01/* spanID: 1234567 */ SELECT * from TABLE",
			expectedCommand: "SELECT",
			expectedComment: "/* spanID: 1234567 */ ",
		},
	}

	for _, c := range tc {
		t.Run(c.resource, func(t *testing.T) {
			payload := readMysql([]byte(c.resource))
			comment, output, _ := trimCommentGetFirst(payload, 6)
			t.Log(string(comment))
			t.Log(string(output))
			m := &mysqlInfo{}
			assert.Equal(t, true, m.isValidSQL(output))
			assert.Equal(t, c.expectedCommand, strings.ToUpper(string(output)))
			assert.Equal(t, c.expectedComment, string(comment))
		})
	}
}

func TestCheckmysql(t *testing.T) {
	type cases struct {
		msg string
	}

	tc := []cases{
		{
			msg: "\x1e\x00\x00\x00\x16\x49\x4e\x53\x45\x52\x54\x20\x49\x4e\x54\x4f\x20" +
				"\x63\x69\x74\x69\x65\x73\x28\x6e\x61\x6d\x65\x2c\x20\x70\x6f\x70\x75\x6c\x61\x74\x69" +
				"\x6f\x6e\x29\x20\x56\x41\x4c\x55\x45\x53\x28\x3f\x2c\x20\x3f\x29\x3b",
		},
	}

	for _, c := range tc {
		t.Run("testing", func(t *testing.T) {
			isMysql := checkMysql([]byte(c.msg))
			assert.Equal(t, true, isMysql)
		})
	}
}

func TestCheckMysqlHeader(t *testing.T) {
	type cases struct {
		name               string
		resource           string
		direction          Direction
		expectedPacketType mysqlPacketType
	}

	tc := []cases{
		{
			name: "INSERT",
			resource: "\x1e\x00\x00\x00\x16\x49\x4e\x53\x45\x52\x54\x20\x49\x4e\x54\x4f\x20" +
				"\x63\x69\x74\x69\x65\x73\x28\x6e\x61\x6d\x65\x2c\x20\x70\x6f\x70\x75\x6c\x61\x74\x69" +
				"\x6f\x6e\x29\x20\x56\x41\x4c\x55\x45\x53\x28\x3f\x2c\x20\x3f\x29\x3b",
			expectedPacketType: packetRequest,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			hd := &headerDecoder{}
			offset, err := hd.decode([]byte(c.resource), 0)
			if err != nil {
				t.Fatalf("error occurred %v", err)
			}

			packetType, err := hd.checkHeader([]byte(c.resource), offset)
			if err != nil {
				t.Fatalf("error occurred %v", err)
			}

			assert.Equal(t, c.expectedPacketType, packetType)
		})
	}
}

func TestParseRequest(t *testing.T) {
	type cases struct {
		name               string
		resource           string
		expectedPacketType mysqlPacketType
		expectedResource   string
	}

	tc := []cases{
		{
			name: "INSERT",
			resource: "\x1e\x00\x00\x00\x16\x49\x4e\x53\x45\x52\x54\x20\x49\x4e\x54\x4f\x20" +
				"\x63\x69\x74\x69\x65\x73\x28\x6e\x61\x6d\x65\x2c\x20\x70\x6f\x70\x75\x6c\x61\x74\x69" +
				"\x6f\x6e\x29\x20\x56\x41\x4c\x55\x45\x53\x28\x3f\x2c\x20\x3f\x29\x3b",
			expectedPacketType: packetRequest,
			expectedResource:   "INSERT INTO cities ( name, population ) VALUES ( ? )",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			m := &mysqlInfo{}
			packetType, err := m.parseRequest([]byte(c.resource[4:]))
			if err != nil {
				t.Fatalf("error occurred %v", err)
			}

			assert.Equal(t, c.expectedPacketType, packetType)
			assert.Equal(t, c.expectedResource, m.resource)
		})
	}
}
