original = seed
original += suffix
updated = original
updated += suffix
add_key(original, original)
add_key(updated, updated)
add_key(ok, len(original) == expected_len && len(updated) == expected_len + len(suffix))
