#!/bin/sh

rootdir=$(dirname $(dirname $0))

cd $rootdir

#go vet ./...
#go build -o ./sfdapp -tags=$1 ./cmd/app

cd ./e2e
npm install --only=dev

CYPRESS_BASE_URL=$1 \
    CYPRESS_BASICAUTH_USER=$2 \
    CYPRESS_BASICAUTH_PASS=$3 \
    npx cypress run --browser $4

