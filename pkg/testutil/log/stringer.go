// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"

	"github.com/kylelemons/godebug/diff"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	jserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/json"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/cli-utils/pkg/kinds"
	"sigs.k8s.io/yaml"
)

type jsonStringer struct {
	O interface{}
}

// AsJSON returns a new stringer object that delays marshaling until the
// String method is called. For logging at higher verbosity levels, to
// avoid formatting when the output isn't going to be used.
func AsJSON(o interface{}) fmt.Stringer {
	return &jsonStringer{O: o}
}

// String returns the object as json, or the error string if marshalling fails.
func (ojs *jsonStringer) String() string {
	bytes, err := json.Marshal(ojs.O)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

type yamlStringer struct {
	O      interface{}
	Scheme *runtime.Scheme
}

// AsYAML returns a new stringer object that delays marshaling until the
// String method is called. For logging at higher verbosity levels, to
// avoid formatting when the output isn't going to be used.
// The primary use is for logging Kubernetes objects, but should also work
// with other types, like Go structs.
func AsYAML(o interface{}) fmt.Stringer {
	return &yamlStringer{O: o}
}

// AsYAMLWithScheme is similar to AsYAML, except it allows specifying which
// scheme to use to encode the object, instead of defaulting to the global
// `core.Scheme`.
func AsYAMLWithScheme(obj runtime.Object, scheme *runtime.Scheme) fmt.Stringer {
	return &yamlStringer{O: obj, Scheme: scheme}
}

// String returns the object as yaml, or the error string if marshalling fails.
func (oys *yamlStringer) String() string {
	// Use scheme-aware serialization, if possible.
	// This adds type fields and orders consistently.
	if rObj, ok := oys.O.(runtime.Object); ok {
		scheme := oys.Scheme
		// Default to the global scheme, if unspecified
		if scheme == nil {
			scheme = k8sscheme.Scheme
		}
		// Make best effort to ensure GVK is set
		_, isUnstructured := rObj.(*unstructured.Unstructured)
		if !isUnstructured && rObj.GetObjectKind().GroupVersionKind().Empty() {
			gvk, err := kinds.Lookup(rObj, scheme)
			// do nothing if lookup errors
			if err == nil {
				// copy the object to avoid side effects
				rObj = rObj.DeepCopyObject()
				rObj.GetObjectKind().SetGroupVersionKind(gvk)
			}
		}
		// Encode
		yamlSerializer := jserializer.NewYAMLSerializer(jserializer.DefaultMetaFactory, scheme, scheme)
		bytes, err := runtime.Encode(yamlSerializer, rObj)
		if err != nil {
			return err.Error()
		}
		return string(bytes)
	}
	// Default to general yaml serializer
	bytes, err := yaml.Marshal(oys.O)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

type yamlDiffStringer struct {
	Old, New interface{}
	Scheme   *runtime.Scheme
}

// AsYAMLDiff returns a new stringer object that delays marshaling and diffing
// until the String method is called. For logging at higher verbosity levels, to
// avoid formatting when the output isn't going to be used.
// The primary use is for comparing two Kubernetes objects, but should also work
// with other types, like Go structs.
// Does not do any object type or version conversion.
func AsYAMLDiff(old, new interface{}) fmt.Stringer {
	return &yamlDiffStringer{Old: old, New: new}
}

// AsYAMLDiffWithScheme is similar to AsYAMLDiff, except it allows specifying
// which scheme to use to encode the objects, instead of defaulting to the
// global `core.Scheme`.
func AsYAMLDiffWithScheme(old, new runtime.Object, scheme *runtime.Scheme) fmt.Stringer {
	return &yamlDiffStringer{Old: old, New: new, Scheme: scheme}
}

// String returns a diff (- Removed, + Added) of the objects as yaml, or the
// error string if marshalling fails.
// Uses diff.Diff to print full yaml, instead of cmp.Diff which truncates.
func (yds *yamlDiffStringer) String() string {
	if yds.Scheme != nil {
		// Must be either runtime.Object or nil.
		// Don't panic trying to cast nil interface{} to runtime.Object.
		oldObj, _ := yds.Old.(runtime.Object)
		newObj, _ := yds.New.(runtime.Object)
		return diff.Diff(
			AsYAMLWithScheme(oldObj, yds.Scheme).String(),
			AsYAMLWithScheme(newObj, yds.Scheme).String())
	}
	return diff.Diff(AsYAML(yds.Old).String(), AsYAML(yds.New).String())
}
