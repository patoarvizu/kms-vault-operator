package kmsvaultsecret

import (
	"errors"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

type VaultTokenAuth struct{}

func (k8s VaultTokenAuth) login(vaultConfig *vaultapi.Config) (string, error) {
	vaultToken, set := os.LookupEnv("VAULT_TOKEN")
	if !set {
		return "", errors.New("VAULT_TOKEN environment variable not found")
	}
	return vaultToken, nil
}
