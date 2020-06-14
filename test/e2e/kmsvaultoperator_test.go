package e2e

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	apis "github.com/patoarvizu/kms-vault-operator/pkg/apis"
	operator "github.com/patoarvizu/kms-vault-operator/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const encryptedSecret = "AQICAHgKbLYZWOFlPGwA/1foMoxcBOxv7LddQQW9biqG70YNkwF+dKr15L/4Pl/d26uDd7KqAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMz0gfMT1P5MBTd/fGAgEQgCANG/RycP+0ZXj2qZORafZO4fGdU7KGFINsrs1JDnx1mg=="
const encryptedSecretWithContext = "AQICAHgKbLYZWOFlPGwA/1foMoxcBOxv7LddQQW9biqG70YNkwEfRLaPGqGLTXl/5MT6YX7ZAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMoXRvpmXuVp48k9zYAgEQgCBwncIXMiSO08MWoYp5yXRbn1sflcDPXUt6c6GERNhNOA=="
const encryptedSecretWithContext2 = "AQICAHgKbLYZWOFlPGwA/1foMoxcBOxv7LddQQW9biqG70YNkwHe8HLsPbxG8LglQiSoTR/gAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQM6xvLMMiqXwEstpLkAgEQgCBYB78eEFNHt1QZgFocfnGIJXg+v8W90y0cSnQCqmC/fg=="

func setup(t *testing.T, ctx *test.TestCtx) {
	ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	err := e2eutil.WaitForOperatorDeployment(t, framework.Global.KubeClient, "vault", "kms-vault-operator", 1, time.Second*5, time.Second*60)
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

func createKMSVaultSecret(secrets map[string]string, emptySecret bool, highSecretContext map[string]string, lowSecretContext map[string]string, secretPath string, engineVersion string, finalizers []string, includeSecrets []string, t *testing.T, o *framework.CleanupOptions) *operator.KMSVaultSecret {
	secretsMap := convertToSecretMap(secrets, lowSecretContext, emptySecret)
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
			SecretContext:  highSecretContext,
			IncludeSecrets: includeSecrets,
		},
	}

	err := framework.Global.Client.Create(context.TODO(), secret, o)
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}
	return secret
}

func createPartialKMSVaultSecret(secrets map[string]string, highSecretContext map[string]string, lowSecretContext map[string]string, t *testing.T, o *framework.CleanupOptions) *operator.PartialKMSVaultSecret {
	secretsMap := convertToSecretMap(secrets, lowSecretContext, false)
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
			Secrets:       secretsMap,
			SecretContext: highSecretContext,
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
	vaultSecret, err := framework.Global.KubeClient.CoreV1().Secrets("vault").Get("vault-unseal-keys", metav1.GetOptions{})
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

func TestMonitoringObjectsCreated(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	metricsService := &v1.Service{}
	err := wait.Poll(time.Second*2, time.Second*60, func() (done bool, err error) {
		err = framework.Global.Client.Get(context.TODO(), dynclient.ObjectKey{Namespace: "vault", Name: "kms-vault-operator-metrics"}, metricsService)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		t.Fatal("Could not get metrics Service")
	}
	httpMetricsPortFound := false
	crMetricsPortFound := false
	for _, p := range metricsService.Spec.Ports {
		if p.Name == "http-metrics" && p.Port == 8383 {
			httpMetricsPortFound = true
			continue
		}
		if p.Name == "cr-metrics" && p.Port == 8686 {
			crMetricsPortFound = true
			continue
		}
	}
	if !httpMetricsPortFound {
		t.Fatal("Service kms-vault-operator-metrics doesn't have http-metrics port 8383")
	}
	if !crMetricsPortFound {
		t.Fatal("Service kms-vault-operator-metrics doesn't have cr-metrics port 8686")
	}

	framework.Global.Scheme.AddKnownTypes(monitoringv1.SchemeGroupVersion, &monitoringv1.ServiceMonitor{})
	serviceMonitor := &monitoringv1.ServiceMonitor{}
	err = wait.Poll(time.Second*2, time.Second*60, func() (done bool, err error) {
		err = framework.Global.Client.Client.Get(context.TODO(), dynclient.ObjectKey{Namespace: "vault", Name: "kms-vault-operator-metrics"}, serviceMonitor)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		t.Fatal("Could not find metrics ServiceMonitor")
	}
	httpMetricsEndpointFound := false
	crMetricsEndpointFound := false
	for _, e := range serviceMonitor.Spec.Endpoints {
		if e.Port == "http-metrics" {
			httpMetricsEndpointFound = true
			continue
		}
		if e.Port == "cr-metrics" {
			crMetricsEndpointFound = true
			continue
		}
	}
	if !httpMetricsEndpointFound {
		t.Error("ServiceMonitor kms-vault-operator-metrics doesn't have endpoint http-metrics")
	}
	if !crMetricsEndpointFound {
		t.Error("ServiceMonitor kms-vault-operator-metrics doesn't have endpoint cr-metrics")
	}
	ctx.Cleanup()
}

func TestKMSVaultSecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestUnencryptedSecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": "UnencryptedSecret"}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretDoesntExist(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestUnencryptedSecretV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": "UnencryptedSecret"}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretDoesntExist(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretFinalizersV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{"delete.k8s.patoarvizu.dev"}, []string{}, t, nil)
	validateSecretExists(secret, "Hello", t)
	framework.Global.Client.Delete(context.TODO(), secret)
	validateSecretDoesntExist(secret, "Hello", t)
	ctx.Cleanup()
}

func TestKMSVaultSecretFinalizersV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{"delete.k8s.patoarvizu.dev"}, []string{}, t, nil)
	validateSecretExists(secret, "Hello", t)
	framework.Global.Client.Delete(context.TODO(), secret)
	validateSecretDoesntExist(secret, "Hello", t)
	ctx.Cleanup()
}

func TestKMSVaultSecretWithHighContextV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, map[string]string{"Hello": "World"}, make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretWithHighContextV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, map[string]string{"Hello": "World"}, make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretWithLowContextV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, make(map[string]string), map[string]string{"Hello": "World"}, "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretWithLowContextV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, make(map[string]string), map[string]string{"Hello": "World"}, "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretWithHighAndLowContextV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := &operator.KMSVaultSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: operator.KMSVaultSecretSpec{
			Path: "secret/test-secret",
			KVSettings: operator.KVSettings{
				EngineVersion: "v1",
			},
			Secrets: []operator.Secret{
				{
					Key:             "Hello",
					EncryptedSecret: encryptedSecretWithContext,
				},
				{
					Key:             "Hello2",
					EncryptedSecret: encryptedSecretWithContext2,
					SecretContext:   map[string]string{"Hello2": "World2"},
				},
			},
			SecretContext: map[string]string{"Hello": "World"},
		},
	}
	err := framework.Global.Client.Create(context.TODO(), secret, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}
	validateSecretExists(secret, "Hello", t)
	validateSecretExists(secret, "Hello2", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretWithHighAndLowContextV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := &operator.KMSVaultSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: operator.KMSVaultSecretSpec{
			Path: "secret/data/test-secret",
			KVSettings: operator.KVSettings{
				EngineVersion: "v2",
			},
			Secrets: []operator.Secret{
				{
					Key:             "Hello",
					EncryptedSecret: encryptedSecretWithContext,
				},
				{
					Key:             "Hello2",
					EncryptedSecret: encryptedSecretWithContext2,
					SecretContext:   map[string]string{"Hello2": "World2"},
				},
			},
			SecretContext: map[string]string{"Hello": "World"},
		},
	}
	err := framework.Global.Client.Create(context.TODO(), secret, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}
	validateSecretExists(secret, "Hello", t)
	validateSecretExists(secret, "Hello2", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretSomeFailedKeysV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret, "Failed": "EncryptionError"}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	validateSecretDoesntExist(secret, "Failed", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretSomeFailedKeysV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret, "Failed": "EncryptionError"}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	validateSecretDoesntExist(secret, "Failed", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretPartialSecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	createPartialKMSVaultSecret(map[string]string{"PartialHello": encryptedSecret}, make(map[string]string), make(map[string]string), t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{"partial-secret"}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	validateSecretExists(secret, "PartialHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretPartialSecretV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	createPartialKMSVaultSecret(map[string]string{"PartialHello": encryptedSecret}, make(map[string]string), make(map[string]string), t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	secret := createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{"partial-secret"}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "Hello", t)
	validateSecretExists(secret, "PartialHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretEmptySecretIgnoreValidEncryptedStringV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"EmptyHello": encryptedSecret}, true, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "EmptyHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretEmptySecretIgnoreValidEncryptedStringV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"EmptyHello": encryptedSecret}, true, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "EmptyHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretEmptySecretV1(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"EmptyHello": ""}, true, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "EmptyHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestKMSVaultSecretEmptySecretV2(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secret := createKMSVaultSecret(map[string]string{"EmptyHello": ""}, true, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{}, t, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	validateSecretExists(secret, "EmptyHello", t)
	cleanUpVaultSecret(secret, t)
	ctx.Cleanup()
}

func TestWebhookRejectsBadEncoding(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secretsMap := convertToSecretMap(map[string]string{"Secret": "BadEncoding"}, make(map[string]string), false)
	secret := &operator.KMSVaultSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: operator.KMSVaultSecretSpec{
			Path: "secret/data/test-secret",
			KVSettings: operator.KVSettings{
				EngineVersion: "v1",
			},
			Secrets:       secretsMap,
			SecretContext: map[string]string{"Environment": "Test"},
		},
	}
	err := framework.Global.Client.Create(context.TODO(), secret, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	if err == nil {
		t.Error(fmt.Sprint("Secret with bad encoding should've thrown an error"))
	}
	ctx.Cleanup()
}

func TestWebhookRejectsBadEncryption(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	setup(t, ctx)
	secretsMap := convertToSecretMap(map[string]string{"Secret": encryptedSecret}, make(map[string]string), false)
	secret := &operator.KMSVaultSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: operator.KMSVaultSecretSpec{
			Path: "secret/data/test-secret",
			KVSettings: operator.KVSettings{
				EngineVersion: "v1",
			},
			Secrets:       secretsMap,
			SecretContext: map[string]string{"Environment": "Test"},
		},
	}
	err := framework.Global.Client.Create(context.TODO(), secret, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 60, RetryInterval: time.Second * 1})
	if err == nil {
		t.Error(fmt.Sprint("Secret with bad encryption should've thrown an error"))
	}
	ctx.Cleanup()
}
