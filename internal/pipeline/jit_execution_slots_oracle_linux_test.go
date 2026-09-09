// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
)

// These cases keep point reads/writes visible beside locals so slot allocation
// cannot accidentally turn a local binding into a committed point update.
func TestJITExecutionSlotsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "native_loop_break_continue",
			source: `x = value
for i = 0; i < 10; i += 1 {
    if i == 2 { continue }
    if i > 4 { break }
    x += i
}
add_key(result, x)`,
			key: "result", want: int64(15),
		},
		{
			name: "native_loop_existing_index",
			source: `i = 99
x = value
for i = 0; i < 5; i += 1 { x += i }
add_key(result, i)`,
			key: "result", want: int64(5),
		},
		{
			name: "native_loop_zero_iterations",
			source: `x = value
for i = 0; i < 0; i += 1 { x += i }
add_key(result, x)`,
			key: "result", want: int64(7),
		},
		{
			name: "native_loop_readonly_point",
			source: `x = 0
for i = 0; i < value; i += 1 { x += i }
add_key(result, x)`,
			key: "result", want: int64(21),
		},
		{
			name: "native_loop_nested_branch",
			source: `x = value
for i = 0; i < 8; i += 1 {
    if i > 1 { if i == 3 { break } else { continue } }
    x += i
}
add_key(result, x)`,
			key: "result", want: int64(8),
		},
		{
			name: "native_loop_float_guard",
			source: `x = 0.5
for i = 0; i < 5; i += 1 { x += i }
add_key(result, x)`,
			key: "result", want: float64(10.5),
		},

		{
			name: "region_self_assignment_starts_from_point",
			source: `value = value + 1
x = value * 2
add_key(result, x)
add_key(original, pt_kvs_get("value"))`,
			key: "result", want: int64(16),
		},
		{
			name: "region_nil_local_does_not_read_point",
			source: `value = nil
x = 3
value = value + 1
add_key(after, true)`,
			key: "value", want: int64(7), fails: true,
		},
		{
			name: "region_reload_point_after_write",
			source: `a = value + 1
b = a + 1
add_key(value, 100)
c = value + 1
d = c + 1
add_key(result, d)`,
			key: "result", want: int64(102),
		},
		{
			name: "region_local_survives_point_write",
			source: `value = 2
x = value + 1
add_key(value, 100)
y = value + 1
z = y + 1
add_key(result, z)`,
			key: "result", want: int64(4),
		},
		{
			name: "region_unexecuted_binding_reads_point",
			source: `if false { value = 100; temporary = 1 }
a = value + 1
b = a + 1
add_key(result, b)`,
			key: "result", want: int64(9),
		},
		{
			name: "region_expired_binding_reads_point",
			source: `if true { value = 100; temporary = value + 1 }
a = value + 1
b = a + 1
add_key(result, b)`,
			key: "result", want: int64(9),
		},
		{
			name: "native_integer_wrap_each_operation",
			source: `x = 9223372036854775807
y = (x + 1) * 2 - x
add_key(result, y)`,
			key: "result", want: int64(-9223372036854775807),
		},
		{
			name: "native_integer_subtree_before_float_promotion",
			source: `x = 0.5
y = (9223372036854775807 + 1) + x - 0
add_key(result, y)`,
			key: "result", want: float64(-9223372036854775808),
		},
		{
			name: "native_float_comparison",
			source: `x = 3.5
y = (x * 1 - 0) >= 3.5
add_key(result, y)`,
			key: "result", want: true,
		},
		{
			name: "native_guard_observes_type_change",
			source: `x = 2
for i = 0; i < 2; i += 1 {
 if i == 1 { x = "ab" }
 y = x + x + x + x
 add_key(result, y)
}`,
			key: "result", want: "abababab",
		},
		{
			name: "native_then_error_keeps_committed_prefix",
			source: `x = 4
y = (x + 3) * 2 - x
add_key(before, y)
x = nil
y = (x + 3) * 2 - x
add_key(after, y)`,
			key: "before", want: int64(10), fails: true,
		},
		{
			name: "branch_rebind_visible_after_block",
			source: `value = 1
if true { value = value + 4 }
add_key(result, value)
add_key(after_branch, value + 1)`,
			key: "result", want: int64(5),
		},
		{
			name: "local_shadows_point_until_explicit_read",
			source: `value = 9
point_value = pt_kvs_get("value")
add_key(local_value, value)
			add_key(point_value, point_value)
			add_key(result, value + point_value)`,
			key: "result", want: int64(16),
		},
		{
			name: "loop_continue_break_preserves_outer_slot",
			source: `total = 10
for i = 0; i < 6; i += 1 {
 if i == 1 { continue }
 if i == 4 { break }
 total += i
}
			add_key(result, total)
			add_key(loop_last, i)`,
			key: "result", want: int64(15),
		},
		{
			name: "named_builtin_then_local_arithmetic",
			source: `payload = message
path = "value"
gjson(input=payload, json_path=path, key_name="decoded")
value = pt_kvs_get("decoded")
value += 2
add_key(result, value)`,
			input: `{"value":40}`, key: "result", want: float64(42),
		},
		{
			name: "missing_compound_does_not_create_target",
			source: `missing += 1
add_key(observed, missing)
add_key(done, true)`,
			key: "observed", want: nil,
		},
		{
			name: "explicit_nil_compound_keeps_nil_target",
			source: `value = nil
value += 1
add_key(observed, value)`,
			key: "observed", want: nil, fails: true,
		},
		{
			name: "rhs_error_keeps_prefix_without_target_write",
			source: `add_key(before, "retained")
add_key(changed, 7)
target += 1 / divisor
add_key(after, true)`,
			key: "target", want: nil, fails: true,
		},
		{
			name: "load_json_nested_alias_read_modify_consistency",
			source: `data = load_json(message)
alias = data["outer"]
before = alias["value"]
alias["value"] = alias["value"] + 2
after = alias["value"]
original = data["outer"]["value"]
add_key(result, before + after + original)`,
			input: `{"outer":{"value":3}}`, key: "result", want: float64(13),
		},
	}, func(pt *point.Point) {
		pt.AddKVs(point.NewKV("value", int64(7)))
	})
}
