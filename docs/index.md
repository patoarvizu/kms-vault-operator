# kms-vault-operator

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square)

KMS Vault operator

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| authMethodVariables | list | `[{"name":"VAULT_K8S_ROLE","value":"kms-vault-operator"},{"name":"VAULT_K8S_LOGIN_ENDPOINT","value":"auth/kubernetes/login"}]` | The set of environment variables required to configure the authentication to be used by the operator. The set of variables will vary depending on the value of `vaultAuthenticationMethod` and they're documented [here](https://github.com/patoarvizu/kms-vault-operator#vault). |
| aws | object | `{"iamCredentialsSecrets":[{"name":"AWS_ACCESS_KEY_ID","valueFrom":{"secretKeyRef":{"key":"AWS_ACCESS_KEY_ID","name":"aws-secrets"}}},{"name":"AWS_SECRET_ACCESS_KEY","valueFrom":{"secretKeyRef":{"key":"AWS_SECRET_ACCESS_KEY","name":"aws-secrets"}}}],"region":"us-east-1"}` | The value to set on the `AWS_DEFAULT_REGION` environment variable. |
| aws.iamCredentialsSecrets | list | `[{"name":"AWS_ACCESS_KEY_ID","valueFrom":{"secretKeyRef":{"key":"AWS_ACCESS_KEY_ID","name":"aws-secrets"}}},{"name":"AWS_SECRET_ACCESS_KEY","valueFrom":{"secretKeyRef":{"key":"AWS_SECRET_ACCESS_KEY","name":"aws-secrets"}}}]` | The set of environment variables and their references to `Secret`s that need to be added to the operator for KMS operations. |
| global.imagePullPolicy | string | `"IfNotPresent"` | The imagePullPolicy to be used on both the operator and webhook. |
| global.imageVersion | string | `nil` | The image version used for both the operator and webhook. |
| global.podAnnotations | string | `nil` | A map of annotations to be set on both the operator and webhook pods. Useful if using an annotation-based system like [kube2iam](https://github.com/jtblin/kube2iam) for dynamically injecting credentials. |
| global.prometheusMonitoring.enable | bool | `true` | Controls whether the `ServiceMonitor` objects are created for both the operator and the webhook. |
| imagePullPolicy | string | `"IfNotPresent"` | The imagePullPolicy to be used on the operator. |
| imageVersion | string | `"latest"` | The image version used for the operator. |
| podAnnotations | string | `nil` | A map of annotations to be set on the operator pods. Useful if using an annotation-based system like [kube2iam](https://github.com/jtblin/kube2iam) for dynamically injecting credentials. |
| prometheusMonitoring.enable | bool | `false` | Create the `Service` and `ServiceMonitor` objects to enable Prometheus monitoring on the operator. |
| serviceAccount.name | string | `"kms-vault-operator"` | The name of the `ServiceAccount` to be created. |
| syncPeriodSeconds | int | `120` | The value to be set on the `--sync-period-seconds` flag. |
| tls.enable | bool | `false` | Controls whether the operator Vault client should use TLS when talking to the target Vault server. |
| tls.mountPath | string | `nil` | The path where the CA cert from the secret should be mounted. This is required if `tls.enable` is set to `true`. |
| tls.secretName | string | `nil` | The name of the `Secret` from which the CA cert will be mounted. This is required if `tls.enable` is set to `true`. |
| validatingWebhook.caBundle | string | `"Cg=="` | The base64-encoded public CA certificate to be set on the `ValidatingWebhookConfiguration`. Note that it defaults to `Cg==` which is a base64-encoded empty string. If this value is not automatically set by cert-manager, or some other mutating webhook, this should be set explicitly. |
| validatingWebhook.certManager.apiVersion | string | `"cert-manager.io/v1alpha2"` | The `apiVersion` of the `Certificate` object created by the chart. It depends on the versions made available by the specific cert-manager running on the cluster. |
| validatingWebhook.certManager.duration | string | `"2160h"` | The value to be set directly on the `duration` field of the `Certificate`. |
| validatingWebhook.certManager.injectSecret | bool | `true` | Enables auto-injection of a certificate managed by [cert-manager](https://github.com/jetstack/cert-manager). |
| validatingWebhook.certManager.issuerRef | object | `{"kind":"ClusterIssuer","name":"selfsigning-issuer"}` | The `name` and `kind` of the cert-manager issuer to be used. |
| validatingWebhook.certManager.renewBefore | string | `"360h"` | The value to be set directly on the `renewBefore` field of the `Certificate`. |
| validatingWebhook.enabled | bool | `false` | Deploy the resources to enable the webhook used for custom resource validation. The rest of the settings under `validatingWebhook` are ignored if this is set to `false`. |
| validatingWebhook.failurePolicy | string | `"Fail"` | The value to set directly on the `failurePolicy` of the `ValidatingWebhookConfiguration`. Valid values are `Fail` or `Ignore`. |
| validatingWebhook.imagePullPolicy | string | `"IfNotPresent"` | The imagePullPolicy to be used on the webhook. |
| validatingWebhook.imageVersion | string | `"latest"` | The image version used for the webhook. |
| validatingWebhook.namespaceSelectorExpressions | list | `[{"key":"kms-vault-operator","operator":"DoesNotExist"}]` | A label selector expression to determine what namespaces should be in scope for the validating webhook. |
| validatingWebhook.podAnnotations | string | `nil` | A map of annotations to be set on the webhook pods. Useful if using an annotation-based system like [kube2iam](https://github.com/jtblin/kube2iam) for dynamically injecting credentials. |
| validatingWebhook.prometheusMonitoring.enable | bool | `false` | Create the `Service` and `ServiceMonitor` objects to enable Prometheus monitoring on the webhook. |
| validatingWebhook.tls.mountPath | string | `"/tls"` | The path where the certificate key pair will be mounted. |
| validatingWebhook.tls.secretName | string | `"kms-vault-validating-webhook"` | The name of the `Secret` that contains the certificate key pair to be used by the webhook. This is only used if `validatingWebhook.certManager.injectSecret` is set to `false`. |
| vault.address | string | `"https://vault:8200"` | The API endpoint of the target Vault cluster. |
| vaultAuthenticationMethod | string | `"k8s"` | The value to be set on the `--vault-authentication-method` flag. |
| watchNamespace | string | `""` | The value to be set on the `WATCH_NAMESPACE` environment variable. |
