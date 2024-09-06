// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kinds

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectAsClientObject casts from runtime.Object to client.Object.
// This method ensures a consistent error message for this common operation.
func ObjectAsClientObject(rObj runtime.Object) (client.Object, error) {
	cObj, ok := rObj.(client.Object)
	if !ok {
		return nil, fmt.Errorf("unsupported resource type (%s): failed to cast to client.Object",
			ObjectSummary(rObj))
	}
	return cObj, nil
}

// ObjectAsClientObjectList casts from runtime.Object to client.ObjectList.
// This method ensures a consistent error message for this common operation.
func ObjectAsClientObjectList(rObj runtime.Object) (client.ObjectList, error) {
	cObj, ok := rObj.(client.ObjectList)
	if !ok {
		return nil, fmt.Errorf("unsupported resource type (%s): failed to cast to client.ObjectList",
			ObjectSummary(rObj))
	}
	return cObj, nil
}
