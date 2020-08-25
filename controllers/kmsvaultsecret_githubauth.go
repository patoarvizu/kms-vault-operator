package controllers

import (
	"errors"
	"os"
)

const (
	gitHubAuthDefaultEndpoint = "auth/github/login"
)

type VaultGitHubAuth struct{}

func (auth VaultGitHubAuth) login() error {
	githubToken, ok := os.LookupEnv("VAULT_GITHUB_TOKEN")
	if !ok {
		return errors.New("Environment variable VAULT_GITHUB_TOKEN not set")
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
		return err
	}
	vaultClient.SetToken(secretAuth.Auth.ClientToken)
	return nil
}
