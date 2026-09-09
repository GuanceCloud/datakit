base = ["a", "b", "c"]
view = base
grown = append(view, "d")
grown = append(grown, "e")
base[0] = "changed"
add_key(base_view, base)
add_key(view_view, view)
add_key(grown_view, grown)
base[1] = "changed-after-output"
add_key(ok, len(grown) == 5 && grown[3] == "d" && grown[4] == "e" && view[0] == "changed" && view[1] == "changed-after-output")
