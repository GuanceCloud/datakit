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

func TestFilter(t *testing.T) {
	// Test cases grouped by scenario
	testCases := []struct {
		name    string
		include []string
		exclude []string
		target  string
		want    bool
		wantErr bool // Whether NewFilter should return error
	}{
		// Basic include/exclude logic
		{
			name:    "matched by exclude pattern",
			include: []string{".*"},
			exclude: []string{"model"},
			target:  "model",
			want:    false,
			wantErr: false,
		},
		{
			name:    "matched by include and not excluded",
			include: []string{"user_.*"},
			exclude: []string{"test"},
			target:  "user_db",
			want:    true,
			wantErr: false,
		},
		{
			name:    "not matched by any include",
			include: []string{"prod_.*"},
			exclude: []string{},
			target:  "test_db",
			want:    false,
			wantErr: false,
		},
		{
			name:    "matched by one of multiple excludes",
			include: []string{".*"},
			exclude: []string{"msdb", "rdsadmin"},
			target:  "rdsadmin",
			want:    false,
			wantErr: false,
		},
		{
			name:    "matched by one of multiple includes",
			include: []string{"a.*", "b.*"},
			exclude: []string{},
			target:  "b_test",
			want:    true,
			wantErr: false,
		},

		// Empty include scenarios
		{
			name:    "empty include and not excluded",
			include: []string{},
			exclude: []string{"model"},
			target:  "user_db",
			want:    true,
			wantErr: false,
		},
		{
			name:    "empty include but excluded",
			include: []string{},
			exclude: []string{"model"},
			target:  "model",
			want:    false,
			wantErr: false,
		},
		{
			name:    "both include and exclude empty",
			include: []string{},
			exclude: []string{},
			target:  "any_db",
			want:    true,
			wantErr: false,
		},

		// Regex pattern matching details
		{
			name:    "partial match in exclude",
			include: []string{".*"},
			exclude: []string{"model"},
			target:  "model_test",
			want:    false,
			wantErr: false,
		},
		{
			name:    "exact match with anchors",
			include: []string{".*"},
			exclude: []string{"^model$"},
			target:  "model_test",
			want:    true,
			wantErr: false,
		},
		{
			name:    "wildcard pattern match",
			include: []string{"db_\\d+"},
			exclude: []string{},
			target:  "db_123",
			want:    true,
			wantErr: false,
		},
		{
			name:    "wildcard pattern no match",
			include: []string{"db_\\d+"},
			exclude: []string{},
			target:  "db_abc",
			want:    false,
			wantErr: false,
		},

		// Invalid regex patterns
		{
			name:    "invalid include regex",
			include: []string{"[a-z"}, // Unclosed bracket
			exclude: []string{},
			target:  "",
			want:    false,
			wantErr: true,
		},
		{
			name:    "invalid exclude regex",
			include: []string{".*"},
			exclude: []string{"(abc"}, // Unclosed bracket
			target:  "",
			want:    false,
			wantErr: true,
		},
		{
			name:    "valid regex patterns",
			include: []string{"a.*b"},
			exclude: []string{"^\\d+"},
			target:  "a123b",
			want:    true,
			wantErr: false,
		},
	}

	// Execute all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create filter configuration
			config := FilterConfig{
				Include: tc.include,
				Exclude: tc.exclude,
			}

			// Initialize filter and check for errors
			filter, err := NewFilter(config)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NewFilter() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return // No need to check Allow() if initialization failed
			}

			// Verify Allow() result
			if got := filter.Allow(tc.target); got != tc.want {
				t.Errorf("Allow(%q) = %v, want %v", tc.target, got, tc.want)
			}
		})
	}
}
