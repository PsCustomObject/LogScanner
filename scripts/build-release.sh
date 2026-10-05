#!/bin/sh
# Build dependency-free release binaries from the repository root.
set -eu

version=${1:-0.1.0-dev}
case "$version" in
  *[!a-zA-Z0-9._-]*|'') echo 'Invalid version' >&2; exit 2 ;;
esac

mkdir -p dist
for platform in windows linux darwin; do
  for arch in amd64 arm64; do
    suffix=''
    label=$platform
    if [ "$platform" = windows ]; then suffix='.exe'; fi
    if [ "$platform" = darwin ]; then label=macos; fi
    asset="logscanner-$label-$arch$suffix"
    echo "Building $asset"
    CGO_ENABLED=0 GOOS="$platform" GOARCH="$arch" \
      go build -trimpath -ldflags "-s -w -X main.version=$version" \
      -o "dist/$asset" ./cmd/logscanner
  done
done

# Hash only this release's six binaries, not unrelated files in dist.
python3 - <<'PY'
from pathlib import Path
import hashlib

assets = []
for platform in ('windows', 'linux', 'macos'):
    for arch in ('amd64', 'arm64'):
        suffix = '.exe' if platform == 'windows' else ''
        assets.append(Path('dist') / f'logscanner-{platform}-{arch}{suffix}')
lines = [f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.name}\n' for p in assets]
Path('dist/SHA256SUMS.txt').write_text(''.join(lines), encoding='utf-8')
PY
