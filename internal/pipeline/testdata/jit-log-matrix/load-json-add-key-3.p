doc = load_json(message)
request = doc["payload"]["request"]
add_key(path, request["path"])
add_key(method, request["method"])
add_key(duration_ms, request["duration_ms"])
