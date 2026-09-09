doc = load_json(message)
request = doc["payload"]["request"]
pt_kvs_set("path", request["path"])
pt_kvs_set("method", request["method"])
pt_kvs_set("duration_ms", request["duration_ms"])
