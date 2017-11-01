#!/usr/bin/env bash

set -e

# http://stackoverflow.com/a/21142256/2055281

echo "mode: atomic" > coverage.txt

for d in $(find ./* -maxdepth 10 -type d); do
    if ls $d/*.go &> /dev/null; then
        go test  -coverprofile=profile.out -covermode=atomic $d
        if [ -f profile.out ]; then
            echo "$(pwd)"
            cat profile.out | grep -v "mode: " >> coverage.txt
            rm profile.out
        fi
    fi
done

