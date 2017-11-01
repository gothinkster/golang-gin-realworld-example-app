#!/bin/bash

if [[ -n "$(go fmt ./...)" ]]; then
    echo "There is unformatted code, you should use `go fmt ./...` to format it"
    exit 1
fi
