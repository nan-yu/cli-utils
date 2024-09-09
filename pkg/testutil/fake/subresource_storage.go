// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cli-utils/pkg/kinds"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SubresourceStorage is a wrapper around MemoryStorage that allows modifying
// a specific top-level field without updating any other fields.
type SubresourceStorage struct {
	// Storage is the backing store for full resource objects
	Storage *MemoryStorage
	// Field is the sub-resource field managed by this SubresourceStorage
	Field string
}

func (ss *SubresourceStorage) getSubresourceInterface(uObj *unstructured.Unstructured) (interface{}, bool, error) {
	return object.NestedField(uObj.Object, ss.Field)
}

func (ss *SubresourceStorage) setSubresourceInterface(uObj *unstructured.Unstructured, value interface{}) error {
	return unstructured.SetNestedField(uObj.Object, value, ss.Field)
}

func (ss *SubresourceStorage) validateSubResourceUpdateOptions(opts *client.SubResourceUpdateOptions) error {
	return ss.Storage.validateUpdateOptions(&opts.UpdateOptions)
}

// Update the sub-resource field. All other fields are ignored.
func (ss *SubresourceStorage) Update(ctx context.Context, obj client.Object, opts *client.SubResourceUpdateOptions) error {
	ss.Storage.lock.Lock()
	defer ss.Storage.lock.Unlock()

	err := ss.validateSubResourceUpdateOptions(opts)
	if err != nil {
		return err
	}

	id, err := kinds.LookupID(obj, ss.Storage.scheme)
	if err != nil {
		return err
	}

	cachedObj, found := ss.Storage.objects[id]
	if !found {
		return newNotFound(id)
	}

	storageGVK, err := ss.Storage.storageGVK(obj)
	if err != nil {
		return err
	}

	// Convert to Unstructured and the storage version.
	// Don't use prepareObject, because we don't want to minimize yet.
	uObj, err := kinds.ToUnstructuredWithVersion(obj, storageGVK, ss.Storage.scheme)
	if err != nil {
		return err
	}

	newSubresourceValue, hasSubresource, err := ss.getSubresourceInterface(uObj)
	if err != nil {
		return err
	}

	// TODO: Figure out how to check if the resource in the scheme has this sub-resource.
	if !hasSubresource {
		return fmt.Errorf("the %s object %s does not have a %q sub-resource field",
			id.GroupKind, id.ObjectKey, ss.Field)
	}

	if len(opts.DryRun) > 0 {
		// don't merge or store the result
		return nil
	}

	if obj.GetUID() != "" && obj.GetUID() != cachedObj.GetUID() {
		return newConflictingUID(id, obj.GetResourceVersion(), cachedObj.GetResourceVersion())
	}
	if obj.GetResourceVersion() != "" && obj.GetResourceVersion() != cachedObj.GetResourceVersion() {
		return newConflictingResourceVersion(id, obj.GetResourceVersion(), cachedObj.GetResourceVersion())
	}

	// Copy cached object so we can diff the changes later
	updatedObj := cachedObj.DeepCopy()

	err = incrementResourceVersion(updatedObj)
	if err != nil {
		return fmt.Errorf("failed to increment resourceVersion: %w", err)
	}

	// Assume status doesn't affect generation (don't increment).

	err = ss.setSubresourceInterface(updatedObj, newSubresourceValue)
	if err != nil {
		return err
	}

	// Copy latest values back to input object
	obj.SetUID(updatedObj.GetUID())
	obj.SetResourceVersion(updatedObj.GetResourceVersion())
	obj.SetGeneration(updatedObj.GetGeneration())

	klog.V(5).Infof("Updating Status %s (ResourceVersion: %q)",
		kinds.ObjectSummary(updatedObj), updatedObj.GetResourceVersion())

	cachedObj, diff, err := ss.Storage.putWithoutLock(id, updatedObj)
	if err != nil {
		return err
	}
	// Copy everything back to input object, even if no diff
	if err := ss.Storage.scheme.Convert(cachedObj, obj, nil); err != nil {
		return fmt.Errorf("failed to update input object: %w", err)
	}
	// TODO: Remove GVK from typed objects
	obj.GetObjectKind().SetGroupVersionKind(cachedObj.GroupVersionKind())
	if diff {
		return ss.Storage.sendPutEvent(ctx, id, watch.Modified)
	}
	return nil
}

// Patch the sub-resource field. All other fields are ignored.
func (ss *SubresourceStorage) Patch(_ context.Context, _ client.Object, _ client.Patch, _ *client.SubResourcePatchOptions) error {
	ss.Storage.lock.Lock()
	defer ss.Storage.lock.Unlock()

	// TODO: Implement sub-resource patch, if needed
	return errors.New("fake.SubresourceStorage.Patch: not yet implemented")
}
