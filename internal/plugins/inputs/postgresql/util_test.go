// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import "testing"

// Test case structure
type testCase struct {
	name     string
	query    string
	expected string
}

// Test data set
var testCases = []testCase{
	{
		name:     "SELECT query without SET statements",
		query:    "SELECT * FROM pg_settings WHERE name = $1",
		expected: "SELECT * FROM pg_settings WHERE name = $1",
	},
	{
		name:     "Multiple statements without SET",
		query:    "SELECT * FROM pg_settings; DELETE FROM pg_settings;",
		expected: "SELECT * FROM pg_settings; DELETE FROM pg_settings;",
	},
	{
		name:     "Simple SET command with SELECT",
		query:    "SET search_path TO 'my_schema', public; SELECT * FROM pg_settings",
		expected: "SELECT * FROM pg_settings",
	},
	{
		name:     "SET TIME ZONE command with SELECT",
		query:    "SET TIME ZONE 'Europe/Rome'; SELECT * FROM pg_settings",
		expected: "SELECT * FROM pg_settings",
	},
	{
		name:     "Multiple SET LOCAL commands with SELECT",
		query:    "SET LOCAL request_id = 1234; SET LOCAL hostname TO 'Bob''s Laptop'; SELECT * FROM pg_settings",
		expected: "SELECT * FROM pg_settings",
	},
	{
		name:     "Large number of repeated SET statements",
		query:    repeat("SET LONG;", 1024) + "SELECT *;",
		expected: "SELECT *;",
	},
	{
		name:     "SET statement with long string",
		query:    "SET " + repeat("'quotable'", 1024) + "; SELECT *;",
		expected: "SELECT *;",
	},
	{
		name:     "SET statement with extremely long string",
		query:    "SET 'l" + repeat("o", 1024) + "ng'; SELECT *;",
		expected: "SELECT *;",
	},
	{
		name:     "SET statement with comment",
		query:    " /** pl/pgsql **/ SET 'comment'; SELECT *;",
		expected: "SELECT *;",
	},
	{
		name:     "SET statement in the middle",
		query:    "SELECT 1; SET a=1; SELECT 2;",
		expected: "SELECT 1; SET a=1; SELECT 2;",
	},
	{
		name:     "Non-SQL string",
		query:    "this isn't SQL",
		expected: "this isn't SQL",
	},
	{
		name:     "Empty string",
		query:    "",
		expected: "",
	},
}

// Helper function: repeat string n times (performance optimized)
func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	// Pre-allocate memory for better performance
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

// Test function
func TestTrimLeadingSetStmts(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := TrimLeadingSetStmts(tc.query)
			if actual != tc.expected {
				t.Errorf("Test case %q failed\nInput: %q\nActual output: %q\nExpected output: %q",
					tc.name, tc.query, actual, tc.expected)
			}
		})
	}
}
