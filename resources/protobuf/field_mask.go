// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package protobuf

import (
	"fmt"
	"reflect"
	"strings"

	"cloudeos-terraform-provider/types/amap"

	"github.com/aristanetworks/glog"
	"github.com/aristanetworks/goarista/key"
	"github.com/golang/protobuf/protoc-gen-go/generator" // nolint:staticcheck
	fieldmask_utils "github.com/mennanov/fieldmask-utils"
	"google.golang.org/genproto/protobuf/field_mask"
)

// NewFieldMask initialized a new field mask with no contents
func NewFieldMask(p ...string) *field_mask.FieldMask {
	return &field_mask.FieldMask{
		Paths: p,
	}
}

// StructTagHasModifiable iterates the tags contained in the string and searches
// for the modifiable tag. A string is taken as argument for most compatability
// between complex types, reflect, and AST use cases (all aliased strings)
func StructTagHasModifiable(t string) bool {
	parts := strings.Split(strings.Trim(t, "`"), " ")
	for _, p := range parts {
		if p == "modifiable:\"true\"" {
			return true
		}
	}
	return false
}

// FieldMaskPathIsModifiable walks the given path along the given target
// and finally inspects the leaf field's tags to determine modifiability.
func FieldMaskPathIsModifiable(p string, tgt interface{}) bool {
	var sf reflect.StructField
	ty := reflect.TypeOf(tgt)
	parts := PathToGolangCasing(p)

	for _, p := range parts {
		nextTy, exists := ty.FieldByName(p)
		if !exists {
			return false // doesn't exist, so no
		}
		// go complains about single-line assignments
		sf = nextTy
		ty = nextTy.Type

		if ty.Kind() == reflect.Ptr {
			ty = ty.Elem()
		}
	}

	return StructTagHasModifiable(string(sf.Tag))
}

// GetFields returns a map of the resource's field values that has their respective field masks set.
func GetFields(fm *field_mask.FieldMask, tgt interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := FieldMaskForEach(fm, tgt, func(name string, val interface{}, modifiable bool) error {
		result[name] = val
		return nil
	})

	return result, err
}

// GetValueFromPath walks the given path along the given target
// and returns the value at the leaf-field
func GetValueFromPath(p string, tgt interface{}) (interface{}, error) {
	ty := reflect.TypeOf(tgt)
	val := reflect.ValueOf(tgt)
	parts := PathToGolangCasing(p)

	// deref any pointers from the arguments
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
		val = val.Elem()
	}

	for _, p := range parts {
		nextTy, exists := ty.FieldByName(p)
		if !exists {
			return nil, fmt.Errorf("failed to find path component %q", p)
		}
		nextVal := val.FieldByName(p)

		// go complains about single-line assignments
		ty = nextTy.Type
		if ty.Kind() == reflect.Ptr {
			ty = ty.Elem()
		}
		val = nextVal
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
	}

	return val.Interface(), nil
}

// OverlayMaskedFields takes in a FieldMask along with a source and destination
// struct. The FieldMask passed to this function should be from the source message.
//
// All fields set in the FieldMask will be copied to the destination struct with
// no other modifications being made.
func OverlayMaskedFields(fm *field_mask.FieldMask, src, dst interface{}) error {
	if fm == nil {
		return fmt.Errorf("cannot overlay nil FieldMask")
	}
	if src == nil {
		return fmt.Errorf("source for overlay cannot be nil")
	}
	if dst == nil {
		return fmt.Errorf("destination for overlay cannot be nil")
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(fm, generator.CamelCase)
	if err != nil {
		return fmt.Errorf("failed to extract masked fields: %s", err)
	}

	err = fieldmask_utils.StructToStruct(mask, src, dst)
	if err != nil {
		return fmt.Errorf("failed to copy field(s): %s", err)
	}

	return nil
}

// PassesFilter iterates the given FieldMask (belonging to filter) and checks
// that all masked fields match in both structs/types. If any field does not,
// false is returned.
func PassesFilter(fm *field_mask.FieldMask, filter, obj interface{}) (bool, error) {
	if fm == nil {
		return false, nil
	}

	for _, p := range fm.Paths {

		filt, err := GetValueFromPath(p, filter)
		if err != nil {
			return false, fmt.Errorf("failed to get value for %q from filter: %s", p, err)
		}

		val, err := GetValueFromPath(p, obj)
		if err != nil {
			return false, fmt.Errorf("failed to get value for %q from filter-target: %s", p, err)
		}

		if !reflect.DeepEqual(val, filt) {
			return false, nil
		}
	}

	return true, nil
}

// FieldMaskForEach iterates the given field mask and calls the callback.
//
// The callback is provided with the name (Golang case), value of the field,
// and if the field is modifiable. If there is an error in reflection or the
// callback returns an error, it is returned from this function.
func FieldMaskForEach(
	fm *field_mask.FieldMask, tgt interface{},
	cb func(string, interface{}, bool) error,
) error {
	mask := MaskToGolangCasing(fm)

	for _, p := range mask.Paths {
		// if nested, only check the first path component as that is the
		// modifiable we care about
		var isModifiable bool
		parts := strings.Split(p, ".")
		if len(parts) > 1 {
			isModifiable = FieldMaskPathIsModifiable(parts[0], tgt)
		} else {
			isModifiable = FieldMaskPathIsModifiable(p, tgt)
		}

		val, err := GetValueFromPath(p, tgt)
		if err != nil {
			return fmt.Errorf("failed to get value for %q: %s", p, err)
		}

		err = cb(p, val, isModifiable)
		if err != nil {
			return err
		}
	}
	return nil
}

// OverlayKeyMap takes a map of Aeris keys to values and a pointer to an object
// cast as an interface. The pointer is reflected on and the values inserted into
// the object.
func OverlayKeyMap(fields *amap.Map, tgt interface{}) (err error) {
	// safety to return error on reflect panic
	defer func() {
		if e := recover(); e != nil {
			glog.Errorf("failed to reflect on target in OverlayKeyMap: %s", err)
			err = e.(error)
		}
	}()

	elem := reflect.ValueOf(tgt).Elem()
	if err := fields.Iter(func(nameKey key.Key, val interface{}) error {
		name := generator.CamelCase(nameKey.String())

		f := elem.FieldByName(name)
		if !f.IsValid() {
			return fmt.Errorf("unexpected field %s for type %T", name, tgt)
		} else if !f.CanSet() {
			return fmt.Errorf("cannot set field %s in type %T", name, tgt)
		}

		f.Set(reflect.ValueOf(val))
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// FieldMaskIntersect returns the intersection of two field masks.
// This is useful when checking if fields you care about were updated in another
// field mask. One case is the filtering of a subscription limited to updates
// about certain fields.
func FieldMaskIntersect(a, b *field_mask.FieldMask) []string {
	amap := make(map[string]struct{})
	for _, p := range a.Paths {
		amap[p] = struct{}{}
	}

	result := make([]string, 0)
	for _, p := range b.Paths {
		_, exist := amap[p]
		if exist {
			result = append(result, p)
		}
	}

	return result
}
