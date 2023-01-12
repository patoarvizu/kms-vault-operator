/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/go-logr/logr"
	"github.com/radovskyb/watcher"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	vaultapi "github.com/hashicorp/vault/api"
	k8sv1alpha1 "github.com/patoarvizu/kms-vault-operator/api/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type VaultAuthMethod interface {
	login() error
}

func renewToken(m VaultAuthMethod) error {
	tokenLookup, err := vaultClient.Auth().Token().LookupSelf()
	if err == nil {
		expiration := tokenLookup.Data["expire_time"]
		t, err := time.Parse(time.RFC3339, expiration.(string))
		if err == nil {
			now := time.Now()
			if t.After(now) {
				return nil
			}
			renewable, _ := tokenLookup.TokenIsRenewable()
			if renewable {
				vaultClient.Auth().Token().RenewSelf(0)
				return nil
			}
		}
	}
	return m.login()
}

func watchCertificate() {
	logger := log.WithValues("Function", "WatchCertificate")
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)
	watchedCACert := os.Getenv("VAULT_CACERT")
	if len(watchedCACert) > 0 {
		_ = w.Add(watchedCACert)
	}
	watchedCAPath := os.Getenv("VAULT_CAPATH")
	if len(watchedCAPath) > 0 {
		_ = w.Add(watchedCAPath)
	}
	go func() {
		for {
			select {
			case <-w.Event:
				logger.Info("Updating CA certificate for client")
				err := setVaultClient()
				if err != nil {
					logger.Error(err, "Error refreshing client")
					os.Exit(1)
				}
			}
		}
	}()
	go w.Start(time.Millisecond * 100)
}

type KVWriter interface {
	write(*k8sv1alpha1.KMSVaultSecret, *vaultapi.Client) error
	delete(*k8sv1alpha1.KMSVaultSecret, *vaultapi.Client) error
}

const (
	K8sAuthenticationMethod      string = "k8s"
	TokenAuthenticationMethod    string = "token"
	UserpassAuthenticationMethod string = "userpass"
	AppRoleAuthenticationMethod  string = "approle"
	GitHubAuthenticationMethod   string = "github"
	AWSIAMAuthenticationMethod   string = "iam"
	KVv1                         string = "v1"
	KVv2                         string = "v2"
	DeletedFinalizer             string = "delete.k8s.patoarvizu.dev"
)

var log = logf.Log.WithName("controller_kmsvaultsecret")
var rec record.EventRecorder
var reqLogger logr.Logger
var vaultClient *vaultapi.Client
var vaultAuthMethod VaultAuthMethod

func setVaultClient() error {
	c, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if err != nil {
		return err
	}
	vaultClient = c
	return nil
}

// KMSVaultSecretReconciler reconciles a KMSVaultSecret object
type KMSVaultSecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=k8s.patoarvizu.dev,resources=kmsvaultsecrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=k8s.patoarvizu.dev,resources=partialkmsvaultsecrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=k8s.patoarvizu.dev,resources=kmsvaultsecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create

func (r *KMSVaultSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling KMSVaultSecret")

	instance := &k8sv1alpha1.KMSVaultSecret{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	for _, partialSecretName := range instance.Spec.IncludeSecrets {
		partialSecretInstance := &k8sv1alpha1.PartialKMSVaultSecret{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: partialSecretName}, partialSecretInstance)
		if err != nil {
			reqLogger.Info(fmt.Sprintf("Error getting included secret %s, skipping it...", partialSecretName))
			continue
		}
		instance.Spec.Secrets = append(instance.Spec.Secrets, partialSecretInstance.Spec.Secrets...)
	}

	err = renewToken(vaultAuthMethod)
	if err != nil {
		reqLogger.Error(err, "Error getting authenticated Vault client")
		return reconcile.Result{RequeueAfter: time.Second * 15}, err
	}

	writer := kvWriter(instance.Spec.KVSettings.EngineVersion)
	if instance.ObjectMeta.DeletionTimestamp != nil {
		reqLogger.Info("Resource deleted, cleaning up")
		err = writer.delete(instance, vaultClient)
		if err != nil {
			reqLogger.Error(err, "Error deleting secret from Vault")
			return reconcile.Result{}, err
		}
		instance.Finalizers = removeFinalizer(instance.Finalizers, DeletedFinalizer)
		r.Client.Update(ctx, instance)
		return reconcile.Result{}, nil
	}

	err = writer.write(instance, vaultClient)
	if err != nil {
		reqLogger.Error(err, "Error writing secret to Vault")
		return reconcile.Result{RequeueAfter: time.Second * 15}, err
	}
	if !instance.Status.Created {
		instance.Status.Created = true
		rec.Event(instance, corev1.EventTypeNormal, "SecretCreated", fmt.Sprintf("Wrote secret %s to %s", instance.Name, instance.Spec.Path))
		r.Client.Status().Update(ctx, instance)
	}
	return reconcile.Result{RequeueAfter: time.Second * time.Duration(SyncPeriodSeconds)}, nil
}

func (r *KMSVaultSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	rec = mgr.GetEventRecorderFor("kms-vault-controller")
	err := setVaultClient()
	if err != nil {
		return err
	}
	watchCertificate()
	vaultAuthMethod = vaultAuthentication(VaultAuthenticationMethod)
	err = vaultAuthMethod.login()
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1alpha1.KMSVaultSecret{}).
		Complete(r)
}

func removeFinalizer(allFinalizers []string, finalizer string) []string {
	var result []string
	for _, f := range allFinalizers {
		if f == finalizer {
			continue
		}
		result = append(result, f)
	}
	return result
}

func decryptSecrets(secret *k8sv1alpha1.KMSVaultSecret) (map[string]interface{}, error) {
	logger := log.WithValues("Function", "decryptSecrets")
	awsSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	decryptedSecretData := map[string]interface{}{}
	svc := kms.New(awsSession)
	for _, s := range secret.Spec.Secrets {
		if s.EmptySecret {
			if len(s.EncryptedSecret) > 0 {
				logger.Info("Secret is marked as empty, ignoring content", "secretKey", s.Key, "encodedString", s.EncryptedSecret)
			}
			decryptedSecretData[s.Key] = ""
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(s.EncryptedSecret)
		if err != nil {
			logger.Info("Error decoding secret, skipping", "secretKey", s.Key, "encodedString", s.EncryptedSecret)
			rec.Event(secret, corev1.EventTypeWarning, "DecodingError", fmt.Sprintf("Error decoding key %s", s.Key))
			continue
		}
		result, err := svc.Decrypt(&kms.DecryptInput{CiphertextBlob: decoded, EncryptionContext: getApplicableContext(s.SecretContext, secret.Spec.SecretContext)})
		if err != nil {
			logger.Info("Error decrypting secret, skipping", "secretKey", s.Key, "encodedString", s.EncryptedSecret)
			rec.Event(secret, corev1.EventTypeWarning, "DecryptingError", fmt.Sprintf("Error decrypting key %s", s.Key))
			continue
		}
		decryptedSecretData[s.Key] = string(result.Plaintext)
	}
	return decryptedSecretData, nil
}

func getApplicableContext(lowerContext map[string]string, higherContext map[string]string) map[string]*string {
	if len(lowerContext) > 0 {
		return convertContextMap(lowerContext)
	} else {
		return convertContextMap(higherContext)
	}
}

func convertContextMap(context map[string]string) map[string]*string {
	m := make(map[string]*string)
	for k, v := range context {
		m[k] = &v
	}
	return m
}

func vaultAuthentication(vaultAuthenticationMethod string) VaultAuthMethod {
	switch vaultAuthenticationMethod {
	case K8sAuthenticationMethod:
		return VaultK8sAuth{}
	case UserpassAuthenticationMethod:
		return VaultUserpassAuth{}
	case AppRoleAuthenticationMethod:
		return VaultAppRoleAuth{}
	case GitHubAuthenticationMethod:
		return VaultGitHubAuth{}
	case AWSIAMAuthenticationMethod:
		return VaultIAMAuth{}
	default:
		return VaultTokenAuth{}
	}
}

func kvWriter(kvVersion string) KVWriter {
	switch kvVersion {
	case KVv2:
		return KVv2Writer{}
	default:
		return KVv1Writer{}
	}
}
