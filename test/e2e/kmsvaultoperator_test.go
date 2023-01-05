package e2e

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	k8sv1alpha1 "github.com/patoarvizu/kms-vault-operator/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

const encryptedSecret = "AQICAHgKbLYZWOFlPGwA/1foMoxcBOxv7LddQQW9biqG70YNkwF+dKr15L/4Pl/d26uDd7KqAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMz0gfMT1P5MBTd/fGAgEQgCANG/RycP+0ZXj2qZORafZO4fGdU7KGFINsrs1JDnx1mg=="
const encryptedSecretWithContext = "AQICAHgKbLYZWOFlPGwA/1foMoxcBOxv7LddQQW9biqG70YNkwEfRLaPGqGLTXl/5MT6YX7ZAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMoXRvpmXuVp48k9zYAgEQgCBwncIXMiSO08MWoYp5yXRbn1sflcDPXUt6c6GERNhNOA=="
const encryptedSecretWithContext2 = "AQICAHgKbLYZWOFlPGwA/1foMoxcBOxv7LddQQW9biqG70YNkwHe8HLsPbxG8LglQiSoTR/gAAAAYzBhBgkqhkiG9w0BBwagVDBSAgEAME0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQM6xvLMMiqXwEstpLkAgEQgCBYB78eEFNHt1QZgFocfnGIJXg+v8W90y0cSnQCqmC/fg=="

func createKMSVaultSecret(secrets map[string]string, emptySecret bool, highSecretContext map[string]string, lowSecretContext map[string]string, secretPath string, engineVersion string, finalizers []string, includeSecrets []string) *k8sv1alpha1.KMSVaultSecret {
	secretsMap := convertToSecretMap(secrets, lowSecretContext, emptySecret)
	secret := &k8sv1alpha1.KMSVaultSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-secret",
			Namespace:  "default",
			Finalizers: finalizers,
		},
		Spec: k8sv1alpha1.KMSVaultSecretSpec{
			Path: secretPath,
			KVSettings: k8sv1alpha1.KVSettings{
				EngineVersion: engineVersion,
			},
			Secrets:        secretsMap,
			SecretContext:  highSecretContext,
			IncludeSecrets: includeSecrets,
		},
	}

	err := k8sClient.Create(context.TODO(), secret)
	if err != nil {
		return nil
	}
	return secret
}

func createPartialKMSVaultSecret(secrets map[string]string, highSecretContext map[string]string, lowSecretContext map[string]string) *k8sv1alpha1.PartialKMSVaultSecret {
	secretsMap := convertToSecretMap(secrets, lowSecretContext, false)
	partialSecret := &k8sv1alpha1.PartialKMSVaultSecret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PartialKMSVaultSecret",
			APIVersion: "k8s.patoarvizu.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "partial-secret",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.PartialKMSVaultSecretSpec{
			Secrets:       secretsMap,
			SecretContext: highSecretContext,
		},
	}
	err := k8sClient.Create(context.TODO(), partialSecret)
	if err != nil {
		return nil
	}
	return partialSecret
}

func convertToSecretMap(secrets map[string]string, secretContext map[string]string, emptySecret bool) []k8sv1alpha1.Secret {
	s := []k8sv1alpha1.Secret{}
	for k, v := range secrets {
		s = append(s, k8sv1alpha1.Secret{Key: k, EncryptedSecret: v, SecretContext: secretContext, EmptySecret: emptySecret})
	}
	return s
}

func cleanUpVaultSecret(secret *k8sv1alpha1.KMSVaultSecret) error {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		return err
	}
	deletePath := strings.Replace(secret.Spec.Path, "secret/data/", "secret/metadata/", 1)
	_, err = vaultClient.Logical().Delete(deletePath)
	k8sClient.Delete(context.TODO(), secret)
	return err
}

func authenticatedVaultClient() (*vaultapi.Client, error) {
	vaultSecret := &v1.Secret{}
	err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "vault", Name: "vault-unseal-keys"}, vaultSecret)
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

func validateSecretExists(secret *k8sv1alpha1.KMSVaultSecret, key string) error {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		return err
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
	return err
}

func validateSecretDoesntExist(secret *k8sv1alpha1.KMSVaultSecret, key string) error {
	vaultClient, err := authenticatedVaultClient()
	if err != nil {
		return err
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
	return err
}

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:  []string{filepath.Join("..", "..", "config", "crd", "bases")},
		UseExistingCluster: newTrue(),
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = k8sv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = Describe("With K/V v1 secrets", func() {
	var (
		secret        *k8sv1alpha1.KMSVaultSecret
		partialSecret *k8sv1alpha1.PartialKMSVaultSecret
		err           error
	)

	Context("When a valid KMSVaultSecret is created", func() {
		It("Should be injected into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When an invalid KMSVaultSecret is created", func() {
		It("Should not be injected into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": "UnencryptedSecret"}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretDoesntExist(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When a KMSVaultSecret with a finalizer is created", func() {
		It("Should be injected into Vault but removed when it's deleted from Kubernetes", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{"delete.k8s.patoarvizu.dev"}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			k8sClient.Delete(context.TODO(), secret)
			err = validateSecretDoesntExist(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When a KMSVaultSecret with an encryption context at the top level is created", func() {
		It("Should be injected into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, map[string]string{"Hello": "World"}, make(map[string]string), "secret/test-secret", "v1", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When a KMSVaultSecret with an encryption context at a lower level is created", func() {
		It("Should be injected into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, make(map[string]string), map[string]string{"Hello": "World"}, "secret/test-secret", "v1", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When a KMSVaultSecret with both a high and a low encryption context is created", func() {
		It("Should be injected into Vault", func() {
			secret = &k8sv1alpha1.KMSVaultSecret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KMSVaultSecret",
					APIVersion: "k8s.patoarvizu.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Spec: k8sv1alpha1.KMSVaultSecretSpec{
					Path: "secret/test-secret",
					KVSettings: k8sv1alpha1.KVSettings{
						EngineVersion: "v1",
					},
					Secrets: []k8sv1alpha1.Secret{
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
			err = k8sClient.Create(context.TODO(), secret)
			Expect(err).ToNot(HaveOccurred())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = validateSecretExists(secret, "Hello2")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("If a KMSVaultSecret is created with some keys that failed to decrypt", func() {
		It("Should still inject the valid keys into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecret, "Failed": "EncryptionError"}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = validateSecretDoesntExist(secret, "Failed")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("If a PartialKMSVaultSecret is created", func() {
		It("Can be included in a full KMSVaultSecret and get injected into Vault", func() {
			partialSecret = createPartialKMSVaultSecret(map[string]string{"PartialHello": encryptedSecret}, make(map[string]string), make(map[string]string))
			Expect(partialSecret).ToNot(BeNil())
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{"partial-secret"})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = validateSecretExists(secret, "PartialHello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Delete(context.TODO(), partialSecret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("If a KMSVaultSecret is created with a a valid encrypted string but with 'emptySecret' set to true", func() {
		It("Should ignore the encrypted string and inject the secret as empty", func() {
			secret = createKMSVaultSecret(map[string]string{"EmptyHello": encryptedSecret}, true, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "EmptyHello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("If a KMSVaultSecret is created with 'emptySecret' set to true", func() {
		It("Should inject an empty secret into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"EmptyHello": ""}, true, make(map[string]string), make(map[string]string), "secret/test-secret", "v1", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "EmptyHello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("With K/V v2 secrets", func() {
	var (
		secret        *k8sv1alpha1.KMSVaultSecret
		partialSecret *k8sv1alpha1.PartialKMSVaultSecret
		err           error
	)

	Context("When a valid KMSVaultSecret is created", func() {
		It("Should be injected into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When an invalid KMSVaultSecret is created", func() {
		It("Should not be injected into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": "UnencryptedSecret"}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretDoesntExist(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When a KMSVaultSecret with a finalizer is created", func() {
		It("Should be injected into Vault but removed when it's deleted from Kubernetes", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{"delete.k8s.patoarvizu.dev"}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			k8sClient.Delete(context.TODO(), secret)
			err = validateSecretDoesntExist(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When a KMSVaultSecret with an encryption context at the top level is created", func() {
		It("Should be injected into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, map[string]string{"Hello": "World"}, make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When a KMSVaultSecret with an encryption context at a lower level is created", func() {
		It("Should be injected into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecretWithContext}, false, make(map[string]string), map[string]string{"Hello": "World"}, "secret/data/test-secret", "v2", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When a KMSVaultSecret with both a high and a low encryption context is created", func() {
		It("Should be injected into Vault", func() {
			secret = &k8sv1alpha1.KMSVaultSecret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KMSVaultSecret",
					APIVersion: "k8s.patoarvizu.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Spec: k8sv1alpha1.KMSVaultSecretSpec{
					Path: "secret/data/test-secret",
					KVSettings: k8sv1alpha1.KVSettings{
						EngineVersion: "v2",
					},
					Secrets: []k8sv1alpha1.Secret{
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
			err = k8sClient.Create(context.TODO(), secret)
			Expect(err).ToNot(HaveOccurred())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = validateSecretExists(secret, "Hello2")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("If a KMSVaultSecret is created with some keys that failed to decrypt", func() {
		It("Should still inject the valid keys into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecret, "Failed": "EncryptionError"}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = validateSecretDoesntExist(secret, "Failed")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("If a PartialKMSVaultSecret is created", func() {
		It("Can be included in a full KMSVaultSecret and get injected into Vault", func() {
			partialSecret = createPartialKMSVaultSecret(map[string]string{"PartialHello": encryptedSecret}, make(map[string]string), make(map[string]string))
			Expect(partialSecret).ToNot(BeNil())
			secret = createKMSVaultSecret(map[string]string{"Hello": encryptedSecret}, false, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{"partial-secret"})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "Hello")
			Expect(err).ToNot(HaveOccurred())
			err = validateSecretExists(secret, "PartialHello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Delete(context.TODO(), partialSecret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("If a KMSVaultSecret is created with a a valid encrypted string but with 'emptySecret' set to true", func() {
		It("Should ignore the encrypted string and inject the secret as empty", func() {
			secret = createKMSVaultSecret(map[string]string{"EmptyHello": encryptedSecret}, true, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "EmptyHello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("If a KMSVaultSecret is created with 'emptySecret' set to true", func() {
		It("Should inject an empty secret into Vault", func() {
			secret = createKMSVaultSecret(map[string]string{"EmptyHello": ""}, true, make(map[string]string), make(map[string]string), "secret/data/test-secret", "v2", []string{}, []string{})
			Expect(secret).ToNot(BeNil())
			err = validateSecretExists(secret, "EmptyHello")
			Expect(err).ToNot(HaveOccurred())
			err = cleanUpVaultSecret(secret)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("With webhook", func() {
	var (
		secret *k8sv1alpha1.KMSVaultSecret
		err    error
	)

	Context("When a secret is incorrectly encoded", func() {
		It("Should be rejected by the webhook", func() {
			secretsMap := convertToSecretMap(map[string]string{"Secret": "BadEncoding"}, make(map[string]string), false)
			Expect(secretsMap).ToNot(BeNil())
			secret = &k8sv1alpha1.KMSVaultSecret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KMSVaultSecret",
					APIVersion: "k8s.patoarvizu.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Spec: k8sv1alpha1.KMSVaultSecretSpec{
					Path: "secret/data/test-secret",
					KVSettings: k8sv1alpha1.KVSettings{
						EngineVersion: "v1",
					},
					Secrets:       secretsMap,
					SecretContext: map[string]string{"Environment": "Test"},
				},
			}
			err = k8sClient.Create(context.TODO(), secret)
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When a secret is incorrectly encrypted", func() {
		It("Should be rejected by the webhook", func() {
			secretsMap := convertToSecretMap(map[string]string{"Secret": encryptedSecret}, make(map[string]string), false)
			Expect(secretsMap).ToNot(BeNil())
			secret = &k8sv1alpha1.KMSVaultSecret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KMSVaultSecret",
					APIVersion: "k8s.patoarvizu.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Spec: k8sv1alpha1.KMSVaultSecretSpec{
					Path: "secret/data/test-secret",
					KVSettings: k8sv1alpha1.KVSettings{
						EngineVersion: "v1",
					},
					Secrets:       secretsMap,
					SecretContext: map[string]string{"Environment": "Test"},
				},
			}
			err = k8sClient.Create(context.TODO(), secret)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func newTrue() *bool {
	b := true
	return &b
}
