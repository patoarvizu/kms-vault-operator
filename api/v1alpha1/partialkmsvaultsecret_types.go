package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PartialKMSVaultSecretSpec defines the desired state of PartialKMSVaultSecret
// +k8s:openapi-gen=true
type PartialKMSVaultSecretSpec struct {
	// +listType=map
	// +listMapKey=key
	Secrets       []Secret          `json:"secrets"`
	SecretContext map[string]string `json:"secretContext,omitempty"`
}

// PartialKMSVaultSecretStatus defines the observed state of PartialKMSVaultSecret
// +k8s:openapi-gen=true
type PartialKMSVaultSecretStatus struct {
	Created bool `json:"created,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PartialKMSVaultSecret is the Schema for the partialkmsvaultsecrets API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=partialkmsvaultsecrets,scope=Namespaced,shortName=pkmsvs
type PartialKMSVaultSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PartialKMSVaultSecretSpec   `json:"spec,omitempty"`
	Status PartialKMSVaultSecretStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PartialKMSVaultSecretList contains a list of PartialKMSVaultSecret
type PartialKMSVaultSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PartialKMSVaultSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PartialKMSVaultSecret{}, &PartialKMSVaultSecretList{})
}
