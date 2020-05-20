// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package gen

// To run this command you need protoc:
// brew install protobuf

//go:generate protoc --proto_path=${GOPATH}/src --go_out=plugins=grpc,:${GOPATH}/src github.com/aristanetworks/cloudvision-go/device/inventory.proto
