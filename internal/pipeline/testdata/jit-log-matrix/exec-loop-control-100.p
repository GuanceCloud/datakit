x = seed
for i = 0; i < 100; i += 1 { if i < 20 { continue }; if i > 79 { break }; x += i }
add_key(result, x)
