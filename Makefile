
.PHONY: bindata
# Generate all the assets.go files using go-bindata.
bindata:
	go get -u github.com/go-bindata/go-bindata/...
	go-bindata -pkg accounts -o accounts/assets.go accounts/templates/...
	# go-bindata -pkg admin -o admin/assets.go templates/... admin/templates/... accounts/templates/...
	go-bindata -fs -prefix "frontend/web/build" -o ./cmd/app/assets.go \
		frontend/web/build/...

.PHONY: web
web:
	cd ./frontend/web
	yarn --version --cwd=./frontend/web
	cd ./../../

assets: web bindata

.PHONY: tls
tls:
	openssl genrsa -out /tmp/https-server.key 2048
	openssl ecparam -genkey -name secp384r1 -out /tmp/https-server.key
	openssl req -new -x509 -sha256 -key /tmp/https-server.key -out /tmp/https-server.crt -days 3650 -subj "/C=DO/ST=D.N./L=StoDgo/O=RD/CN=localhost"

.PHONY: bin
bin:
	go build -ldflags="-w" -o=./sfd ./cmd/app/...

build: web bindata bin