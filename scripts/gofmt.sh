#!/bin/bash

if [ -n "$(go fmt ./...)" ]; then
    echo "Go code is not formatted:"
    gofmt -d .
    exit 1
fi
