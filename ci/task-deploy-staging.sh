#!/usr/bin/env bash

set -euxo pipefail

cd $WORKDIR
mv $WORKDIR/services/sfd-app/staging-ci.yml $WORKDIR/services/sfd-app/docker-compose.yml

chmod +x $WORKDIR/services/sfd-app/ci/lib/extract_acme.sh
sudo $WORKDIR/services/sfd-app/ci/lib/extract_acme.sh \
    $WORKDIR/letsencrypt/acme.json leresolver $APPHOST

sudo chown $USER:$USER "${APPHOST}.crt"
sudo chown $USER:$USER "${APPHOST}.key"

mv "${APPHOST}.crt" $WORKDIR/services/sfd-app/
mv "${APPHOST}.key" $WORKDIR/services/sfd-app/

mkdir -p $WORKDIR/sfd-app/htpasswd.d
mkdir -p $WORKDIR/sfd-app/assets

./compose.sh up --detach --remove-orphans --build --force-recreate

