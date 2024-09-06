// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kinds

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GVKToString returns a human readable format for GroupVersionKind:
// `GROUP/VERSION.KIND`
func GVKToString(gvk schema.GroupVersionKind) string {
	return fmt.Sprintf("%s.%s", gvk.GroupVersion(), gvk.Kind)
}

// ObjectSummary returns a human readable format for objects.
// Depending on the type and what's populated, the output may be one of the
// following:
// - `GROUP/VERSION.KIND(TYPE)[NAMESPACE/NAME]`
// - `GROUP/VERSION.KIND(TYPE)`
// - `(TYPE)`
func ObjectSummary(obj runtime.Object) string {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gvkStr := ""
	if !gvk.Empty() {
		gvkStr = GVKToString(gvk)
	}
	if cObj, ok := obj.(client.Object); ok {
		keyStr := client.ObjectKeyFromObject(cObj).String()
		if keyStr != string(types.Separator) {
			return fmt.Sprintf("%s(%T)[%s]", gvkStr, obj, keyStr)
		}
	}
	return fmt.Sprintf("%s(%T)", gvkStr, obj)
}
