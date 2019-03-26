package kmsvaultsecret

import (
	vaultapi "github.com/hashicorp/vault/api"
	k8sv1alpha1 "github.com/patoarvizu/kms-vault-operator/pkg/apis/k8s/v1alpha1"
)

type KVv1Writer struct{}

func (KVv1 KVv1Writer) write(secret *k8sv1alpha1.KMSVaultSecretSpec, vaultClient *vaultapi.Client) error {
	decryptedSecret, err := decryptSecret(secret.Secret.EncryptedSecret)
	if err != nil {
		return err
	}
	writeData := map[string]interface{}{
		secret.Secret.Key: decryptedSecret,
	}
	_, writeErr := vaultClient.Logical().Write(secret.Path, writeData)
	if writeErr != nil {
		return writeErr
	}
	return nil
}
