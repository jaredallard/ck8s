package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComputerType is the type of computer requested for this pod to run on
type ComputerType string

const (
	ComputerTypeTurtle   ComputerType = "turtle"
	ComputerTypeRegular  ComputerType = "regular"
	ComputerTypeAdvanced ComputerType = "advanced"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ComputerPodSpec defines the desired state of ComputerPod
// +k8s:openapi-gen=true
type ComputerPodSpec struct {
	// ComputerType is a type of computer to run on
	ComputerType ComputerType `json:"computerType,omitempty"`

	// PastebinID is a pastebin ID to download code from
	PastebinID string `json:"pastebinID,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComputerPod is the Schema for the computerpods API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=computerpods,scope=Namespaced
type ComputerPod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ComputerPodSpec  `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status corev1.PodStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComputerPodList contains a list of ComputerPod
type ComputerPodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ComputerPod `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&ComputerPod{}, &ComputerPodList{})
}
