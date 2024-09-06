// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kinds

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewObjectForGVK creates a new runtime.Object using the type registered to
// the scheme for the specified GroupVersionKind.
// This is a wrapper around scheme.New to provide a consistent error message.
func NewObjectForGVK(gvk schema.GroupVersionKind, scheme *runtime.Scheme) (runtime.Object, error) {
	rObj, err := scheme.New(gvk)
	if err != nil {
		return nil, fmt.Errorf("unsupported resource type (%s): %w", GVKToString(gvk), err)
	}
	return rObj, nil
}

// NewClientObjectForGVK creates a new client.Object using the type registered
// to the scheme for the specified GroupVersionKind.
//
// In practice, most runtime.Object are client.Object.
// However, some objects are not, namely objects used for config files that are
// not persisted by the Kubernetes API server.
// This method makes this common operation easier and ensures a consistent
// error message.
func NewClientObjectForGVK(gvk schema.GroupVersionKind, scheme *runtime.Scheme) (client.Object, error) {
	rObj, err := NewObjectForGVK(gvk, scheme)
	if err != nil {
		return nil, err
	}
	return ObjectAsClientObject(rObj)
}
