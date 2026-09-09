items = ["seed"]
items = append(items, "one")
items = append(items, "two")
items = append(items, "three")
add_key(result, items)
add_key(ok, len(items) == 4)
