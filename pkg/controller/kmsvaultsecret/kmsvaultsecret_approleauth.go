package kmsvaultsecret

import (
	"errors"
	"os"
)

const (
	appRoleDefaultEndpoint = "auth/approle/login"
)

type VaultAppRoleAuth struct{}

func (auth VaultAppRoleAuth) login() error {
	roleId, ok := os.LookupEnv("VAULT_APPROLE_ROLE_ID")
	if !ok {
		return errors.New("Environment variable VAULT_APPROLE_ROLE_ID not set")
	}
	secretId, ok := os.LookupEnv("VAULT_APPROLE_SECRET_ID")
	if !ok {
		return errors.New("Environment variable VAULT_APPROLE_SECRET_ID not set")
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
		return err
	}
	vaultClient.SetToken(secretAuth.Auth.ClientToken)
	return nil
}
