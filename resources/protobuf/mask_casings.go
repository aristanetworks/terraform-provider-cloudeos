// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package protobuf

import (
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/generator" // nolint:staticcheck
	"google.golang.org/genproto/protobuf/field_mask"
)

// PathToGolangCasing converts the lowerCamelCase field mask standard to
// the GoCamelCase that the fields will be set/reflected on.
//
// Returned as the individual components. The user can decide whether to join
// rather than always joining (and usually splitting immediately)
func PathToGolangCasing(p string) []string {
	parts := strings.Split(p, ".")
	for i := range parts {
		parts[i] = generator.CamelCase(parts[i])
	}

	return parts
}

// MaskToGolangCasing converts the lowerCamelCase field mask standard to
// the GoCamelCase that the fields will be set/reflected on.
func MaskToGolangCasing(fm *field_mask.FieldMask) *field_mask.FieldMask {
	// be idempotent on nil
	if fm == nil {
		return nil
	}

	fixed := make([]string, 0, len(fm.Paths))
	for _, p := range fm.Paths {
		parts := strings.Split(p, ".")
		for i := range parts {
			parts[i] = generator.CamelCase(parts[i])
		}
		fixed = append(fixed, strings.Join(parts, "."))
	}

	return &field_mask.FieldMask{
		Paths: fixed,
	}
}

// FieldToMaskCasing converts a single field to lowerCamelCase.
// There should be no separators in the string (.) as this will give
// undesired results.
func FieldToMaskCasing(name string) string {
	// go full CamelCase then drop the leading char to camelCase
	camel := generator.CamelCase(name)
	return strings.ToLower(string(camel[0])) + camel[1:]
}

// PathToMaskCasing converts a series of field names to lowerCamelCase
// and then joins them with a "."
func PathToMaskCasing(names ...string) string {
	parts := make([]string, len(names))
	for _, n := range names {
		parts = append(parts, FieldToMaskCasing(n))
	}
	return strings.Join(parts, ".")
}
