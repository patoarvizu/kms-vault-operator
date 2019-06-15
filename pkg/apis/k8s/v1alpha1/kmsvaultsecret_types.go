package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KMSVaultSecretSpec defines the desired state of KMSVaultSecret
// +k8s:openapi-gen=true
type KMSVaultSecretSpec struct {
	Path       string     `json:"path"`
	Secrets    []Secret   `json:"secrets"`
	KVSettings KVSettings `json:"kvSettings"`

	// +kubebuilder:validation:Enum=k8s,token
	VaultAuthMethod string `json:"vaultAuthMethod"`

	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

type KVSettings struct {
	// +kubebuilder:validation:Enum=v1,v2
	EngineVersion string `json:"engineVersion"`
	// +kubebuilder:validation:Minimum=0
	CASIndex int `json:"casIndex,omitempty"`
}

type Secret struct {
	Key             string            `json:"key"`
	EncryptedSecret string            `json:"encryptedSecret"`
	SecretContext   map[string]string `json:"secretContext,omitempty"`
}

// KMSVaultSecretStatus defines the observed state of KMSVaultSecret
// +k8s:openapi-gen=true
type KMSVaultSecretStatus struct {
	Created bool `json:"created,omitempty"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KMSVaultSecret is the Schema for the kmsvaultsecrets API
// +k8s:openapi-gen=true
type KMSVaultSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KMSVaultSecretSpec   `json:"spec,omitempty"`
	Status KMSVaultSecretStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KMSVaultSecretList contains a list of KMSVaultSecret
type KMSVaultSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KMSVaultSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KMSVaultSecret{}, &KMSVaultSecretList{})
}
