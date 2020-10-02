#!/bin/sh

rootdir=$(dirname $(dirname $0))

cd $rootdir

go vet ./...
go build -o ./sfdapp -tags=$1 ./cmd/app

cd ./e2e
npm install --only=dev

npx start-server-and-test ../sfdapp :9000 "cypress run --browser $2"

