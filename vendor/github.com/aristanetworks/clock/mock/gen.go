// Copyright (c) 2018 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

// To generate the mock object, gomock and mockgen need to be installed, by running
//    go get github.com/golang/mock/gomock
//    go get github.com/golang/mock/mockgen
// then run 'go generate' to auto-generate mock.

package mock

//go:generate mockgen -destination=clock.gen.go -package=mock github.com/aristanetworks/clock Clock,Ticker,Timer
