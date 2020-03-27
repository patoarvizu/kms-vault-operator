package kmsvaultsecret

import (
	vaultapi "github.com/hashicorp/vault/api"
	k8sv1alpha1 "github.com/patoarvizu/kms-vault-operator/pkg/apis/k8s/v1alpha1"
)

type KVv1Writer struct{}

func (KVv1 KVv1Writer) write(secret *k8sv1alpha1.KMSVaultSecret, vaultClient *vaultapi.Client) error {
	decryptedSecretData, err := decryptSecrets(secret)
	if err != nil {
		return err
	}
	_, err = vaultClient.Logical().Write(secret.Spec.Path, decryptedSecretData)
	if err != nil {
		return err
	}
	return nil
}

func (KVv1 KVv1Writer) delete(secret *k8sv1alpha1.KMSVaultSecret, vaultClient *vaultapi.Client) error {
	_, err := vaultClient.Logical().Delete(secret.Spec.Path)
	return err
}
