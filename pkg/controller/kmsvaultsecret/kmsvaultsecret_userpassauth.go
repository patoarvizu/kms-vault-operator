package kmsvaultsecret

import (
	"errors"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

const (
	userpassLoginEndpoint = "auth/userpass/login"
)

type VaultUserpassAuth struct{}

func (k8s VaultUserpassAuth) login(vaultConfig *vaultapi.Config) (string, error) {
	vaultUsername, usernameSet := os.LookupEnv("VAULT_USERNAME")
	if !usernameSet {
		return "", errors.New("Environment variable VAULT_USERNAME not set")
	}
	vaultPassword, passwordSet := os.LookupEnv("VAULT_PASSWORD")
	if !passwordSet {
		return "", errors.New("Environment variable VAULT_PASSWORD not set")
	}
	vaultClient, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"username": vaultUsername,
		"password": vaultPassword,
	}
	secretAuth, err := vaultClient.Logical().Write(userpassLoginEndpoint, data)
	if err != nil {
		return "", err
	}
	return secretAuth.Auth.ClientToken, nil
}
