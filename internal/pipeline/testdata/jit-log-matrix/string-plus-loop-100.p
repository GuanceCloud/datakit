s = seed
for i = 0; i < 100; i += 1 { s += suffix }
n = len(s)
add_key(ok, n == expected_len)
