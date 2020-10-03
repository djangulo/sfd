#!/bin/sh

GOPATH=$(pwd)/go
rootdir=$(dirname $(dirname $0))

cd $rootdir

go vet ./...
go build -o $rootdir/sfdapp -tags=$1 ./cmd/app
cp $rootdir/config/config.conf /tmp
SFD_CONFIG_FILE=/tmp/config.conf go test -tags=$1 -v ./...

