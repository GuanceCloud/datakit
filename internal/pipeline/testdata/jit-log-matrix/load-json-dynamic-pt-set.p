doc = load_json(message)
name = doc["payload"]["service"]["name"]
pt_kvs_set(name, doc["payload"]["request"]["status"])
pt_kvs_set(name + "_raw", doc["payload"], false, true)
