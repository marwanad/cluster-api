/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Machine Reconciler", func() {
	It("Should create a Machine", func() {
		// TODO
	})
})

func TestReconcileRequest(t *testing.T) {
	RegisterTestingT(t)

	infraConfig := unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "InfrastructureConfig",
			"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha2",
			"metadata": map[string]interface{}{
				"name":      "infra-config1",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"providerID": "test://id-1",
			},
			"status": map[string]interface{}{
				"ready": true,
				"addresses": []interface{}{
					map[string]interface{}{
						"type":    "InternalIP",
						"address": "10.0.0.1",
					},
				},
			},
		},
	}
	machine1 := clusterv1.Machine{
		TypeMeta: metav1.TypeMeta{
			Kind: "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "create",
			Namespace:  "default",
			Finalizers: []string{clusterv1.MachineFinalizer, metav1.FinalizerDeleteDependents},
		},
		Spec: clusterv1.MachineSpec{
			InfrastructureRef: corev1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha2",
				Kind:       "InfrastructureConfig",
				Name:       "infra-config1",
			},
			Bootstrap: clusterv1.Bootstrap{Data: pointer.StringPtr("data")},
		},
	}
	machine2 := clusterv1.Machine{
		TypeMeta: metav1.TypeMeta{
			Kind: "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "update",
			Namespace:  "default",
			Finalizers: []string{clusterv1.MachineFinalizer, metav1.FinalizerDeleteDependents},
		},
		Spec: clusterv1.MachineSpec{
			InfrastructureRef: corev1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha2",
				Kind:       "InfrastructureConfig",
				Name:       "infra-config1",
			},
			Bootstrap: clusterv1.Bootstrap{Data: pointer.StringPtr("data")},
		},
	}
	time := metav1.Now()
	machine3 := clusterv1.Machine{
		TypeMeta: metav1.TypeMeta{
			Kind: "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "delete",
			Namespace:         "default",
			Finalizers:        []string{clusterv1.MachineFinalizer, metav1.FinalizerDeleteDependents},
			DeletionTimestamp: &time,
		},
		Spec: clusterv1.MachineSpec{
			InfrastructureRef: corev1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha2",
				Kind:       "InfrastructureConfig",
				Name:       "infra-config1",
			},
			Bootstrap: clusterv1.Bootstrap{Data: pointer.StringPtr("data")},
		},
	}
	clusterList := clusterv1.ClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind: "ClusterList",
		},
		Items: []clusterv1.Cluster{
			{
				TypeMeta: metav1.TypeMeta{
					Kind: "Cluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testcluster",
					Namespace: "default",
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind: "Cluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rainbow",
					Namespace: "foo",
				},
			},
		},
	}

	type expected struct {
		result reconcile.Result
		err    bool
	}
	testCases := []struct {
		request     reconcile.Request
		existsValue bool
		expected    expected
	}{
		{
			request:     reconcile.Request{NamespacedName: types.NamespacedName{Name: machine1.Name, Namespace: machine1.Namespace}},
			existsValue: false,
			expected: expected{
				result: reconcile.Result{},
				err:    false,
			},
		},
		{
			request:     reconcile.Request{NamespacedName: types.NamespacedName{Name: machine2.Name, Namespace: machine2.Namespace}},
			existsValue: true,
			expected: expected{
				result: reconcile.Result{},
				err:    false,
			},
		},
		{
			request:     reconcile.Request{NamespacedName: types.NamespacedName{Name: machine3.Name, Namespace: machine3.Namespace}},
			existsValue: true,
			expected: expected{
				result: reconcile.Result{},
				err:    false,
			},
		},
	}

	for _, tc := range testCases {
		clusterv1.AddToScheme(scheme.Scheme)
		r := &MachineReconciler{
			Client: fake.NewFakeClient(&clusterList, &machine1, &machine2, &machine3, &infraConfig),
			Log:    log.Log,
		}

		result, err := r.Reconcile(tc.request)
		if tc.expected.err {
			Expect(err).ToNot(BeNil())
		} else {
			Expect(err).To(BeNil())
		}

		Expect(result).To(Equal(tc.expected.result))
	}
}

func TestIsDeletionReady(t *testing.T) {
	RegisterTestingT(t)

	bootstrapConfig := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "BootstrapConfig",
			"apiVersion": "bootstrap.cluster.x-k8s.io/v1alpha2",
			"metadata": map[string]interface{}{
				"name":      "delete-bootstrap",
				"namespace": "default",
			},
		},
	}

	infraConfig := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "InfrastructureConfig",
			"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha2",
			"metadata": map[string]interface{}{
				"name":      "delete-infra",
				"namespace": "default",
			},
		},
	}

	machine := clusterv1.Machine{
		TypeMeta: metav1.TypeMeta{
			Kind: "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "delete",
			Namespace: "default",
		},
		Spec: clusterv1.MachineSpec{
			InfrastructureRef: corev1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha2",
				Kind:       "InfrastructureConfig",
				Name:       "delete-infra",
			},
			Bootstrap: clusterv1.Bootstrap{
				ConfigRef: &corev1.ObjectReference{
					APIVersion: "bootstrap.cluster.x-k8s.io/v1alpha2",
					Kind:       "BootstrapConfig",
					Name:       "delete-bootstrap",
				},
			},
		},
	}

	testCases := []struct {
		bootstrapExists bool
		infraExists     bool
		expected        bool
	}{
		{
			bootstrapExists: true,
			infraExists:     true,
			expected:        false,
		},
		{
			bootstrapExists: false,
			infraExists:     true,
			expected:        false,
		},
		{
			bootstrapExists: true,
			infraExists:     false,
			expected:        false,
		},
		{
			bootstrapExists: false,
			infraExists:     false,
			expected:        true,
		},
	}

	for _, tc := range testCases {
		myscheme := runtime.NewScheme()
		clusterv1.AddToScheme(myscheme)

		objs := []runtime.Object{&machine}

		if tc.bootstrapExists {
			objs = append(objs, bootstrapConfig)
		}

		if tc.infraExists {
			objs = append(objs, infraConfig)
		}

		r := &MachineReconciler{
			Client: fake.NewFakeClientWithScheme(myscheme, objs...),
			Log:    log.Log,
		}

		ok, _ := r.reconcileDeleteExternal(ctx, &machine)
		Expect(ok).To(Equal(tc.expected))
	}
}
