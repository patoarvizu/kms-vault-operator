module github.com/patoarvizu/kms-vault-operator

go 1.12

require (
	cloud.google.com/go v0.37.1 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.2.0 // indirect
	github.com/Azure/go-autorest v11.5.2+incompatible // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30 // indirect
	github.com/aws/aws-sdk-go v1.19.1
	github.com/census-instrumentation/opencensus-proto v0.1.0 // indirect
	github.com/coreos/prometheus-operator v0.26.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/emicklei/go-restful v2.9.1+incompatible // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.18.0 // indirect
	github.com/go-openapi/jsonreference v0.18.0 // indirect
	github.com/go-openapi/spec v0.19.0
	github.com/go-openapi/swag v0.18.0 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190322150113-a30f18a1bae0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.0 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/vault v1.1.0
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mailru/easyjson v0.0.0-20190312143242-1de009706dbe // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/operator-framework/operator-sdk v0.7.1
	github.com/pborman/uuid v0.0.0-20180906182336-adf5a7427709 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pierrec/lz4 v0.0.0-20190222153722-062282ea0dcf // indirect
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/procfs v0.0.0-20190322151404-55ae3d9d5573 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.4.0 // indirect
	github.com/spf13/pflag v1.0.3
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1 // indirect
	golang.org/x/crypto v0.0.0-20190320223903-b7391e95e576 // indirect
	golang.org/x/net v0.0.0-20190322120337-addf6b3196f6 // indirect
	golang.org/x/oauth2 v0.0.0-20190319182350-c85d3e98c914 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	golang.org/x/tools v0.0.0-20190322172927-e779aa49e359 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	google.golang.org/genproto v0.0.0-20190321212433-e79c0c59cdb5 // indirect
	google.golang.org/grpc v1.19.1 // indirect
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31
	k8s.io/code-generator v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/gengo v0.0.0-20190319205223-bc9033e9ec9e // indirect
	k8s.io/kube-openapi v0.0.0-20180711000925-0cf8f7e6ed1d
	sigs.k8s.io/controller-runtime v0.1.10
	sigs.k8s.io/controller-tools v0.1.8 // indirect
	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
)

// Pinned to kubernetes-1.13.1
replace (
	k8s.io/api => k8s.io/api v0.0.0-20181213150558-05914d821849
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181213153335-0fe22c71c476
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31
)

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.29.0
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20181117043124-c2090bec4d9b
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20180711000925-0cf8f7e6ed1d
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.10
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde
)

replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.8.2
