package controllers

import (
	"errors"
	"os"

	awsauth "github.com/hashicorp/vault/builtin/credential/aws"
)

const (
	iamAuthDefaultEndpoint = "auth/aws/login"
)

type VaultIAMAuth struct{}

func (auth VaultIAMAuth) login() error {
	logger := log.WithValues("Auth", "IAM")
	authIAMAWSAccessKeyId, ok := os.LookupEnv("VAULT_IAM_AWS_ACCESS_KEY_ID")
	if !ok {
		logger.Info("Environment variable VAULT_IAM_AWS_ACCESS_KEY_ID not set, using default credential chain")
	}
	authIAMAWSSecretAccessKey, ok := os.LookupEnv("VAULT_IAM_AWS_SECRET_ACCESS_KEY")
	if !ok {
		logger.Info("Environment variable VAULT_IAM_AWS_SECRET_ACCESS_KEY not set, using default credential chain")
	}
	authIAMRole, ok := os.LookupEnv("VAULT_IAM_ROLE")
	if !ok {
		logger.Info("Environment variable VAULT_IAM_ROLE not found, Vault will try to guess the role name")
	}
	iamAuthEndpoint, ok := os.LookupEnv("VAULT_IAM_AUTH_ENDPOINT")
	if !ok {
		iamAuthEndpoint = iamAuthDefaultEndpoint
	}
	credentials, err := awsauth.RetrieveCreds(authIAMAWSAccessKeyId, authIAMAWSSecretAccessKey, "")
	if err != nil {
		return err
	}
	// TODO: Support passing header value
	loginData, err := awsauth.GenerateLoginData(credentials, "", "")
	if err != nil {
		return err
	}
	if loginData == nil {
		return errors.New("Couldn't generate IAM login data")
	}
	loginData["role"] = authIAMRole
	secretAuth, err := vaultClient.Logical().Write(iamAuthEndpoint, loginData)
	if err != nil {
		return err
	}
	vaultClient.SetToken(secretAuth.Auth.ClientToken)
	return nil
}
