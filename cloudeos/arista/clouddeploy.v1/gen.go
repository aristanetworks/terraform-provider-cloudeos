// Copyright (c) 2022 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

// Steps for making changes in proto models in clouddeploy.proto
// 1. Amend the original proto and regenerate services.gen.proto in arista.git (via boomtown). The original proto can be found at resources/arista/clouddeploy.v1
// 2. Replace the current proto and the services.gen.proto with the ones generated in step 1 into this repo (Don't forget to correct the license to MPL 2.0)
// 3. Invoke go generate in this repo, which will regenerate all the bindings.

package clouddeploy_v1

//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go
//go:generate go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
//go:generate protoc -I ../.. -I ../../../cloudvision-apis --go_out=../../../.. --go-grpc_out=../../../.. arista/clouddeploy.v1/clouddeploy.proto
//go:generate protoc -I ../.. -I ../../../cloudvision-apis --go_out=../../../.. --go-grpc_out=../../../.. arista/clouddeploy.v1/services.gen.proto
//go:generate goimports -w .
