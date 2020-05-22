// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cloudeos

import (
	"encoding/json"
	"log"

	"github.com/iancoleman/strcase"
	"google.golang.org/genproto/protobuf/field_mask"
)

/*******************************************************
* Use getOuterFieldMask() to get fieldMask of the provided protobuf struct.
The returned fieldMask will have field names of struct set.
*
* For slices/arrays of embedded structs within a protobuf struct, the fieldMask is expected to be
set within the fieldMask of the embedded struct.
Use getOuterFieldMask() for this or invoke getFieldMask() with prefix=""
For eg:
type Bar struct {
	id string
	fieldMask *field_mask.FieldMask
}

type Foo struct {
	name string
	bar []*bar
	fieldMask *field_mask.FieldMask
}

An object of type Foo should look as below
{name:"foo1", bar:{id:"xyz", fieldMask:{paths:id}},
bar:{id:"abc", fieldMask:{paths:id}}, fieldMask:{paths:name, paths:bar}

* For embedded structs within a protobuf struct, the field names of the embedded struct should be
prefixed with the field name representing the embedded struct in the outer protobuf struct.
For eg:
type struct Bar {
	id string
	fieldMask *field_mask.FieldMask
}
type struct Foo {
	name string
	bar *Bar
	fieldMask *field_mask.FieldMask
}

An object of type Foo should look as below
{name:"foo1", bar:{id:"xyz", fieldMask:{}}, fieldMask:{paths:name, paths:bar.id}

Use appendInnerFieldMask() to append prefixed inner field
mask to the outer field mask.
*********************************************************/

// Returns outer field mask if prefix is empty.
// Returns inner field masks if prefix is provided.
func getFieldMask(pbStruct interface{}, prefix string) (*field_mask.FieldMask, error) {
	// Convert the protobuf struct to json. Because of the omitempty tags in protobuf struct,
	// the marshalled json will have only fields that are set to non-default values.
	pbJSON, err := json.Marshal(pbStruct)
	if err != nil {
		log.Print("Failed to marshal protobuf struct")
		return nil, err
	}

	// Convert the json to map so as to retrieve the field names.
	pbMap := make(map[string]interface{})
	err = json.Unmarshal(pbJSON, &pbMap)
	if err != nil {
		log.Print("Failed to unmarshal protobuf json to map")
		return nil, err
	}

	// The marshalled json has field names corresponding to the json tags of
	// the protobuf fields, while fieldMask in the grpc message is expected to be in
	// lower camel case. So extract the field names in the json and convert it
	// to lower camel case before adding it to the fieldMask.
	var fieldMask field_mask.FieldMask
	for key := range pbMap {
		path := strcase.ToLowerCamel(key)
		if prefix != "" {
			path = prefix + strcase.ToLowerCamel(key)
		}
		fieldMask.Paths = append(fieldMask.Paths, path)
	}

	return &fieldMask, nil
}

func getOuterFieldMask(pbStruct interface{}) (*field_mask.FieldMask, error) {
	fm, err := getFieldMask(pbStruct, "")
	return fm, err
}

func getPrefixedFieldMask(pbStruct interface{}, prefix string) (*field_mask.FieldMask, error) {
	fm, err := getFieldMask(pbStruct, prefix)
	return fm, err
}

// Appends prefixed inner fields in the outer field mask.
func appendInnerFieldMask(innerPbStruct interface{}, outerFieldMask *field_mask.FieldMask,
	innerFieldMaskPrefix string) error {
	// We might have set the embedded struct field name in the outerFieldMask.
	// Remove that when we set the prefixed inner fields of the embedded
	// struct in the outer field mask.
	for i, path := range outerFieldMask.Paths {
		if path == innerFieldMaskPrefix[0:len(innerFieldMaskPrefix)-1] {
			outerFieldMask.Paths = append(outerFieldMask.Paths[:i], outerFieldMask.Paths[i+1:]...)
		}
	}

	// Get prefixed inner field mask and append it to outer field mask
	innerFieldMask, err := getPrefixedFieldMask(innerPbStruct, innerFieldMaskPrefix)
	if err != nil {
		return err
	}

	outerFieldMask.Paths = append(outerFieldMask.Paths,
		innerFieldMask.Paths...)
	return nil
}
