// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package provider

import (
	"context"
)

// A Provider "owns" some states on a target device streams out notifications on any
// changes to those states.
type Provider interface {

	// Run() kicks off the provider.  This method does not return until ctx
	// is cancelled or an error is encountered,
	// and is thus usually invoked by doing `go provider.Run()'.
	Run(ctx context.Context) error
}
