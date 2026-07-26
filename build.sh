#!/bin/sh
CGO_ENABLED=0 GOAMD64=v3 go build -trimpath -buildvcs=false -ldflags="-s -w" -o ./bin/newsfeed