x = seed
for i = 0; i < 100; i += 1 { if i < 50 { x += i } else { x -= i } }
add_key(result, x)
