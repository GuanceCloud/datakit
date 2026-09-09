values = {"state":"new", "owner":"edge"}
values["state"] = "ready"
values["count"] = "three"
add_key(ok, values["state"] == "ready" && values["count"] == "three")
