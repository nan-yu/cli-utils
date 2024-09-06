// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// UnstructuredFields Implements fields.Fields to do field selection on any
// field in an unstructured object.
type UnstructuredFields struct {
	Object *unstructured.Unstructured
}

// Has returns whether the provided field exists.
func (uf *UnstructuredFields) Has(field string) (exists bool) {
	_, found, err := unstructured.NestedString(uf.Object.Object, uf.fields(field)...)
	return err == nil && found
}

// Get returns the value for the provided field.
func (uf *UnstructuredFields) Get(field string) (value string) {
	val, found, err := unstructured.NestedString(uf.Object.Object, uf.fields(field)...)
	if err != nil || !found {
		return ""
	}
	return val
}

func (uf *UnstructuredFields) fields(field string) []string {
	field = strings.TrimPrefix(field, ".")
	return strings.Split(field, ".")
}
