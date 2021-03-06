// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package obfuscate

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

type compactSpacesTestCase struct {
	before string
	after  string
}

func TestMain(m *testing.M) {
	flag.Parse()

	// disable loggging in tests
	seelog.UseLogger(seelog.Disabled)

	// prepare JSON obfuscator tests
	suite, err := loadTests()
	if err != nil {
		log.Fatalf("Failed to load JSON obfuscator tests: %s", err.Error())
	}
	if len(suite) == 0 {
		log.Fatal("no tests in suite")
	}
	jsonSuite = suite

	os.Exit(m.Run())
}

func TestNewObfuscator(t *testing.T) {
	assert := assert.New(t)
	o := NewObfuscator(nil)
	assert.Nil(o.es)
	assert.Nil(o.mongo)

	o = NewObfuscator(nil)
	assert.Nil(o.es)
	assert.Nil(o.mongo)

	o = NewObfuscator(&Config{
		ES:    JSONConfig{Enabled: true},
		Mongo: JSONConfig{Enabled: true},
	})
	assert.NotNil(o.es)
	assert.NotNil(o.mongo)
}

func TestCompactWhitespaces(t *testing.T) {
	assert := assert.New(t)

	resultsToExpect := []compactSpacesTestCase{
		{
			"aa",
			"aa",
		},

		{
			" aa bb",
			"aa bb",
		},

		{
			"aa    bb  cc  dd ",
			"aa bb cc dd",
		},

		{
			"    ",
			"",
		},

		{
			"a b       cde     fg       hi                     j  jk   lk lkjfdsalfd     afsd sfdafsd f",
			"a b cde fg hi j jk lk lkjfdsalfd afsd sfdafsd f",
		},

		{
			"   ¡™£¢∞§¶    •ªº–≠œ∑´®†¥¨ˆøπ “‘«åß∂ƒ©˙∆˚¬…æΩ≈ç√ ∫˜µ≤≥÷    ",
			"¡™£¢∞§¶ •ªº–≠œ∑´®†¥¨ˆøπ “‘«åß∂ƒ©˙∆˚¬…æΩ≈ç√ ∫˜µ≤≥÷",
		},
	}

	for _, testCase := range resultsToExpect {
		assert.Equal(testCase.after, compactWhitespaces(testCase.before))
	}
}

func TestReplaceDigits(t *testing.T) {
	assert := assert.New(t)

	for _, tt := range []struct {
		in       []byte
		expected []byte
	}{
		{
			[]byte("table123"),
			[]byte("table?"),
		},
		{
			[]byte(""),
			[]byte(""),
		},
		{
			[]byte("2020-table"),
			[]byte("?-table"),
		},
		{
			[]byte("sales_2019_07_01"),
			[]byte("sales_?_?_?"),
		},
		{
			[]byte("45"),
			[]byte("?"),
		},
	} {
		assert.Equal(tt.expected, replaceDigits(tt.in))
	}
}

func TestObfuscateStatsGroup(t *testing.T) {
	o := NewObfuscator(nil)
	for _, tt := range []struct {
		typ, resource string
		out           string // output obfuscated resource
	}{
		{"sql", "SELECT 1 FROM db", "SELECT ? FROM db"},
		{"sql", "SELECT 1\nFROM Blogs AS [b\nORDER BY [b]", nonParsableResource},
		{"redis", "ADD 1, 2", "ADD"},
		{"other", "ADD 1, 2", "ADD 1, 2"},
	} {
		out := o.ObfuscateStatsGroup(tt.typ, tt.resource)
		assert.Equal(t, out, tt.out)
	}
}

// TestObfuscateDefaults ensures that running the obfuscator with no config continues to obfuscate/quantize
// SQL queries and Redis commands in span resources.
func TestObfuscateDefaults(t *testing.T) {
	t.Run("redis", func(t *testing.T) {
		cmd := "SET k v\nGET k"
		out, err := NewObfuscator(nil).Obfuscate("redis", cmd)
		assert.NoError(t, err)
		assert.Equal(t, out.Query, out.Query)
	})

	t.Run("sql", func(t *testing.T) {
		query := "UPDATE users(name) SET ('Jim')"
		out, err := NewObfuscator(nil).Obfuscate("sql", query)
		assert.NoError(t, err)
		assert.Equal(t, "UPDATE users ( name ) SET ( ? )", out.Query)
	})
}

func TestObfuscateConfig(t *testing.T) {
	// testConfig returns a test function which creates a span of type typ,
	// having a tag with key/val, runs the obfuscator on it using the given
	// configuration and asserts that the new tag value matches exp.
	testConfig := func(
		typ, key, val, exp string,
		cfg *Config,
	) func(*testing.T) {
		return func(t *testing.T) {
			out, err := NewObfuscator(cfg).Obfuscate(typ, val)
			assert.NoError(t, err)
			assert.Equal(t, exp, out.Query)
		}
	}

	t.Run("redis/enabled", testConfig(
		"redis",
		"redis.raw_command",
		"SET key val",
		"SET key ?",
		&Config{Redis: Enablable{Enabled: true}},
	))

	t.Run("redis/disabled", testConfig(
		"redis",
		"redis.raw_command",
		"SET key val",
		"SET key val",
		&Config{},
	))

	t.Run("http/enabled", testConfig(
		"http",
		"http.url",
		"http://mysite.mydomain/1/2?q=asd",
		"http://mysite.mydomain/?/??",
		&Config{HTTP: HTTPConfig{
			RemovePathDigits:  true,
			RemoveQueryString: true,
		}},
	))

	t.Run("http/disabled", testConfig(
		"http",
		"http.url",
		"http://mysite.mydomain/1/2?q=asd",
		"http://mysite.mydomain/1/2?q=asd",
		&Config{},
	))

	t.Run("web/enabled", testConfig(
		"web",
		"http.url",
		"http://mysite.mydomain/1/2?q=asd",
		"http://mysite.mydomain/?/??",
		&Config{HTTP: HTTPConfig{
			RemovePathDigits:  true,
			RemoveQueryString: true,
		}},
	))

	t.Run("web/disabled", testConfig(
		"web",
		"http.url",
		"http://mysite.mydomain/1/2?q=asd",
		"http://mysite.mydomain/1/2?q=asd",
		&Config{},
	))

	t.Run("json/enabled", testConfig(
		"elasticsearch",
		"elasticsearch.body",
		`{"role": "database"}`,
		`{"role":"?"}`,
		&Config{
			ES: JSONConfig{Enabled: true},
		},
	))

	t.Run("json/disabled", testConfig(
		"elasticsearch",
		"elasticsearch.body",
		`{"role": "database"}`,
		`{"role": "database"}`,
		&Config{},
	))

	t.Run("memcached/enabled", testConfig(
		"memcached",
		"memcached.command",
		"set key 0 0 0\r\nvalue",
		"set key 0 0 0",
		&Config{Memcached: Enablable{Enabled: true}},
	))

	t.Run("memcached/disabled", testConfig(
		"memcached",
		"memcached.command",
		"set key 0 0 0 noreply\r\nvalue",
		"set key 0 0 0 noreply\r\nvalue",
		&Config{},
	))
}

func TestLiteralEscapes(t *testing.T) {
	o := NewObfuscator(nil)

	t.Run("default", func(t *testing.T) {
		assert.False(t, o.SQLLiteralEscapes())
	})

	t.Run("true", func(t *testing.T) {
		o.SetSQLLiteralEscapes(true)
		assert.True(t, o.SQLLiteralEscapes())
	})

	t.Run("false", func(t *testing.T) {
		o.SetSQLLiteralEscapes(false)
		assert.False(t, o.SQLLiteralEscapes())
	})
}

func BenchmarkCompactWhitespaces(b *testing.B) {
	str := "a b       cde     fg       hi                     j  jk   lk lkjfdsalfd     afsd sfdafsd f"
	for i := 0; i < b.N; i++ {
		compactWhitespaces(str)
	}
}

func BenchmarkReplaceDigits(b *testing.B) {
	tbl := []byte("sales_2019_07_01_orders")
	for i := 0; i < b.N; i++ {
		replaceDigits(tbl)
	}
}
