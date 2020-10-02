#!/bin/sh

## Extracts .crt and .key from a traefik-generated acme.json
## Usage
##    extract_acme.sh acme.json resolver_name domain_name

workdir="$(dirname ${0})"

# if ! jq_loc="$(type -p "jq")" || [[ -z $jq_loc ]]; then
#     echo -e "\e[31mERROR\e[0m: jq is required"
#     exit 1
# fi

cat "${1}" \
    | jq -r --arg RESOLVER_NAME ${2} \
        --arg DOMAIN_NAME ${3} '.[$RESOLVER_NAME].Certificates[] | select(.domain.main==$DOMAIN_NAME) | .certificate' \
    | base64 -d \
    >   "${3}.crt"

cat "${1}" \
    | jq -r --arg RESOLVER_NAME ${2} \
        --arg DOMAIN_NAME ${3} '.[$RESOLVER_NAME].Certificates[] | select(.domain.main==$DOMAIN_NAME) | .key' \
    | base64 -d \
    > "${3}.key"
