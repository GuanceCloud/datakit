#!/usr/bin/env bash
# Used by package jobs and native tests. Never expose the deploy token in shell tracing.
set +x
set -euo pipefail
case "${PIPELINE_JIT_ENABLED:-0}" in
  0|false|off|no|'') exit 0 ;;
  1|true|on|yes) ;;
  *) echo 'Invalid PIPELINE_JIT_ENABLED' >&2; exit 1 ;;
esac
: "${PP_JIT_READ_TOKEN:?configure the DK CI variable PP_JIT_READ_TOKEN}"
: "${PP_JIT_VERSION:?select a published JIT version}"
: "${PLATYPUS_JIT_RUNTIME_DIR:?}"
[[ "$PP_JIT_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Expected a production vX.Y.Z version' >&2; exit 1; }
base="https://gitlab.jiagouyun.com/api/v4/projects/guance%2Fplatplusplus/packages/generic/platypus-jit/$PP_JIT_VERSION"
mkdir -p "$PLATYPUS_JIT_RUNTIME_DIR"
staging=$(mktemp -d "$PLATYPUS_JIT_RUNTIME_DIR/.download.XXXXXX")
trap 'rm -rf -- "$staging"' EXIT
for arch in amd64 arm64; do
  archive="platypus-jit-linux-$arch.tar.gz"
  for suffix in '' .sha256; do
    status=$(printf 'DEPLOY-TOKEN: %s\n' "$PP_JIT_READ_TOKEN" |
      curl --header @- --fail --silent --show-error --retry 3 \
        --connect-timeout 15 --max-time 180 --proto '=https' \
        --output "$staging/$archive$suffix" --write-out '%{http_code}' \
        "$base/$archive$suffix")
    [[ "$status" == 200 ]] || { echo "JIT download failed: $archive$suffix HTTP $status" >&2; exit 1; }
  done
  # Validate the checksum's filename too; never let it select another local file.
  read -r digest filename < "$staging/$archive.sha256"
  [[ "$digest" =~ ^[0-9a-f]{64}$ && "$filename" == "$archive" ]] || { echo 'Invalid archive checksum file' >&2; exit 1; }
  (cd "$staging" && printf '%s  %s\n' "$digest" "$archive" | sha256sum --check --strict)
  # The publisher emits exactly four regular files, without links or directories.
  tar -tzf "$staging/$archive" > "$staging/members"
  printf '%s\n' "linux-$arch/glibc-ceiling.txt" "linux-$arch/libplatypus_jit.so" \
    "linux-$arch/libplatypus_jit.so.sha256" "linux-$arch/manifest.json" | sort > "$staging/expected"
  sort "$staging/members" > "$staging/actual"
  cmp "$staging/expected" "$staging/actual"
  tar -tvzf "$staging/$archive" > "$staging/types"
  if grep -qv '^-' "$staging/types"; then echo 'Unexpected non-regular archive member' >&2; exit 1; fi
  tar -xzf "$staging/$archive" --no-same-owner --no-same-permissions -C "$staging"
done
# Parse JSON rather than extracting text; only a validated SHA is passed to CI.
revision=$(python3 - "$staging" <<'PY_MANIFEST'
import json
from pathlib import Path
import re
import sys
root = Path(sys.argv[1])
revisions = []
for arch in ('amd64', 'arm64'):
    manifest = json.loads((root / f'linux-{arch}' / 'manifest.json').read_text())
    revision = manifest.get('source_revision')
    if not isinstance(revision, str) or not re.fullmatch(r'[0-9a-f]{40}', revision):
        sys.exit(f'{arch}: invalid manifest source_revision')
    revisions.append(revision)
if revisions[0] != revisions[1]:
    sys.exit('JIT architecture manifests have different source revisions')
print(revisions[0])
PY_MANIFEST
)
# Both archives must pass before replacing prior files. The DK packager verifies
# the manifests, inner checksums, ELF architecture, ABI and expected source SHA.
for arch in amd64 arm64; do
  test ! -L "$PLATYPUS_JIT_RUNTIME_DIR/linux-$arch"
  mkdir -p "$PLATYPUS_JIT_RUNTIME_DIR/linux-$arch"
  cp --remove-destination "$staging/linux-$arch/"* "$PLATYPUS_JIT_RUNTIME_DIR/linux-$arch/"
done
printf '%s\n' "$revision" > "$staging/source-revision"
mv -f "$staging/source-revision" "$PLATYPUS_JIT_RUNTIME_DIR/source-revision"
echo "Prepared Pipeline JIT $PP_JIT_VERSION (amd64 + arm64)"
