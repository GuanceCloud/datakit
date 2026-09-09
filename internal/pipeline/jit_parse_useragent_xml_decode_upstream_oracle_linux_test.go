// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Cases are derived from pipeline-go v1.4.3's fn_parse_date_test.go. They use
// explicit calendar fields so the oracle does not depend on two wall-clock
// reads happening on the same day.
func TestJITParseDateUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "full_fields",
			source: `grok(_, "%{INT:year}-%{INT:month}-%{INT:day} %{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:msec}")
parse_date(key="time", yy=year, MM=month, dd=day, hh=hour, mm=min, ss=sec, ms=msec, zone="+8")`,
			input: "2020-12-12 16:40:03.290",
		},
		{
			name: "month_name_and_cst",
			source: `grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", yy=year, MM=month, dd=day, hh=hour, mm=min, ss=sec, zone=tz)`,
			input: "Mon Sep  6 16:40:03 CST 2021",
		},
		{
			name: "partial_year_00",
			source: `grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", yy=year, MM=month, dd=day, hh=hour, mm=min, ss=sec, zone=tz)`,
			input: "Mon Sep  6 16:40:03 CST 00",
		},
		{
			name: "partial_year_09",
			source: `grok(_, "%{NOTSPACE:wd}\\s+%{NOTSPACE:month}\\s+%{INT:day}\\s+%{INT:hour}:%{INT:min}:%{INT:sec}\\s+%{NOTSPACE:tz}\\s+%{INT:year}")
parse_date(key="time", yy=year, MM=month, dd=day, hh=hour, mm=min, ss=sec, zone=tz)`,
			input: "Mon Sep  6 16:40:03 CST 09",
		},
		{
			name: "nanoseconds",
			source: `grok(_, "%{INT:year}-%{INT:month}-%{INT:day} %{INT:hour}:%{INT:min}:%{INT:sec}\\.%{INT:ns}")
parse_date(key="time", yy=year, MM=month, dd=day, hh=hour, mm=min, ss=sec, ns=ns, zone="UTC")`,
			input: "2020-12-12 16:40:03.290290330",
		},
		{
			name: "invalid_hour",
			source: `grok(_, "%{INT:year}-%{INT:month}-%{INT:day} %{INT:hour}:%{INT:min}:%{INT:sec}")
parse_date(key="time", yy=year, MM=month, dd=day, hh=hour, mm=min, ss=sec, zone="UTC")`,
			input: "2020-12-12 25:40:03", fails: true,
		},
		{
			name: "invalid_minute",
			source: `grok(_, "%{INT:year}-%{INT:month}-%{INT:day} %{INT:hour}:%{INT:min}:%{INT:sec}")
parse_date(key="time", yy=year, MM=month, dd=day, hh=hour, mm=min, ss=sec, zone="UTC")`,
			input: "2020-12-12 12:61:03", fails: true,
		},
	})
}

// Cases are derived from pipeline-go's fn_useragent_test.go.
func TestJITUserAgentUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name:   "chrome_windows",
			source: `add_key(agent,message); user_agent(agent)`,
			input:  "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
		},
		{
			name:   "safari_macos",
			source: `add_key(agent,message); user_agent(agent)`,
			input:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15",
		},
		{
			name:   "constant_prefix",
			source: `add_key(agent,message); add_key(os,"existing"); user_agent(agent,"ua_")`,
			input:  "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
		},
		{
			name:   "dynamic_prefix",
			source: `add_key(agent,message); p="ua_"; user_agent(agent,p)`,
			input:  "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
		},
		{
			name:   "missing_key",
			source: `user_agent(agent,"ua_")`,
			input:  "ignored",
		},
	})
}

// Cases are derived from pipeline-go's fn_xml_test.go, including malformed
// XML and XPath behavior (which is a successful no-op there, not a terminal).
func TestJITXMLUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"text", `xml(_,"/entry/fieldx/text()",out)`, `<entry><fieldx>valuex</fieldx></entry>`, "out", "valuex", false},
		{"array", `xml(_,"/entry/fieldarray//fielda[1]/text()",out)`, `<entry><fieldarray><fielda>one</fielda><fielda>two</fielda></fieldarray></entry>`, "out", "one", false},
		{"attribute", `xml(_,"/OrderEvent/@actionCode",out)`, `<OrderEvent actionCode="5"><OrderNumber>ORD12345</OrderNumber></OrderEvent>`, "out", "5", false},
		{"mixed_content", `xml(_,"/root/item",out)`, `<root><item>before<b>bold</b>after<i>italic</i>end</item></root>`, "out", "beforeboldafteritalicend", false},
		{"prefixed_namespace", `xml(_,"/r:root/r:item/text()",out)`, `<r:root xmlns:r="urn:test"><r:item>namespaced</r:item></r:root>`, "out", "namespaced", false},
		{"default_namespace_unqualified_xpath", `xml(_,"/root/item/text()",out)`, `<root xmlns="urn:test"><item>namespaced</item></root>`, "", nil, false},
		{"default_namespace_with_cdata_and_xmlns_text", `xml(_,"/root/item/text()",out)`, `<root xmlns="urn:test" note=" xmlns='not-a-declaration'"><item><![CDATA[<fake xmlns="urn:fake">text</fake>]]></item></root>`, "out", `<fake xmlns="urn:fake">text</fake>`, false},
		{"scalar_count_is_noop", `xml(_,"count(/root/item)",out)`, `<root><item>one</item><item>two</item></root>`, "", nil, false},
		{"scalar_string_is_noop", `xml(_,"string(/root/item)",out)`, `<root><item>one</item></root>`, "", nil, false},
		{"invalid_xml", `xml(_,"/OrderEvent/VendorNumber/text()",out)`, `Not a valid XML`, "", nil, false},
		{"invalid_xpath", `xml(_,"invalid xpath expr",out)`, `<OrderEvent><VendorNumber>V11111</VendorNumber></OrderEvent>`, "", nil, false},
	})
}

// The byte strings are generated with the same x/text GBK encoder used by
// pipeline-go's fn_decoder_test.go, then exercised through the raw ABI.
func TestJITDecodeGBKUpstreamOracle(t *testing.T) {
	texts := []string{"测试一下", "不知道", "测试一下123456", "哈哈哈哈哈", "-汪98阿萨德离开家"}
	cases := make([]jsonOracleCase, 0, len(texts))
	for _, text := range texts {
		encoded, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(text))
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, jsonOracleCase{
			name: text, source: `decode(_,"gbk")`, input: string(encoded), key: "message", want: text,
		})
	}
	runJSONOracle(t, cases)
}
