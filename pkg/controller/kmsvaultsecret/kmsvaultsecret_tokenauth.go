package kmsvaultsecret

import (
	"errors"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

type VaultTokenAuth struct{}

func (k8s VaultTokenAuth) login(vaultConfig *vaultapi.Config) (*vaultapi.Secret, error) {
	vaultToken, set := os.LookupEnv("VAULT_TOKEN")
	if !set {
		return nil, errors.New("VAULT_TOKEN environment variable not found")
	}
	return &vaultapi.Secret{Auth: &vaultapi.SecretAuth{ClientToken: vaultToken}}, nil
}
