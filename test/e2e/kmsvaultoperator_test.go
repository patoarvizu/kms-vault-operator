package e2e

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	apis "github.com/patoarvizu/kms-vault-operator/pkg/apis"
	operator "github.com/patoarvizu/kms-vault-operator/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const encryptedSecret = "AQICAHgKbLYZWOFlPGwA/1foMoxcBOxv7LddQQW9biqG70YNkwF+dKr15L/4Pl/d26uDd7KqAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMz0gfMT1P5MBTd/fGAgEQgCANG/RycP+0ZXj2qZORafZO4fGdU7KGFINsrs1JDnx1mg=="

func setup(t *testing.T, ctx *test.TestCtx) {
	testNamespace, err := ctx.GetNamespace()
	if err != nil {
		t.Error(err)
	}
	awsSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws-secrets",
			Namespace: testNamespace,
		},
		StringData: map[string]string{
			"AWS_ACCESS_KEY_ID":     os.Getenv("AWS_ACCESS_KEY_ID"),
			"AWS_SECRET_ACCESS_KEY": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		},
	}
	framework.Global.Client.Create(context.TODO(), awsSecret, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	err = e2eutil.WaitForOperatorDeployment(t, framework.Global.KubeClient, testNamespace, "kms-vault-operator", 1, time.Second*5, time.Second*60)
	if err != nil {
		t.Fatal(err)
	}
	kmsVaultSecretList := &operator.KMSVaultSecretList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
	}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, kmsVaultSecretList)
	if err != nil {
		t.Fatalf("Failed to add to scheme: %s", err)
	}
}

func createKMSVaultSecret(encryptedText string, finalizers []string, t *testing.T, ctx *test.TestCtx, o *framework.CleanupOptions) *operator.KMSVaultSecret {
	secret := &operator.KMSVaultSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-secret",
			Namespace:  "default",
			Finalizers: finalizers,
		},
		Spec: operator.KMSVaultSecretSpec{
			Path:            "secret/test-secret",
			VaultAuthMethod: "k8s",
			KVSettings: operator.KVSettings{
				EngineVersion: "v1",
			},
			Secrets: []operator.Secret{
				operator.Secret{
					Key:             "Hello",
					EncryptedSecret: encryptedText,
				},
			},
		},
	}

	err := framework.Global.Client.Create(context.TODO(), secret, o)
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}
	return secret
}

func cleanUpVaultSecret(t *testing.T) {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		t.Fatalf("Failed to get Vault client: %v", err)
	}
	vaultClient.Logical().Delete("secret/test-secret")
}

func validateSecretExists(t *testing.T) {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		t.Fatalf("Failed to get Vault client: %v", err)
	}
	err = wait.Poll(time.Second*2, time.Second*30, func() (done bool, err error) {
		r, err := vaultClient.Logical().Read("secret/test-secret")
		if err != nil {
			return false, err
		}
		if r == nil {
			return false, nil
		}
		if val, ok := r.Data["Hello"]; ok {
			if val != "World" {
				return false, errors.New("Encrypted string wasn't decrypted correctly")
			}
		} else {
			return false, errors.New("Secret wasn't successfully put in Vault")
		}
		return true, nil
	})
	if err != nil {
		t.Error(err)
	}
}

func validateSecretDoesntExist(secret *operator.KMSVaultSecret, t *testing.T) {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		t.Fatalf("Failed to get Vault client: %v", err)
	}

	err = wait.Poll(time.Second*2, time.Second*30, func() (done bool, err error) {
		r, err := vaultClient.Logical().Read("secret/test-secret")
		if err != nil {
			return false, err
		}
		if r == nil {
			return true, nil
		} else {
			return false, errors.New("Vault secret should not exist")
		}
	})
	if err != nil {
		t.Error(err)
	}
}

func authenticatedVaultClient() (*vaultapi.Client, error) {
	vaultSecret, err := framework.Global.KubeClient.CoreV1().Secrets("default").Get("vault-unseal-keys", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	vaultClient, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if err != nil {
		return nil, err
	}
	vaultClient.SetToken(string(vaultSecret.Data["vault-root"]))
	vaultClient.Auth()

	return vaultClient, nil
}

func TestKMSVaultSecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)

	createKMSVaultSecret(encryptedSecret, []string{}, t, ctx, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})

	validateSecretExists(t)

	cleanUpVaultSecret(t)

	ctx.Cleanup()
}

func TestUnencryptedSecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)

	secret := createKMSVaultSecret("UnencryptedSecret", []string{}, t, ctx, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})

	validateSecretDoesntExist(secret, t)

	ctx.Cleanup()
}

func TestKMSVaultSecretFinalizersV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)

	secret := createKMSVaultSecret(encryptedSecret, []string{"delete.k8s.patoarvizu.dev"}, t, ctx, nil)

	validateSecretExists(t)

	framework.Global.Client.Delete(context.TODO(), secret)

	validateSecretDoesntExist(secret, t)

	ctx.Cleanup()
}
