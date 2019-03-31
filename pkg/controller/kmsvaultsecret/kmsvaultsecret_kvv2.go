package kmsvaultsecret

import (
	"encoding/json"
	"errors"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	k8sv1alpha1 "github.com/patoarvizu/kms-vault-operator/pkg/apis/k8s/v1alpha1"
)

type KVv2Writer struct{}

func (KVv2 KVv2Writer) write(secret *k8sv1alpha1.KMSVaultSecret, vaultClient *vaultapi.Client) error {
	read, _ := vaultClient.Logical().Read(secret.Spec.Path)
	if read != nil {
		metadata := read.Data["metadata"].(map[string]interface{})
		version, err := metadata["version"].(json.Number).Int64()
		if err != nil {
			return errors.New("Can't parse secret metadata")
		}
		if secret.Spec.KVSettings.CASIndex+1 < int(version) {
			return errors.New("CAS index is lower than the latest version")
		}
		if secret.Spec.KVSettings.CASIndex+1 == int(version) {
			return nil
		}
	}
	decryptedSecretData, err := decryptSecrets(secret.Spec.Secrets)
	if err != nil {
		return err
	}
	writeData := map[string]interface{}{
		"data": decryptedSecretData,
		"options": map[string]int{
			"cas": secret.Spec.KVSettings.CASIndex,
		},
	}
	_, err = vaultClient.Logical().Write(secret.Spec.Path, writeData)
	if err != nil {
		return err
	}
	return nil
}

func (KVv2 KVv2Writer) delete(secret *k8sv1alpha1.KMSVaultSecret, vaultClient *vaultapi.Client) error {
	deletePath := strings.Replace(secret.Spec.Path, "secret/data/", "secret/metadata/", 1)
	_, err := vaultClient.Logical().Delete(deletePath)
	return err
}
