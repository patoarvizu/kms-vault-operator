package kmsvaultsecret

import (
	"io/ioutil"
	"os"
)

type VaultK8sAuth struct{}

func (auth VaultK8sAuth) login() error {
	var vaultK8sRole string
	vaultK8sRole, roleSet := os.LookupEnv("VAULT_K8S_ROLE")
	if !roleSet {
		vaultK8sRole = "kms-vault-operator"
	}
	var vaultK8sLoginEndpoint string
	vaultK8sLoginEndpoint, endpointSet := os.LookupEnv("VAULT_K8S_LOGIN_ENDPOINT")
	if !endpointSet {
		vaultK8sLoginEndpoint = "auth/kubernetes/login"
	}
	vaultToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"jwt":  string(vaultToken),
		"role": vaultK8sRole,
	}
	secretAuth, err := vaultClient.Logical().Write(vaultK8sLoginEndpoint, data)
	if err != nil {
		return err
	}
	vaultClient.SetToken(secretAuth.Auth.ClientToken)
	return nil
}
