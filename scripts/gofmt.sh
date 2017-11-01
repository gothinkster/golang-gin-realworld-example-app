#!/bin/bash

gofmt=$(govendor fmt +l)
echo $gofmt

if [ ${#gofmt} != 0 ]; then
    echo "There is unformatted code, you should use `go fmt ./\.\.\.` to format it."
    exit 1
else
    echo "Codes are formatted."
    exit 0
fi
