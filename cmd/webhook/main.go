package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	kmsvaultv1alpha1 "github.com/patoarvizu/kms-vault-operator/api/v1alpha1"
	"github.com/radovskyb/watcher"
	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	validatingwh "github.com/slok/kubewebhook/pkg/webhook/validating"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slok/kubewebhook/pkg/observability/metrics"
)

type webhookCfg struct {
	certFile    string
	keyFile     string
	addr        string
	metricsAddr string
}

var cfg = &webhookCfg{}
var cachedCertificate tls.Certificate

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

	w := watcher.New()
	defer w.Close()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)
	err := w.Add(cfg.certFile)
	if err != nil {
		logger.Errorf("Error: %v", err)
	}
	go func() {
		for {
			select {
			case <-w.Event:
				err = cacheCertificate(cfg.certFile, cfg.keyFile)
				if err != nil {
					logger.Errorf("Error refreshing certificate: %v", err)
					os.Exit(1)
				}
			case <-w.Closed:
				logger.Errorf("Certificate file watch closed")
				os.Exit(1)
			}
		}
	}()
	go w.Start(time.Millisecond * 100)

	v := validatingwh.ValidatorFunc(validate)

	vhc := validatingwh.WebhookConfig{
		Name: "kms-vault-secret-validator",
		Obj:  &kmsvaultv1alpha1.KMSVaultSecret{},
	}
	reg := prometheus.NewRegistry()
	metricsRec := metrics.NewPrometheus(reg)
	wh, err := validatingwh.NewWebhook(vhc, v, nil, metricsRec, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}
	webhookError := make(chan error)
	go func() {
		err = cacheCertificate(cfg.certFile, cfg.keyFile)
		if err != nil {
			logger.Errorf("Error loading certificate: %v", err)
			os.Exit(1)
		}
		server := http.Server{Addr: cfg.addr, Handler: whhttp.MustHandlerFor(wh), TLSConfig: &tls.Config{GetCertificate: getCertificate}}
		webhookError <- server.ListenAndServeTLS(cfg.certFile, cfg.keyFile)
	}()
	metricsError := make(chan error)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	go func() {
		metricsError <- http.ListenAndServe(cfg.metricsAddr, promHandler)
	}()
	if <-webhookError != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", <-webhookError)
		os.Exit(1)
	}
	if <-metricsError != nil {
		fmt.Fprintf(os.Stderr, "error serving metrics: %s", <-metricsError)
		os.Exit(1)
	}
}

func cacheCertificate(certfile, keyfile string) error {
	cert, err := tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		return err
	}
	cachedCertificate = cert
	return nil
}

func getCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return &cachedCertificate, nil
}
