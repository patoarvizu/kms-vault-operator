package kmsvaultsecret

import (
	"errors"
	"os"
)

const (
	userpassLoginEndpoint = "auth/userpass/login"
)

type VaultUserpassAuth struct{}

func (auth VaultUserpassAuth) login() error {
	vaultUsername, usernameSet := os.LookupEnv("VAULT_USERNAME")
	if !usernameSet {
		return errors.New("Environment variable VAULT_USERNAME not set")
	}
	vaultPassword, passwordSet := os.LookupEnv("VAULT_PASSWORD")
	if !passwordSet {
		return errors.New("Environment variable VAULT_PASSWORD not set")
	}
	data := map[string]interface{}{
		"username": vaultUsername,
		"password": vaultPassword,
	}
	secretAuth, err := vaultClient.Logical().Write(userpassLoginEndpoint, data)
	if err != nil {
		return err
	}
	vaultClient.SetToken(secretAuth.Auth.ClientToken)
	return nil
}
