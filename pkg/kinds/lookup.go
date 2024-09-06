// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kinds

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Lookup returns the GVK of a object based on the types registered with the
// provided Scheme.
func Lookup(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("failed to lookup object type: %w", err)
	}
	return gvk, nil
}

// LookupID returns the object's ID. If the GK isn't already populated, the
// Scheme is used to look it up by object type.
func LookupID(obj client.Object, scheme *runtime.Scheme) (ID, error) {
	id := IDOf(obj)
	if id.GroupKind.Empty() {
		gvk, err := Lookup(obj, scheme)
		if err != nil {
			return id, err
		}
		id.GroupKind = gvk.GroupKind()
	}
	return id, nil
}
