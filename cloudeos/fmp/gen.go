// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package fmp

//go:generate protoc -I  $GOPATH/src/cloudeos-terraform-provider/cloudeos/ --go_out=$GOPATH/src --go-grpc_out=$GOPATH/src fmp/deletes.proto
//go:generate protoc -I  $GOPATH/src/cloudeos-terraform-provider/cloudeos/ --go_out=$GOPATH/src --go-grpc_out=$GOPATH/src fmp/extensions.proto
//go:generate protoc -I  $GOPATH/src/cloudeos-terraform-provider/cloudeos/ --go_out=$GOPATH/src --go-grpc_out=$GOPATH/src fmp/inet.proto
//go:generate protoc -I  $GOPATH/src/cloudeos-terraform-provider/cloudeos/ --go_out=$GOPATH/src --go-grpc_out=$GOPATH/src fmp/wrappers.proto
