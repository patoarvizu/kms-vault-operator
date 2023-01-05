module github.com/patoarvizu/kms-vault-operator

go 1.16

require (
	github.com/aws/aws-sdk-go v1.37.19
	github.com/go-logr/logr v0.4.0
	github.com/hashicorp/go-secure-stdlib/awsutil v0.1.6
	github.com/hashicorp/vault/api v1.1.2-0.20210713235431-1fc8af4c041f
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/prometheus/client_golang v1.11.0
	github.com/radovskyb/watcher v1.0.7
	github.com/slok/kubewebhook v0.10.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)
