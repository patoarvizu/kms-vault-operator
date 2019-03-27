package kmsvaultsecret

import (
	vaultapi "github.com/hashicorp/vault/api"
	k8sv1alpha1 "github.com/patoarvizu/kms-vault-operator/pkg/apis/k8s/v1alpha1"
)

type KVv1Writer struct{}

func (KVv1 KVv1Writer) write(secret *k8sv1alpha1.KMSVaultSecret, vaultClient *vaultapi.Client) error {
	decryptedSecretData, err := decryptSecrets(secret.Spec.Secrets)
	if err != nil {
		return err
	}
	_, writeErr := vaultClient.Logical().Write(secret.Spec.Path, decryptedSecretData)
	if writeErr != nil {
		return writeErr
	}
	return nil
}
