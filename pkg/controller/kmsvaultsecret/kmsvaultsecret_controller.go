package kmsvaultsecret

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"k8s.io/client-go/tools/record"

	k8sv1alpha1 "github.com/patoarvizu/kms-vault-operator/pkg/apis/k8s/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"

	vaultapi "github.com/hashicorp/vault/api"
)

type VaultAuthMethod interface {
	login(*vaultapi.Config) (string, error)
}

type KVWriter interface {
	write(*k8sv1alpha1.KMSVaultSecret, *vaultapi.Client) error
	delete(*k8sv1alpha1.KMSVaultSecret, *vaultapi.Client) error
}

const (
	K8sAuthenticationMethod      string = "k8s"
	TokenAuthenticationMethod    string = "token"
	UserpassAuthenticationMethod string = "userpass"
	KVv1                         string = "v1"
	KVv2                         string = "v2"
	DeletedFinalizer             string = "delete.k8s.patoarvizu.dev"
)

var log = logf.Log.WithName("controller_kmsvaultsecret")
var rec record.EventRecorder
var reqLogger logr.Logger

func Add(mgr manager.Manager) error {
	rec = mgr.GetRecorder("kms-vault-controller")
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKMSVaultSecret{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("kmsvaultsecret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &k8sv1alpha1.KMSVaultSecret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileKMSVaultSecret{}

type ReconcileKMSVaultSecret struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileKMSVaultSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KMSVaultSecret")

	instance := &k8sv1alpha1.KMSVaultSecret{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	vaultClient, err := getAuthenticatedVaultClient(instance.Spec.VaultAuthMethod)
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
		r.client.Update(context.TODO(), instance)
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
		r.client.Status().Update(context.TODO(), instance)
	}
	return reconcile.Result{RequeueAfter: time.Minute * 2}, nil
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

func decryptSecrets(secrets []k8sv1alpha1.Secret) (map[string]interface{}, error) {
	logger := log.WithValues("Function", "decryptSecrets")
	awsSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	decryptedSecretData := map[string]interface{}{}
	svc := kms.New(awsSession)
	for _, s := range secrets {
		decoded, err := base64.StdEncoding.DecodeString(s.EncryptedSecret)
		if err != nil {
			logger.Info("Error decoding secret, skipping", "secretKey", s.Key, "encodedString", s.EncryptedSecret)
			continue
		}
		result, err := svc.Decrypt(&kms.DecryptInput{CiphertextBlob: decoded, EncryptionContext: convertContextMap(s.SecretContext)})
		if err != nil {
			logger.Info("Error decrypting secret, skipping", "secretKey", s.Key, "encodedString", s.EncryptedSecret)
			continue
		}
		decryptedSecretData[s.Key] = string(result.Plaintext)
	}
	return decryptedSecretData, nil
}

func convertContextMap(context map[string]string) map[string]*string {
	m := make(map[string]*string)
	for k, v := range context {
		m[k] = &v
	}
	return m
}

func getAuthenticatedVaultClient(vaultAuthenticationMethod string) (*vaultapi.Client, error) {
	vaultConfig := vaultapi.DefaultConfig()
	vaultClient, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	loginToken, err := vaultAuthentication(vaultAuthenticationMethod).login(vaultConfig)
	if err != nil {
		return nil, err
	}
	vaultClient.SetToken(loginToken)
	vaultClient.Auth()
	return vaultClient, nil
}

func vaultAuthentication(vaultAuthenticationMethod string) VaultAuthMethod {
	switch vaultAuthenticationMethod {
	case K8sAuthenticationMethod:
		return VaultK8sAuth{}
	case UserpassAuthenticationMethod:
		return VaultUserpassAuth{}
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
