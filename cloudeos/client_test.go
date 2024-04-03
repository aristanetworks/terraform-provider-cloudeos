// Copyright (c) 2024 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"fmt"
	"os"
	"testing"
)

func ctProvider(t *testing.T) *CloudeosProvider {
	p := &CloudeosProvider{
		// SA_TOKEN env variable should be set to a valid cvaas service
		// account token.
		srvcAcctToken: os.Getenv("SA_TOKEN"),
		server:        "www.cv-dev.corp.arista.io",
		cvaasDomain:   "",
	}
	if p.srvcAcctToken == "" {
		fmt.Fprintln(os.Stderr, "warning: no client tests can run, SA_TOKEN "+
			"env variable is not set to a service account token")
		t.Skip()
	}
	return p
}

func TestGetAssignment(t *testing.T) {
	p := ctProvider(t)
	_, err := p.getAssignment(p.server)
	if err != nil {
		t.Fatalf("Failed to get assignment: %s", err)
	}
}

func TestGetEnrollmentToken(t *testing.T) {
	p := ctProvider(t)
	_, err := p.getDeviceEnrollmentToken()
	if err != nil {
		t.Fatalf("Failed to get enrollment token: %s", err)
	}
}
