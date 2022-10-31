#!/bin/bash

docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.17-alpine go build -v -ldflags "-w -s" -o app
