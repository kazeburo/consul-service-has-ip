VERSION=0.0.2
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
GO111MODULE=on

all: consul-service-has-ip

.PHONY: consul-service-has-ip

consul-service-has-ip: main.go
	go build $(LDFLAGS) -o consul-service-has-ip

linux: main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o consul-service-has-ip

fmt:
	go fmt ./...

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
	goreleaser --rm-dist
