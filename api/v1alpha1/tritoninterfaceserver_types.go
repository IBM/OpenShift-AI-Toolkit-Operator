/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TritonInterfaceServerSpec defines the desired state of TritonInterfaceServer
type TritonInterfaceServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster

	// +kubebuilder:required
	PvcName string `json:"pvcName"`

	// +kubebuilder:required
	MountPath string `json:"mountPath"`

	// +kubebuilder:default="icr.io/ibmz/ibmz-accelerated-for-nvidia-triton-inference-server@sha256:2cedd535805c316fec7dff6cac8129d873da39348459f645240eec005172b641"
	ServingImage string `json:"servingImage"`

	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=3
	Servers []Server `json:"servers"`

	// +kubebuilder:validation:Optional
	PodResources Resource `json:"podResources"`

	// +kubebuilder:validation:Optional
	GrpcConfig GrpcConfig `json:"grpcConfig"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas"`

	// +kubebuilder:validation:Optional
	RouteConfig RouteConfig `json:"routeConfig"`
}

type Server struct {
	// +kubebuilder:required
	// +kubebuilder:validation:Enum=HTTP;GRPC;Metrics
	Type string `json:"type"`
	// +kubebuilder:required
	Enabled bool `json:"enabled"`
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	ContainerPort uint32 `json:"containerPort"`
}

type Resource struct {
	// +kubebuilder:validation:Required
	Limits PodResource `json:"limits"`
	// +kubebuilder:validation:Required
	Requests PodResource `json:"requests"`
}

type PodResource struct {
	// +kubebuilder:validation:Required
	Cpu string `json:"cpu"`
	// +kubebuilder:validation:Required
	Memory string `json:"memory"`
}

type RouteConfig struct {
	// +kubebuilder:validation:Optional
	RouteAnnotation RouteAnnotationConfig `json:"routeAnnotation"`
}

type RouteAnnotationConfig struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:MinProperties=1
	Annotations map[string]string `json:"annotations"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=All;HTTP;GRPC;Metrics
	ApplyTo string `json:"applyTo"`
}

type GrpcConfig struct {
	// +kubebuilder:validation:Optional
	TlsSpec TlsSpec `json:"tlsSpec"`
}

type TlsSpec struct {
	// +kubebuilder:validation:Optional
	TlsSecretName string `json:"tlsSecretName"`
}

// TritonInterfaceServerStatus defines the observed state of TritonInterfaceServer
type TritonInterfaceServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TritonInterfaceServer is the Schema for the tritoninterfaceservers API
type TritonInterfaceServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TritonInterfaceServerSpec   `json:"spec,omitempty"`
	Status TritonInterfaceServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TritonInterfaceServerList contains a list of TritonInterfaceServer
type TritonInterfaceServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TritonInterfaceServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TritonInterfaceServer{}, &TritonInterfaceServerList{})
}
