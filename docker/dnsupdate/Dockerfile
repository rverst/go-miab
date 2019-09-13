######
# Build a tiny image for the binary, inspired by https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
# 1. build executable binary
######
FROM golang:alpine AS builder

# install dependencies
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# create unprivileged user
RUN adduser -D -g '' appuser

WORKDIR $GOPATH/src/github.com/rverst/go-miab/
COPY . .

ENV GO111MODULE=on

# fetch dependencies
RUN go mod download
RUN go mod verify

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/dnsupdate ./cmd/dnsupdate/main.go

# 2. build small image from googles distroless
# https://github.com/GoogleContainerTools/distroless
# FROM gcr.io/distroless/static
# I think this is not working because some dependency needs some lib
# not provided in the distroless images :(
FROM alpine

# import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# copy binary
COPY --from=builder /go/bin/dnsupdate /

# use the unprivileged user
USER appuser

# set some default environment variables
ENV DNS_INTERVAL=1800
ENV DNS_A=true
ENV DNS_AAAA=false

ENV DNS_USER="admin@example.org"
ENV DNS_PASSWORD="secret"
ENV DNS_ENDPOINT="https://box.example.org"
ENV DNS_DOMAINS="sub1.example.org, *.example.org"

ENTRYPOINT ["/dnsupdate"]
