package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
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

	Spec   ComputerPodSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ComputerPodStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ComputerPodStatus represents information about the status of a computerpod. Status may trail the actual
// state of a system, especially if the node that hosts the computerpod cannot contact the control
// plane.
type ComputerPodStatus struct {
	Phase v1.PodPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=PodPhase"`

	// A human readable message indicating details about why the computerpod is in this condition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// A brief CamelCase message indicating details about why the computerpod is in this state.
	// e.g. 'Evicted'
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`

	// AssignedComputer is the computer resource that this ComputerPod has been scheduled onto
	// +optional
	AssignedComputer string `json:"assignedComputer" protobuf:"bytes,4,opt,name=assignedComputer"`

	// RFC 3339 date and time at which the object was acknowledged by the Kubelet.
	// This is before the Kubelet pulled the container image(s) for the pod.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty" protobuf:"bytes,5,opt,name=startTime"`
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
