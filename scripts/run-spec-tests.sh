#!/usr/bin/env bash
# Runs the RealWorld API spec test suite against this backend.
# Specs live in https://github.com/realworld-apps/realworld under specs/api;
# the Hurl suite is the source of truth, the Bruno collection is generated from it.
#
# Usage: bash ./scripts/run-spec-tests.sh <hurl|bruno>
# Requires: go, curl, plus hurl (https://hurl.dev) for the hurl flavor,
# or node/npx/bun (or a globally installed @usebruno/cli) for the bruno flavor.
set -euo pipefail

FLAVOR="${1:-}"
case "$FLAVOR" in
    hurl|bruno) ;;
    *) echo "usage: bash ./scripts/run-spec-tests.sh <hurl|bruno>" >&2; exit 2 ;;
esac

PORT="${PORT:-8080}"
HOST="${HOST:-http://localhost:$PORT}"
BRUNO_SANDBOX="${BRUNO_SANDBOX:-safe}"

# Pinned spec version (commit in realworld-apps/realworld) and the sha256 of
# its source tarball. Bump both together when adopting a newer spec.
# Override SPECS_REF (branch, tag or commit) to test another version; checksum
# verification is skipped unless a matching SPECS_SHA256 is provided.
SPECS_REF_DEFAULT="98f29fb3f8bcb1dd614b91f2851371bf22c34775"
SPECS_SHA256_DEFAULT="faf81a1eb15fea34705d111e2711d181a505d08f3a92fd321b14b6965d4c5260"
SPECS_REF="${SPECS_REF:-$SPECS_REF_DEFAULT}"
if [ -z "${SPECS_SHA256+x}" ]; then
    if [ "$SPECS_REF" = "$SPECS_REF_DEFAULT" ]; then
        SPECS_SHA256="$SPECS_SHA256_DEFAULT"
    else
        SPECS_SHA256=""
    fi
fi

SPECS_DIR=./tmp/realworld-specs
SPECS_TARBALL=./tmp/realworld-specs.tar.gz

mkdir -p ./tmp ./data

echo "Downloading RealWorld API specs (realworld-apps/realworld@$SPECS_REF)..."
rm -rf "$SPECS_DIR" "$SPECS_TARBALL"
mkdir -p "$SPECS_DIR"
curl -sSL -o "$SPECS_TARBALL" "https://github.com/realworld-apps/realworld/archive/$SPECS_REF.tar.gz"

if [ -n "$SPECS_SHA256" ]; then
    if ! echo "$SPECS_SHA256  $SPECS_TARBALL" | sha256sum -c -; then
        echo "error: specs tarball checksum mismatch for $SPECS_REF." >&2
        echo "The upstream archive changed (or was tampered with)." >&2
        echo "Inspect realworld-apps/realworld, then update SPECS_REF/SPECS_SHA256 in this script." >&2
        exit 1
    fi
else
    echo "warning: SPECS_SHA256 not set for ref $SPECS_REF - skipping checksum verification"
fi

tar xzf "$SPECS_TARBALL" -C "$SPECS_DIR" --strip-components=3 "realworld-$SPECS_REF/specs/api"

# Clean up database
rm -f ./data/gorm.db

echo "Building application..."
go build -o ./tmp/app hello.go

echo "Starting server..."
PORT="$PORT" GIN_MODE=release ./tmp/app &
SERVER_PID=$!

cleanup() {
    echo "Stopping server..."
    kill "$SERVER_PID" 2>/dev/null || true
}
trap cleanup EXIT

echo "Waiting for server to be ready..."
for _ in {1..30}; do
    if curl -s "$HOST/api/ping" > /dev/null; then
        echo "Server is up!"
        break
    fi
    sleep 1
done

case "$FLAVOR" in
hurl)
    echo "Running Hurl API spec tests against $HOST..."
    hurl --test \
      --jobs 1 \
      --variable "host=$HOST" \
      --variable "uid=$(date +%s)$$" \
      "$SPECS_DIR"/hurl/*.hurl
    ;;
bruno)
    if command -v bru &> /dev/null; then
        BRU=(bru)
    elif command -v npx &> /dev/null; then
        BRU=(npx --yes @usebruno/cli)
    elif command -v bun &> /dev/null; then
        BRU=(bun x @usebruno/cli)
    else
        echo "error: need bru, npx or bun to run the Bruno CLI" >&2
        exit 1
    fi

    cd "$SPECS_DIR/bruno"
    for entry in */; do
        folder="${entry%/}"
        [ "$folder" = "environments" ] && continue
        echo ""
        echo "--- ${BRU[*]} run $folder ---"
        "${BRU[@]}" run "$folder" --env local --env-var "host=$HOST" --sandbox "$BRUNO_SANDBOX"
    done
    ;;
esac
