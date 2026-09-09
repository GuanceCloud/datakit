stored = ["cached", "value"]
cache_set("jit-local-boundary", stored)
stored[0] = "changed"
loaded = cache_get("jit-local-boundary")
add_key(result, loaded)
add_key(ok, stored[0] == "changed")
