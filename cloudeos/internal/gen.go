// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package clouddeploy_v1

//go:generate go run $GOPATH/src/arista/resources/boomtown/cmd/boomtown -I . -I $GOPATH/src/arista/resources -f clouddeploy.proto
//go:generate protoc -I $GOPATH/src/arista/resources --go_out=plugins=grpc:$GOPATH/src arista/clouddeploy.v1/clouddeploy.proto
//go:generate protoc -I $GOPATH/src/arista/resources --go_out=plugins=grpc:$GOPATH/src arista/clouddeploy.v1/services.gen.proto
//go:generate goimports -w .
