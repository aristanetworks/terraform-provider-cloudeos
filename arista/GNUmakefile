BINARY=terraform-provider-arista
VERSION=$(shell awk '{ if ($$2=="providerAristaVersion") print $$4 }' version.go | tr -d \")
TEST=./

default: build-all

build-all: linux darwin

linux:
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o $(BINARY)_$(VERSION)_linux_amd64
	GOOS=linux CGO_ENABLED=0 GOARCH=386 go build -o $(BINARY)_$(VERSION)_linux_x86

darwin:
	GOOS=darwin CGO_ENABLED=0 GOARCH=amd64 go build -v -o $(BINARY)_$(VERSION)_darwin_amd64

test:
	go test $(TEST) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v -parallel 20 $(TESTARGS) -timeout 120m

.PHONY: build test testacc
