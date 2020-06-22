package kmsvaultsecret

import (
	"errors"
	"os"
)

type VaultTokenAuth struct{}

func (auth VaultTokenAuth) login() error {
	vaultToken, set := os.LookupEnv("VAULT_TOKEN")
	if !set {
		return errors.New("VAULT_TOKEN environment variable not found")
	}
	vaultClient.SetToken(vaultToken)
	return nil
}
