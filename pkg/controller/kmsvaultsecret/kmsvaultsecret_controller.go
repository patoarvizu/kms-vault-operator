package kmsvaultsecret

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"

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

var log = logf.Log.WithName("controller_kmsvaultsecret")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new KMSVaultSecret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKMSVaultSecret{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kmsvaultsecret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KMSVaultSecret
	err = c.Watch(&source.Kind{Type: &k8sv1alpha1.KMSVaultSecret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KMSVaultSecret
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &k8sv1alpha1.KMSVaultSecret{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileKMSVaultSecret{}

// ReconcileKMSVaultSecret reconciles a KMSVaultSecret object
type ReconcileKMSVaultSecret struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KMSVaultSecret object and makes changes based on the state read
// and what is in the KMSVaultSecret.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKMSVaultSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KMSVaultSecret")

	// Fetch the KMSVaultSecret instance
	instance := &k8sv1alpha1.KMSVaultSecret{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	awsSession, err := session.NewSession()
	if err != nil {
		reqLogger.Info("Error creating AWS session", err)
		return reconcile.Result{}, err
	}

	svc := kms.New(awsSession)
	decoded, err := base64.StdEncoding.DecodeString(instance.Spec.Secret.EncryptedSecret)
	if err != nil {
		reqLogger.Info("Error decoding encrypted secret", err)
		return reconcile.Result{}, err
	}
	result, err := svc.Decrypt(&kms.DecryptInput{CiphertextBlob: decoded})
	if err != nil {
		reqLogger.Info("Error decrypting secret", err)
		return reconcile.Result{}, err
	}
	vaultConfig := vaultapi.DefaultConfig()
	vaultConfig.ConfigureTLS(&vaultapi.TLSConfig{Insecure: true})
	vaultClient, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		reqLogger.Info("Error creating Vault client", err)
		return reconcile.Result{}, err
	}
	vaultToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		reqLogger.Info("Error reading service account token", err)
		return reconcile.Result{}, err
	}
	data := map[string]interface{}{
		"jwt":  string(vaultToken),
		"role": "kms-vault-operator",
	}
	secretAuth, err := vaultClient.Logical().Write("auth/kubernetes/login", data)
	if err != nil {
		reqLogger.Info("Error authenticating with Vault", err)
		return reconcile.Result{}, err
	}
	vaultClient.SetToken(secretAuth.Auth.ClientToken)
	vaultClient.Auth()
	writeData := map[string]interface{}{
		instance.Spec.Secret.Key: string(result.Plaintext),
	}
	_, writeErr := vaultClient.Logical().Write(instance.Spec.Path, writeData)
	if writeErr != nil {
		reqLogger.Info("Error writing secret to Vault")
		return reconcile.Result{}, writeErr
	} else {
		reqLogger.Info(fmt.Sprintf("Wrote %s to %s", instance.Spec.Secret.Key, instance.Spec.Path))
	}

	return reconcile.Result{Requeue: true}, nil
}
