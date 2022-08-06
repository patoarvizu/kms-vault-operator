<!-- BEGIN_TF_DOCS -->

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 0.14.9 |
| <a name="requirement_kubernetes"></a> [kubernetes](#requirement\_kubernetes) | ~> 2.8.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_kubernetes"></a> [kubernetes](#provider\_kubernetes) | ~> 2.8.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [kubernetes_cluster_role_binding_v1.kms_vault_operator](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/cluster_role_binding_v1) | resource |
| [kubernetes_cluster_role_v1.kms_vault_operator](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/cluster_role_v1) | resource |
| [kubernetes_deployment_v1.kms_vault_operator](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/deployment_v1) | resource |
| [kubernetes_deployment_v1.kms_vault_validating_webhook](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/deployment_v1) | resource |
| [kubernetes_manifest.certificate_kms_vault_validating_webhook](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest) | resource |
| [kubernetes_manifest.customresourcedefinition_kmsvaultsecrets_k8s_patoarvizu_dev](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest) | resource |
| [kubernetes_manifest.customresourcedefinition_partialkmsvaultsecrets_k8s_patoarvizu_dev](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest) | resource |
| [kubernetes_manifest.servicemonitor_kms_vault_operator](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest) | resource |
| [kubernetes_manifest.servicemonitor_kms_vault_operator_webhook](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest) | resource |
| [kubernetes_namespace_v1.ns](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/namespace_v1) | resource |
| [kubernetes_role_binding_v1.kms_vault_operator](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/role_binding_v1) | resource |
| [kubernetes_role_v1.kms_vault_operator](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/role_v1) | resource |
| [kubernetes_service_account_v1.kms_vault_operator](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_account_v1) | resource |
| [kubernetes_service_v1.kms_vault_operator_metrics](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_v1) | resource |
| [kubernetes_service_v1.kms_vault_validating_webhook](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_v1) | resource |
| [kubernetes_service_v1.kms_vault_validating_webhook_metrics](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_v1) | resource |
| [kubernetes_validating_webhook_configuration_v1.kms_vault_validating_webhook](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/validating_webhook_configuration_v1) | resource |
| [kubernetes_namespace_v1.ns](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/data-sources/namespace_v1) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_auth_method_env_vars"></a> [auth\_method\_env\_vars](#input\_auth\_method\_env\_vars) | Environment variables required for Vault authentication, depending on the value of `vault_authentication_method`. | <pre>list(object({<br>    name = string<br>    value = string<br>  }))</pre> | <pre>[<br>  {<br>    "name": "VAULT_K8S_ROLE",<br>    "value": "kms-vault-operator"<br>  },<br>  {<br>    "name": "VAULT_K8S_LOGIN_ENDPOINT",<br>    "value": "auth/kubernetes/login"<br>  }<br>]</pre> | no |
| <a name="input_aws_region"></a> [aws\_region](#input\_aws\_region) | The name of the AWS region to use. | `string` | `"us-east-1"` | no |
| <a name="input_create_namespace"></a> [create\_namespace](#input\_create\_namespace) | If true, a new namespace will be created with the name set to the value of the namespace\_name variable. If false, it will look up an existing namespace with the name of the value of the namespace\_name variable. | `bool` | `true` | no |
| <a name="input_enable_prometheus_monitoring"></a> [enable\_prometheus\_monitoring](#input\_enable\_prometheus\_monitoring) | Set to `true` to create additional `Service` and `ServiceMonitor` objects for Prometheus monitoring. Requires the Prometheus operator to already be running in the cluster. | `bool` | `false` | no |
| <a name="input_enable_validating_webhook"></a> [enable\_validating\_webhook](#input\_enable\_validating\_webhook) | Create the additional resources required to create the validating webhook. | `bool` | `false` | no |
| <a name="input_iam_credentials_env_from_vars"></a> [iam\_credentials\_env\_from\_vars](#input\_iam\_credentials\_env\_from\_vars) | Optional environment variables to reference Kubernetes Secrets to inject IAM credentials. | <pre>list(object({<br>    name = string<br>    secret_ref_key = string<br>    secret_ref_name = string<br>  }))</pre> | `[]` | no |
| <a name="input_iam_credentials_env_vars"></a> [iam\_credentials\_env\_vars](#input\_iam\_credentials\_env\_vars) | Environment variables to inject IAM credentials | <pre>list(object({<br>    name = string<br>    value = string<br>  }))</pre> | `[]` | no |
| <a name="input_image_version"></a> [image\_version](#input\_image\_version) | The label of the image to run. | `string` | `"latest"` | no |
| <a name="input_namespace_name"></a> [namespace\_name](#input\_namespace\_name) | The name of the namespace to create or look up. | `string` | `"kms-vault-operator"` | no |
| <a name="input_pod_annotations"></a> [pod\_annotations](#input\_pod\_annotations) | Map of annotations to add to the operator pods. | `map` | `{}` | no |
| <a name="input_service_monitor_custom_labels"></a> [service\_monitor\_custom\_labels](#input\_service\_monitor\_custom\_labels) | Custom labels to add to the `ServiceMonitor` objects. | `map` | `{}` | no |
| <a name="input_sync_period_seconds"></a> [sync\_period\_seconds](#input\_sync\_period\_seconds) | The secret sync frequency, in seconds. | `number` | `120` | no |
| <a name="input_tls_cert_file_name"></a> [tls\_cert\_file\_name](#input\_tls\_cert\_file\_name) | The name of the TLS certificate file mounted on the operator. | `string` | `"tls.crt"` | no |
| <a name="input_tls_enable"></a> [tls\_enable](#input\_tls\_enable) | Set to `true` to mount TLS certificates and enable TLS communication between the operator and Vault. If set to `false`, the operator will run with `VAULT_SKIP_VERIFY = true`, which is not recommended. | `bool` | `false` | no |
| <a name="input_tls_mount_path"></a> [tls\_mount\_path](#input\_tls\_mount\_path) | The path at which the TLS certificates are mounted on the operator. | `string` | `"/tls"` | no |
| <a name="input_tls_secret_name"></a> [tls\_secret\_name](#input\_tls\_secret\_name) | The name of the Kubernetes Secret containing the CA certificate for the target Vault cluster. | `string` | `"vault-tls"` | no |
| <a name="input_vault_address"></a> [vault\_address](#input\_vault\_address) | The address of the target Vault cluster | `string` | `"https://vault:8200"` | no |
| <a name="input_vault_authentication_method"></a> [vault\_authentication\_method](#input\_vault\_authentication\_method) | The name of the authentication method used to connect to Vault. | `string` | `"k8s"` | no |
| <a name="input_watch_namespace"></a> [watch\_namespace](#input\_watch\_namespace) | The value to be set on the `WATCH_NAMESPACE` environment variable. | `string` | `""` | no |
| <a name="input_webhook_ca_bundle"></a> [webhook\_ca\_bundle](#input\_webhook\_ca\_bundle) | The base-64 encoded CA bundle to be set in the `ValidatingWebhookConfiguration` object. It defaults to `Cg==` which is a base-64 encoded empty string. | `string` | `"Cg=="` | no |
| <a name="input_webhook_cert_manager_api_version"></a> [webhook\_cert\_manager\_api\_version](#input\_webhook\_cert\_manager\_api\_version) | The apiVersion value of the `Certificate` object to be created. Only used if `webhook_cert_manager_inject_secret` is set to `true`. | `string` | `"cert-manager.io/v1alpha2"` | no |
| <a name="input_webhook_cert_manager_duration"></a> [webhook\_cert\_manager\_duration](#input\_webhook\_cert\_manager\_duration) | The `duration` field of the cert-manager `Certificate`. | `string` | `"2160h"` | no |
| <a name="input_webhook_cert_manager_inject_secret"></a> [webhook\_cert\_manager\_inject\_secret](#input\_webhook\_cert\_manager\_inject\_secret) | Set to `true` to create a `Certificate` object. Requires the cert-manager operator to already be running in the cluster. | `bool` | `true` | no |
| <a name="input_webhook_cert_manager_kind"></a> [webhook\_cert\_manager\_kind](#input\_webhook\_cert\_manager\_kind) | The `issuerRef.kind` field of the cert-manager `Certificate`. | `string` | `"ClusterIssuer"` | no |
| <a name="input_webhook_cert_manager_name"></a> [webhook\_cert\_manager\_name](#input\_webhook\_cert\_manager\_name) | The `issuerRef.name` field of the cert-manager `Certificate`. | `string` | `"selfsigning-issuer"` | no |
| <a name="input_webhook_cert_manager_renew_before"></a> [webhook\_cert\_manager\_renew\_before](#input\_webhook\_cert\_manager\_renew\_before) | The `renewBefore` field of the cert-manager `Certificate`. | `string` | `"360h"` | no |
| <a name="input_webhook_failure_policy"></a> [webhook\_failure\_policy](#input\_webhook\_failure\_policy) | The `failurePolicy` field of the `ValidatingWebhookConfiguration` object. | `string` | `"Fail"` | no |
| <a name="input_webhook_image_version"></a> [webhook\_image\_version](#input\_webhook\_image\_version) | The image version of the webhook pods. | `string` | `"latest"` | no |
| <a name="input_webhook_namespace_selector_expressions"></a> [webhook\_namespace\_selector\_expressions](#input\_webhook\_namespace\_selector\_expressions) | The set of namespace selector expressions to set on the `ValidatingWebhookConfiguration` object. | <pre>list(object({<br>    key = string<br>    operator = string<br>  }))</pre> | <pre>[<br>  {<br>    "key": "kms-vault-operator",<br>    "operator": "DoesNotExist"<br>  }<br>]</pre> | no |
| <a name="input_webhook_pod_annotations"></a> [webhook\_pod\_annotations](#input\_webhook\_pod\_annotations) | The set of annotations to add to the webhook pods. | `map` | `{}` | no |
| <a name="input_webhook_private_file_name"></a> [webhook\_private\_file\_name](#input\_webhook\_private\_file\_name) | The name of the TLS certificate private key file mounted on the webhook. | `string` | `"tls.key"` | no |
| <a name="input_webhook_replicas"></a> [webhook\_replicas](#input\_webhook\_replicas) | The number of webhook replicas. | `number` | `1` | no |
| <a name="input_webhook_tls_cert_file_name"></a> [webhook\_tls\_cert\_file\_name](#input\_webhook\_tls\_cert\_file\_name) | The name of the TLS certificate file mounted on the webhook. | `string` | `"tls.crt"` | no |
| <a name="input_webhook_tls_mount_path"></a> [webhook\_tls\_mount\_path](#input\_webhook\_tls\_mount\_path) | The path at which the TLS certificates are mounted on the webhook. | `string` | `"/tls"` | no |

## Outputs

No outputs.
<!-- END_TF_DOCS -->