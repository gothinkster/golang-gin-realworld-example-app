#!/bin/bash

# Check if all Go files are properly formatted
unformatted=$(gofmt -l .)
echo "$unformatted"

if [ -n "$unformatted" ]; then
    echo "There is unformatted code, you should use 'go fmt ./...' to format it."
    echo "Unformatted files:"
    echo "$unformatted"
    exit 1
else
    echo "Codes are formatted."
    exit 0
fi
