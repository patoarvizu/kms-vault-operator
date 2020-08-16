module github.com/patoarvizu/kms-vault-operator

go 1.13

require (
	github.com/aws/aws-sdk-go v1.25.48
	github.com/coreos/prometheus-operator v0.38.1-0.20200424145508-7e176fda06cc
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.4
	github.com/hashicorp/vault v1.4.2
	github.com/hashicorp/vault/api v1.0.5-0.20200317185738-82f498082f02
	github.com/operator-framework/operator-sdk v0.18.2
	github.com/prometheus/client_golang v1.5.1
	github.com/radovskyb/watcher v1.0.7
	github.com/slok/kubewebhook v0.9.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/openshift/api => github.com/openshift/api v0.0.0-20190924102528-32369d4db2ad // Required until https://github.com/operator-framework/operator-lifecycle-manager/pull/1241 is resolved

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM

replace k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
