package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ComputerDeploymentSpec defines the desired state of ComputerDeployment
// +k8s:openapi-gen=true
type ComputerDeploymentSpec struct {
	Replicas int64       `json:"replicase,omitempty"`
	Template ComputerPod `json:"template,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComputerDeployment is the Schema for the computerdeployments API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=computerdeployments,scope=Namespaced
type ComputerDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComputerDeploymentSpec  `json:"spec,omitempty"`
	Status appsv1.DeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComputerDeploymentList contains a list of ComputerDeployment
type ComputerDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputerDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputerDeployment{}, &ComputerDeploymentList{})
}
