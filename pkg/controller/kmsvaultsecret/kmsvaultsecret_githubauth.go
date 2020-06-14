package kmsvaultsecret

import (
	"errors"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

const (
	gitHubAuthDefaultEndpoint = "auth/github/login"
)

type VaultGitHubAuth struct{}

func (k8s VaultGitHubAuth) login(vaultConfig *vaultapi.Config) (string, error) {
	githubToken, ok := os.LookupEnv("VAULT_GITHUB_TOKEN")
	if !ok {
		return "", errors.New("Environment variable VAULT_GITHUB_TOKEN not set")
	}
	vaultClient, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"token": githubToken,
	}
	githubAuthEndpoint, ok := os.LookupEnv("VAULT_GITHUB_AUTH_ENDPOINT")
	if !ok {
		githubAuthEndpoint = gitHubAuthDefaultEndpoint
	}
	secretAuth, err := vaultClient.Logical().Write(githubAuthEndpoint, data)
	if err != nil {
		return "", err
	}
	return secretAuth.Auth.ClientToken, nil
}
