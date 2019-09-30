package e2e

import (
	"context"
	"errors"
	"os"
	"strings"
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
const decryptedSecret = "World"
const encryptedSecretWithContext = "AQICAHgKbLYZWOFlPGwA/1foMoxcBOxv7LddQQW9biqG70YNkwEfRLaPGqGLTXl/5MT6YX7ZAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMoXRvpmXuVp48k9zYAgEQgCBwncIXMiSO08MWoYp5yXRbn1sflcDPXUt6c6GERNhNOA=="
const decryptedSecretWithContext = "World"

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

func createKMSVaultSecret(secrets map[string]string, emptySecret bool, secretContext map[string]string, secretPath string, engineVersion string, finalizers []string, includeSecrets []string, t *testing.T, o *framework.CleanupOptions) *operator.KMSVaultSecret {
	secretsMap := convertToSecretMap(secrets, secretContext, emptySecret)
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
			Path: secretPath,
			KVSettings: operator.KVSettings{
				EngineVersion: engineVersion,
			},
			Secrets:        secretsMap,
			IncludeSecrets: includeSecrets,
		},
	}

	err := framework.Global.Client.Create(context.TODO(), secret, o)
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}
	return secret
}

func createPartialKMSVaultSecret(secrets map[string]string, secretContext map[string]string, t *testing.T, o *framework.CleanupOptions) *operator.PartialKMSVaultSecret {
	secretsMap := convertToSecretMap(secrets, secretContext, false)
	partialSecret := &operator.PartialKMSVaultSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PartialKMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "partial-secret",
			Namespace: "default",
		},
		Spec: operator.PartialKMSVaultSecretSpec{
			Secrets: secretsMap,
		},
	}
	err := framework.Global.Client.Create(context.TODO(), partialSecret, o)
	if err != nil {
		t.Fatalf("Failed to create partial secret: %v", err)
	}
	return partialSecret
}

func convertToSecretMap(secrets map[string]string, secretContext map[string]string, emptySecret bool) []operator.Secret {
	s := []operator.Secret{}
	for k, v := range secrets {
		s = append(s, operator.Secret{Key: k, EncryptedSecret: v, SecretContext: secretContext, EmptySecret: emptySecret})
	}
	return s
}

func cleanUpVaultSecret(secret *operator.KMSVaultSecret, t *testing.T) {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		t.Fatalf("Failed to get Vault client: %v", err)
	}
	deletePath := strings.Replace(secret.Spec.Path, "secret/data/", "secret/metadata/", 1)
	_, err = vaultClient.Logical().Delete(deletePath)
}

func validateSecretExists(secret *operator.KMSVaultSecret, key string, t *testing.T) {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		t.Fatalf("Failed to get Vault client: %v", err)
	}
	err = wait.Poll(time.Second*2, time.Second*60, func() (done bool, err error) {
		secretPath := secret.Spec.Path
		r, err := vaultClient.Logical().Read(secretPath)
		if err != nil {
			return false, err
		}
		if r == nil {
			return false, nil
		}
		var vaultData map[string]interface{}
		if secret.Spec.KVSettings.EngineVersion == "v1" {
			vaultData = r.Data
		} else {
			vaultData = r.Data["data"].(map[string]interface{})
		}
		for _, s := range secret.Spec.Secrets {
			if s.Key == key {
				if val, ok := vaultData[s.Key]; ok {
					if s.EmptySecret {
						if val != "" {
							return false, errors.New("Secret should be empty but it contains a value")
						}
						return true, nil
					}
					if val != "World" {
						return false, errors.New("Encrypted string wasn't decrypted correctly")
					}
				} else {
					return false, errors.New("Secret wasn't successfully put in Vault")
				}
			}
		}
		return true, nil
	})
	if err != nil {
		t.Error(err)
	}
}

func validateSecretDoesntExist(secret *operator.KMSVaultSecret, key string, t *testing.T) {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		t.Fatalf("Failed to get Vault client: %v", err)
	}
	err = wait.Poll(time.Second*2, time.Second*60, func() (done bool, err error) {
		r, err := vaultClient.Logical().Read(secret.Spec.Path)
		if err != nil {
			return false, err
		}
		var vaultData map[string]interface{}
		if r != nil {
			if secret.Spec.KVSettings.EngineVersion == "v1" {
				vaultData = r.Data
			} else {
				vaultData = r.Data["data"].(map[string]interface{})
			}
			if len(vaultData) == 0 {
				return true, nil
			}
		}
		if _, ok := vaultData[key]; ok {
			return true, errors.New("Secret should not be in Vault")
		}
		return true, nil
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
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestUnencryptedSecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": "UnencryptedSecret"}, false, make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretDoesntExist(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestUnencryptedSecretV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": "UnencryptedSecret"}, false, make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretDoesntExist(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretFinalizersV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), "secret/test-secret", "v1", []string{"delete.k8s.patoarvizu.dev"}, []string{}, t, nil)
	validateSecretExists(secret, "Hello", t)
	framework.Global.Client.Delete(context.TODO(), secret)
	validateSecretDoesntExist(secret, "Hello", t)
	ctx.Cleanup()
}

func TestKMSVaultSecretFinalizersV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), "secret/data/test-secret", "v2", []string{"delete.k8s.patoarvizu.dev"}, []string{}, t, nil)
	validateSecretExists(secret, "Hello", t)
	framework.Global.Client.Delete(context.TODO(), secret)
	validateSecretDoesntExist(secret, "Hello", t)
	ctx.Cleanup()
}

func TestKMSVaultSecretWithContextV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, map[string]string{"Hello": "World"}, "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretWithContextV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, map[string]string{"Hello": "World"}, "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretSomeFailedKeysV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret, "Failed": "EncryptionError"}, false, make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	validateSecretDoesntExist(secret, "Failed", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretSomeFailedKeysV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret, "Failed": "EncryptionError"}, false, make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	validateSecretDoesntExist(secret, "Failed", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretPartialSecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	createPartialKMSVaultSecret(map[string]string{"PartialHello": encryptedSecret}, make(map[string]string), t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), "secret/test-secret", "v1", []string{}, []string{"partial-secret"}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	validateSecretExists(secret, "PartialHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretPartialSecretV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	createPartialKMSVaultSecret(map[string]string{"PartialHello": encryptedSecret}, make(map[string]string), t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{"partial-secret"}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	validateSecretExists(secret, "PartialHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretEmptySecretIgnoreValidEncryptedStringV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"EmptyHello": encryptedSecret}, true, make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "EmptyHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretEmptySecretIgnoreValidEncryptedStringV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"EmptyHello": encryptedSecret}, true, make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "EmptyHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretEmptySecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"EmptyHello": ""}, true, make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "EmptyHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretEmptySecretV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"EmptyHello": ""}, true, make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "EmptyHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}
