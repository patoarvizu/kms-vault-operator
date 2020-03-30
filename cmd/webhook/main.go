package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	kmsvaultv1alpha1 "github.com/patoarvizu/kms-vault-operator/pkg/apis/k8s/v1alpha1"
	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	validatingwh "github.com/slok/kubewebhook/pkg/webhook/validating"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type webhookCfg struct {
	certFile    string
	keyFile     string
	addr        string
	metricsAddr string
}

var cfg = &webhookCfg{}

func validate(_ context.Context, obj metav1.Object) (bool, validatingwh.ValidatorResult, error) {
	awsSession, err := session.NewSession()
	if err != nil {
		return false, validatingwh.ValidatorResult{}, err
	}
	svc := kms.New(awsSession)
	secret, ok := obj.(*kmsvaultv1alpha1.KMSVaultSecret)
	if ok {
		for _, s := range secret.Spec.Secrets {
			if s.EmptySecret {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(s.EncryptedSecret)
			if err != nil {
				return false, validatingwh.ValidatorResult{
					Valid:   false,
					Message: fmt.Sprintf("Error decoding key %s in KMSVaultSecret %s", s.Key, secret.ObjectMeta.Name),
				}, nil
			}
			_, err = svc.Decrypt(&kms.DecryptInput{CiphertextBlob: decoded, EncryptionContext: getApplicableContext(s.SecretContext, secret.Spec.SecretContext)})
			if err != nil {
				return false, validatingwh.ValidatorResult{
					Valid:   false,
					Message: fmt.Sprintf("Error decrypting key %s in KMSVaultSecret %s", s.Key, secret.ObjectMeta.Name),
				}, nil
			}
		}
	} else {
		partial, ok := obj.(*kmsvaultv1alpha1.PartialKMSVaultSecret)
		if !ok {
			return false, validatingwh.ValidatorResult{}, fmt.Errorf("Object is neither a KMSVaultSecret or a PartialKMSVaultSecret")
		}
		for _, s := range partial.Spec.Secrets {
			if s.EmptySecret {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(s.EncryptedSecret)
			if err != nil {
				return false, validatingwh.ValidatorResult{
					Valid:   false,
					Message: fmt.Sprintf("Error decoding key %s in KMSVaultSecret %s", s.Key, partial.ObjectMeta.Name),
				}, nil
			}
			_, err = svc.Decrypt(&kms.DecryptInput{CiphertextBlob: decoded, EncryptionContext: getApplicableContext(s.SecretContext, partial.Spec.SecretContext)})
			if err != nil {
				return false, validatingwh.ValidatorResult{
					Valid:   false,
					Message: fmt.Sprintf("Error decrypting key %s in KMSVaultSecret %s", s.Key, partial.ObjectMeta.Name),
				}, nil
			}
		}
	}
	return false, validatingwh.ValidatorResult{Valid: true}, nil
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

func main() {
	logger := &log.Std{}
	logger.Infof("Starting webhook!")

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")
	fl.StringVar(&cfg.addr, "listen-addr", ":4443", "The address to start the server")
	fl.StringVar(&cfg.metricsAddr, "metrics-addr", ":8081", "The address where the Prometheus-style metrics are published")

	fl.Parse(os.Args[1:])

	v := validatingwh.ValidatorFunc(validate)

	vhc := validatingwh.WebhookConfig{
		Name: "kms-vault-secret-validator",
		Obj:  &kmsvaultv1alpha1.KMSVaultSecret{},
	}

	wh, err := validatingwh.NewWebhook(vhc, v, nil, nil, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}
	err = http.ListenAndServeTLS(cfg.addr, cfg.certFile, cfg.keyFile, whhttp.MustHandlerFor(wh))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
