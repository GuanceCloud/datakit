#!/usr/bin/env bash

patch_elf_rpath() {
  objs=$(filter_elf "$1")
  if [ -z "$objs" ]; then
    return 0
  fi
  for f in $objs; do
    if ! (echo "$f" | grep -q '^/'); then
      f="$(pwd)/$f"
    fi
    if ! patchelf --add-rpath "/app/datakit-profiler/build:/app/async-profiler/build" "$f" >/dev/null; then
      return 1
    fi
  done
  return 0
}

filter_elf() {
  for f in $1; do
      res=$(file -hb "$f")
      if echo "$res" | grep -q '^ELF'; then
        echo "$f"
      fi
  done
  return 0
}
