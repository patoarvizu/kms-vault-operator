module github.com/patoarvizu/kms-vault-operator

go 1.16

require (
	github.com/aws/aws-sdk-go v1.37.19
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/go-secure-stdlib/awsutil v0.1.6
	github.com/hashicorp/vault/api v1.1.2-0.20210713235431-1fc8af4c041f
	github.com/kisielk/errcheck v1.5.0 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/operator-framework/operator-lib v0.1.0
	github.com/prometheus/client_golang v1.7.1
	github.com/radovskyb/watcher v1.0.7
	github.com/slok/kubewebhook v0.10.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)
