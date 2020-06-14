package kmsvaultsecret

import (
	"errors"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

const (
	appRoleDefaultEndpoint = "auth/approle/login"
)

type VaultAppRoleAuth struct{}

func (k8s VaultAppRoleAuth) login(vaultConfig *vaultapi.Config) (string, error) {
	roleId, ok := os.LookupEnv("VAULT_APPROLE_ROLE_ID")
	if !ok {
		return "", errors.New("Environment variable VAULT_APPROLE_ROLE_ID not set")
	}
	secretId, ok := os.LookupEnv("VAULT_APPROLE_SECRET_ID")
	if !ok {
		return "", errors.New("Environment variable VAULT_APPROLE_SECRET_ID not set")
	}
	vaultClient, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"role_id":   roleId,
		"secret_id": secretId,
	}
	appRoleEndpoint, ok := os.LookupEnv("VAULT_APPROLE_ENDPOINT")
	if !ok {
		appRoleEndpoint = appRoleDefaultEndpoint
	}
	secretAuth, err := vaultClient.Logical().Write(appRoleEndpoint, data)
	if err != nil {
		return "", err
	}
	return secretAuth.Auth.ClientToken, nil
}
