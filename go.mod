module github.com/patoarvizu/kms-vault-operator

go 1.13

require (
	github.com/aws/aws-sdk-go v1.31.9
	github.com/coreos/prometheus-operator v0.41.1
	github.com/go-logr/logr v0.1.0
	github.com/hashicorp/vault v1.4.2
	github.com/hashicorp/vault/api v1.0.5-0.20200317185738-82f498082f02
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/operator-framework/operator-sdk v0.19.1
	github.com/prometheus/client_golang v1.6.0
	github.com/radovskyb/watcher v1.0.7
	github.com/slok/kubewebhook v0.10.0
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM

replace k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
