// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kinds

import (
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/cli-utils/pkg/testutil"
)

func TestLookup(t *testing.T) {
	emptyScheme := runtime.NewScheme()
	coreScheme := runtime.NewScheme()
	if err := corev1.AddToScheme(coreScheme); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name          string
		object        runtime.Object
		scheme        *runtime.Scheme
		expected      schema.GroupVersionKind
		expectedError error
	}{
		{
			name: "unstructured pre-populated GVK not in scheme",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": Service().GroupVersion().String(),
					"kind":       Service().Kind,
					"metadata": map[string]interface{}{
						"name": "test-name",
					},
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"app.kubernetes.io/name": "MyApp",
						},
						"ports": []interface{}{
							map[string]interface{}{
								"protocol":   "TCP",
								"port":       80,
								"targetPort": 9376,
							},
						},
					},
				},
			},
			scheme:   emptyScheme,
			expected: Service(),
		},
		{
			name: "unstructured pre-populated GVK in scheme",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": Service().GroupVersion().String(),
					"kind":       Service().Kind,
					"metadata": map[string]interface{}{
						"name": "test-name",
					},
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"app.kubernetes.io/name": "MyApp",
						},
						"ports": []interface{}{
							map[string]interface{}{
								"protocol":   "TCP",
								"port":       80,
								"targetPort": 9376,
							},
						},
					},
				},
			},
			scheme:   coreScheme,
			expected: Service(),
		},
		{
			name: "typed pre-populated GVK not in scheme",
			object: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: Service().GroupVersion().String(),
					Kind:       Service().Kind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-name",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/name": "MyApp",
					},
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       80,
							TargetPort: intstr.FromInt(9376),
						},
					},
				},
			},
			scheme: emptyScheme,
			expectedError: fmt.Errorf("failed to lookup object type: %w",
				runtime.NewNotRegisteredErrForType(emptyScheme.Name(),
					reflect.TypeOf(corev1.Service{}))),
		},
		{
			name: "typed pre-populated GVK in scheme",
			object: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: Service().GroupVersion().String(),
					Kind:       Service().Kind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-name",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/name": "MyApp",
					},
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       80,
							TargetPort: intstr.FromInt(9376),
						},
					},
				},
			},
			scheme:   coreScheme,
			expected: Service(),
		},
		{
			name: "typed unpopulated GVK not in scheme",
			object: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-name",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/name": "MyApp",
					},
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       80,
							TargetPort: intstr.FromInt(9376),
						},
					},
				},
			},
			scheme: emptyScheme,
			expectedError: fmt.Errorf("failed to lookup object type: %w",
				runtime.NewNotRegisteredErrForType(emptyScheme.Name(),
					reflect.TypeOf(corev1.Service{}))),
		},
		{
			name: "typed unpopulated GVK in scheme",
			object: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-name",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/name": "MyApp",
					},
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       80,
							TargetPort: intstr.FromInt(9376),
						},
					},
				},
			},
			scheme:   coreScheme,
			expected: Service(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Lookup(tc.object, tc.scheme)
			testutil.AssertEqual(t, testutil.EqualError(tc.expectedError), testutil.EqualError(err))
			testutil.AssertEqual(t, tc.expected, actual)
		})
	}
}
