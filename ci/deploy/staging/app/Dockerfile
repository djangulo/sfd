FROM node:14.11.0-alpine3.11 as node_builder
RUN apk update && apk add make
WORKDIR /sfd
COPY ./sfd/ .
RUN npm install -g yarn
RUN yarn --cwd=./frontend/web install
RUN make web

FROM golang:1.13 as builder
LABEL maintainer="Denis Angulo <djal@tuta.io>"
WORKDIR $GOPATH/src/github.com/djangulo/sfd

# Copy app files from node builder, which includes built semantic-ui
# files
COPY --from=node_builder /sfd .
RUN go get -v ./...
RUN make bindata
#RUN CGO_ENABLED=0 GOOS=linux go build -tags=postgres -p 1 -a -o $GOPATH/bin/sfdapp ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -p 1 -a -o $GOPATH/bin/sfd ./cmd/app

#  wait for postgres connection
RUN git clone https://github.com/eficode/wait-for.git /wait-for


FROM alpine:latest

RUN addgroup -S golang \
    && adduser -S -G golang golang

RUN apk update \
    && apk --no-cache add ca-certificates

COPY --from=builder /go/bin/sfd /sfd
RUN chmod +x /sfd \
    && chown golang:golang /sfd

RUN mkdir -p /site-assets \
    && chown golang:golang /site-assets

COPY --from=builder /wait-for/wait-for /wait-for
RUN sed -i 's/\r//g' /wait-for
RUN chmod +x /wait-for \
    && chown golang:golang /wait-for

USER golang

