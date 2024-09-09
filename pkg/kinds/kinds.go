// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kinds

import (
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RoleBinding returns the canonical RoleBinding GroupVersionKind.
func RoleBinding() schema.GroupVersionKind {
	return rbacv1.SchemeGroupVersion.WithKind("RoleBinding")
}

// RoleBindingV1Beta1 returns the canonical v1beta1 RoleBinding GroupVersionKind.
func RoleBindingV1Beta1() schema.GroupVersionKind {
	return rbacv1beta1.SchemeGroupVersion.WithKind("RoleBinding")
}

// Role returns the canonical Role GroupVersionKind.
func Role() schema.GroupVersionKind {
	return rbacv1.SchemeGroupVersion.WithKind("Role")
}

// ResourceQuota returns the canonical ResourceQuota GroupVersionKind.
func ResourceQuota() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("ResourceQuota")
}

// PersistentVolume returns the canonical PersistentVolume GroupVersionKind.
func PersistentVolume() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("PersistentVolume")
}

// PodSecurityPolicy returns the canonical PodSecurityPolicy GroupVersionKind.
func PodSecurityPolicy() schema.GroupVersionKind {
	return policyv1beta1.SchemeGroupVersion.WithKind("PodSecurityPolicy")
}

// Namespace returns the canonical Namespace GroupVersionKind.
func Namespace() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("Namespace")
}

// CustomResourceDefinitionKind is the Kind for CustomResourceDefinitions
const CustomResourceDefinitionKind = "CustomResourceDefinition"

// CustomResourceDefinitionV1Beta1 returns the v1beta1 CustomResourceDefinition GroupVersionKind.
func CustomResourceDefinitionV1Beta1() schema.GroupVersionKind {
	return CustomResourceDefinition().WithVersion(v1.SchemeGroupVersion.Version)
}

// CustomResourceDefinitionV1 returns the v1 CustomResourceDefinition GroupVersionKind.
func CustomResourceDefinitionV1() schema.GroupVersionKind {
	return CustomResourceDefinition().WithVersion("v1")
}

// CustomResourceDefinition returns the canonical CustomResourceDefinition GroupKind
func CustomResourceDefinition() schema.GroupKind {
	return schema.GroupKind{
		Group: v1.GroupName,
		Kind:  CustomResourceDefinitionKind,
	}
}

// ClusterRoleBinding returns the canonical ClusterRoleBinding GroupVersionKind.
func ClusterRoleBinding() schema.GroupVersionKind {
	return rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding")
}

// ClusterRoleBindingV1Beta1 returns the canonical ClusterRoleBinding GroupVersionKind.
func ClusterRoleBindingV1Beta1() schema.GroupVersionKind {
	return rbacv1beta1.SchemeGroupVersion.WithKind("ClusterRoleBinding")
}

// ClusterRole returns the canonical ClusterRole GroupVersionKind.
func ClusterRole() schema.GroupVersionKind {
	return rbacv1.SchemeGroupVersion.WithKind("ClusterRole")
}

// Cluster returns the canonical Cluster GroupVersionKind.
func Cluster() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "clusterregistry.k8s.io", Version: "v1alpha1", Kind: "Cluster"}
}

// Deployment returns the canonical Deployment GroupVersionKind.
func Deployment() schema.GroupVersionKind {
	return appsv1.SchemeGroupVersion.WithKind("Deployment")
}

// Pod returns the canonical Pod GroupVersionKind.
func Pod() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("Pod")
}

// DaemonSet returns the canonical DaemonSet GroupVersionKind.
func DaemonSet() schema.GroupVersionKind {
	return appsv1.SchemeGroupVersion.WithKind("DaemonSet")
}

// Ingress returns the canonical Ingress GroupVersionKind.
func Ingress() schema.GroupVersionKind {
	return networkingv1.SchemeGroupVersion.WithKind("Ingress")
}

// ReplicaSet returns the canonical ReplicaSet GroupVersionKind.
func ReplicaSet() schema.GroupVersionKind {
	return appsv1.SchemeGroupVersion.WithKind("ReplicaSet")
}

// NetworkPolicy returns the canonical NetworkPolicy GroupVersionKind.
func NetworkPolicy() schema.GroupVersionKind {
	return networkingv1.SchemeGroupVersion.WithKind("NetworkPolicy")
}

// ConfigMap returns the canonical ConfigMap GroupVersionKind.
func ConfigMap() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("ConfigMap")
}

// Job returns the canonical Job GroupVersionKind.
func Job() schema.GroupVersionKind {
	return batchv1.SchemeGroupVersion.WithKind("Job")
}

// CronJob returns the canonical CronJob GroupVersionKind.
func CronJob() schema.GroupVersionKind {
	return batchv1.SchemeGroupVersion.WithKind("CronJob")
}

// ReplicationController returns the canonical ReplicationController GroupVersionKind.
func ReplicationController() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("ReplicationController")
}

// StatefulSet returns the canonical StatefulSet GroupVersionKind.
func StatefulSet() schema.GroupVersionKind {
	return appsv1.SchemeGroupVersion.WithKind("StatefulSet")
}

// Service returns the canonical Service GroupVersionKind.
func Service() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("Service")
}

// Secret returns the canonical Secret GroupVersionKind.
func Secret() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("Secret")
}

// ServiceAccount returns the canonical ServiceAccount GroupVersionKind.
func ServiceAccount() schema.GroupVersionKind {
	return corev1.SchemeGroupVersion.WithKind("ServiceAccount")
}

// APIService returns the APIService kind.
func APIService() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "apiregistration.k8s.io", Version: "v1", Kind: "APIService"}
}

// ValidatingWebhookConfiguration returns the ValidatingWebhookConfiguration kind.
func ValidatingWebhookConfiguration() schema.GroupVersionKind {
	return admissionv1.SchemeGroupVersion.WithKind("ValidatingWebhookConfiguration")
}
