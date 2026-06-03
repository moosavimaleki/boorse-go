#!/usr/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

GO_BIN="${GO_BIN:-/home/h-mousavi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.3.linux-amd64/bin/go}"

if [[ ! -x "$GO_BIN" ]]; then
	if command -v go1.26.3 >/dev/null 2>&1; then
		GO_BIN="$(command -v go1.26.3)"
	elif command -v go >/dev/null 2>&1; then
		GO_BIN="$(command -v go)"
	else
		GO_BIN=""
	fi
fi

if [[ -z "$GO_BIN" ]]; then
	echo "Go 1.26.3 was not found. Set GO_BIN to a valid go binary path."
	exit 1
fi

echo "Using: $(env -u GOROOT "$GO_BIN" version)"
env -u GOROOT "$GO_BIN" mod download
mkdir -p "$ROOT_DIR/build"
env -u GOROOT "$GO_BIN" build -o "$ROOT_DIR/build/tsetmc-dev" .
exec "$ROOT_DIR/build/tsetmc-dev"
