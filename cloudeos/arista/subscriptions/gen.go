// Copyright (c) 2022 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package subscriptions

//go:generate protoc -I  $GOPATH/src/terraform-provider-cloudeos/cloudeos/arista --go_out=$GOPATH/src --go-grpc_out=$GOPATH/src subscriptions/subscriptions.proto