// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kinds"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// prioritizedGVKsAllGroups returns an list of GVKs known by the scheme, sorted
// by version priority within each group.
func prioritizedGVKsAllGroups(scheme *runtime.Scheme) []schema.GroupVersionKind {
	// map all known GroupVersionKinds by GroupVersion
	kindsForGVs := map[schema.GroupVersion][]schema.GroupVersionKind{}
	typeMap := scheme.AllKnownTypes()
	for gvk := range typeMap {
		gv := gvk.GroupVersion()
		kindsForGVs[gv] = append(kindsForGVs[gv], gvk)
	}

	// Flatten map into prioritized list
	var gvkList []schema.GroupVersionKind
	for _, gv := range scheme.PrioritizedVersionsAllGroups() {
		gvkList = append(gvkList, kindsForGVs[gv]...)
	}
	return gvkList
}

func toTypedClientObject(obj client.Object, scheme *runtime.Scheme) (client.Object, error) {
	tObj, err := kinds.ToTypedObject(obj, scheme)
	if err != nil {
		return nil, err
	}
	cObj, err := kinds.ObjectAsClientObject(tObj)
	if err != nil {
		return nil, err
	}
	return cObj, nil
}

// matchesListFilters returns true if the object matches the constraints
// specified by the ListOptions: Namespace, LabelSelector, and FieldSelector.
func matchesListFilters(obj runtime.Object, opts *client.ListOptions, scheme *runtime.Scheme) (bool, error) {
	labels, fields, accessor, err := getAttrs(obj, scheme)
	if err != nil {
		return false, err
	}
	if opts.Namespace != "" && opts.Namespace != accessor.GetNamespace() {
		// No match
		return false, nil
	}
	if opts.LabelSelector != nil && !opts.LabelSelector.Matches(labels) {
		// No match
		return false, nil
	}
	if opts.FieldSelector != nil && !opts.FieldSelector.Matches(fields) {
		// No match
		return false, nil
	}
	// Match!
	return true, nil
}

// getAttrs returns the label set and field set from an object that can be used
// for query filtering. This is roughly equivalent to what's in the apiserver,
// except only supporting the few metadata fields that are supported by CRDs.
func getAttrs(obj runtime.Object, scheme *runtime.Scheme) (labels.Set, fields.Fields, metav1.Object, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, nil, err
	}
	labelSet := labels.Set(accessor.GetLabels())

	uObj, err := kinds.ToUnstructured(obj, scheme)
	if err != nil {
		return nil, nil, nil, err
	}
	uFields := &UnstructuredFields{Object: uObj}

	return labelSet, uFields, accessor, nil
}

// convertToListItemType converts the object to the type of an item in the
// specified list. Does both object type conversion and version conversion.
func convertToListItemType(obj runtime.Object, objListType client.ObjectList, scheme *runtime.Scheme) (runtime.Object, bool, error) {
	// Lookup the List type from the scheme
	listGVK, err := kinds.Lookup(objListType, scheme)
	if err != nil {
		return nil, false, err
	}
	// Convert the List type to the Item type
	itemGVK := kinds.ItemGVKForListGVK(listGVK)
	if itemGVK == listGVK {
		return nil, false, fmt.Errorf("list kind does not have required List suffix: %s", listGVK.Kind)
	}

	if _, ok := objListType.(*unstructured.UnstructuredList); ok {
		// Convert to a unstructured object, optionally convert between versions
		uObj, err := kinds.ToUnstructuredWithVersion(obj, itemGVK, scheme)
		if err != nil {
			return nil, false, err
		}
		return uObj, true, nil
	}

	// Convert to a typed object, optionally convert between versions
	tObj, err := kinds.ToTypedWithVersion(obj, itemGVK, scheme)
	if err != nil {
		return nil, false, err
	}
	return tObj, true, nil
}

// toGR is a hack! only used for error messages, where the exact resource isn't very important.
// TODO: use the actual resource from mapper.RESTMapping
func toGR(gk schema.GroupKind) schema.GroupResource {
	return schema.GroupResource{
		Group:    gk.Group,
		Resource: gk.Kind,
	}
}

func newNotFound(id kinds.ID) error {
	return apierrors.NewNotFound(toGR(id.GroupKind), id.ObjectKey.String())
}

func newAlreadyExists(id kinds.ID) error {
	return apierrors.NewAlreadyExists(toGR(id.GroupKind), id.ObjectKey.String())
}

func newConflict(id kinds.ID, err error) error {
	return apierrors.NewConflict(toGR(id.GroupKind), id.ObjectKey.String(), err)
}

func newConflictingUID(id kinds.ID, expectedUID, foundUID string) error {
	return newConflict(id,
		fmt.Errorf("UID conflict: expected %q but found %q",
			expectedUID, foundUID))
}

func newConflictingResourceVersion(id kinds.ID, expectedRV, foundRV string) error {
	return newConflict(id,
		fmt.Errorf("ResourceVersion conflict: expected %q but found %q",
			expectedRV, foundRV))
}
