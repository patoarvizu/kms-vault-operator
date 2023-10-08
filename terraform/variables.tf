variable image_version {
  type = string
  default = "latest"
  description = "The label of the image to run."
}

variable create_namespace {
  type = bool
  default = true
  description = "If true, a new namespace will be created with the name set to the value of the namespace_name variable. If false, it will look up an existing namespace with the name of the value of the namespace_name variable."
}

variable namespace_name {
  type = string
  default = "kms-vault-operator"
  description = "The name of the namespace to create or look up."
}

variable pod_annotations {
  type = map
  default = {}
  description = "Map of annotations to add to the operator pods."
}

variable vault_authentication_method {
  type = string
  default = "k8s"
  description = "The name of the authentication method used to connect to Vault."
}

variable sync_period_seconds {
  type = number
  default = 120
  description = "The secret sync frequency, in seconds."
}

variable watch_namespace {
  type = string
  default = ""
  description = "The value to be set on the `WATCH_NAMESPACE` environment variable."
}

variable aws_region {
  type = string
  default = "us-east-1"
  description = "The name of the AWS region to use."
}

variable vault_address {
  type = string
  default = "https://vault:8200"
  description = "The address of the target Vault cluster"
}

variable auth_method_env_vars {
  type = list(object({
    name = string
    value = string
  }))
  default = [
    {
      name = "VAULT_K8S_ROLE"
      value = "kms-vault-operator"
    },
    {
      name = "VAULT_K8S_LOGIN_ENDPOINT"
      value = "auth/kubernetes/login"
    }
  ]
  description = "Environment variables required for Vault authentication, depending on the value of `vault_authentication_method`."
}

variable iam_credentials_env_vars {
  type = list(object({
    name = string
    value = string
  }))
  default = []
  description = "Environment variables to inject IAM credentials"
}

variable iam_credentials_env_from_vars {
  type = list(object({
    name = string
    secret_ref_key = string
    secret_ref_name = string
  }))
  default = []
  description = "Optional environment variables to reference Kubernetes Secrets to inject IAM credentials."
}

variable enable_prometheus_monitoring {
  type = bool
  default = false
  description = "Set to `true` to create additional `Service` and `ServiceMonitor` objects for Prometheus monitoring. Requires the Prometheus operator to already be running in the cluster."
}

variable service_monitor_custom_labels {
  type = map
  default = {}
  description = "Custom labels to add to the `ServiceMonitor` objects."
}

variable tls_mount_path {
  type = string
  default = "/tls"
  description = "The path at which the TLS certificates are mounted on the operator."
}

variable tls_cert_file_name {
  type = string
  default = "tls.crt"
  description = "The name of the TLS certificate file mounted on the operator."
}

variable tls_enable {
  type = bool
  default = false
  description = "Set to `true` to mount TLS certificates and enable TLS communication between the operator and Vault. If set to `false`, the operator will run with `VAULT_SKIP_VERIFY = true`, which is not recommended."
}

variable tls_secret_name {
  type = string
  default = "vault-tls"
  description = "The name of the Kubernetes Secret containing the CA certificate for the target Vault cluster."
}

variable enable_validating_webhook {
  type = bool
  default = false
  description = "Create the additional resources required to create the validating webhook."
}

variable webhook_cert_manager_inject_secret {
  type = bool
  default = true
  description = "Set to `true` to create a `Certificate` object. Requires the cert-manager operator to already be running in the cluster."
}

variable webhook_cert_manager_api_version {
  type = string
  default = "cert-manager.io/v1alpha2"
  description = "The apiVersion value of the `Certificate` object to be created. Only used if `webhook_cert_manager_inject_secret` is set to `true`."
}

variable webhook_cert_manager_duration {
  type = string
  default = "2160h"
  description = "The `duration` field of the cert-manager `Certificate`."
}

variable webhook_cert_manager_renew_before {
  type = string
  default = "360h"
  description = "The `renewBefore` field of the cert-manager `Certificate`."
}

variable webhook_cert_manager_name {
  type = string
  default = "selfsigning-issuer"
  description = "The `issuerRef.name` field of the cert-manager `Certificate`."
}

variable webhook_cert_manager_kind {
  type = string
  default = "ClusterIssuer"
  description = "The `issuerRef.kind` field of the cert-manager `Certificate`."
}

variable webhook_ca_bundle {
  type = string
  default = "Cg=="
  description = "The base-64 encoded CA bundle to be set in the `ValidatingWebhookConfiguration` object. It defaults to `Cg==` which is a base-64 encoded empty string."
}

variable webhook_failure_policy {
  type = string
  default = "Fail"
  description = "The `failurePolicy` field of the `ValidatingWebhookConfiguration` object."
}

variable webhook_namespace_selector_expressions {
  type = list(object({
    key = string
    operator = string
  }))
  default = [
    {
      key = "kms-vault-operator"
      operator = "DoesNotExist"
    }
  ]
  description = "The set of namespace selector expressions to set on the `ValidatingWebhookConfiguration` object."
}

variable webhook_replicas {
  type = number
  default = 1
  description = "The number of webhook replicas."
}

variable webhook_pod_annotations {
  type = map
  default = {}
  description = "The set of annotations to add to the webhook pods."
}

variable webhook_image_version {
  type = string
  default = "latest"
  description = "The image version of the webhook pods."
}

variable webhook_tls_mount_path {
  type = string
  default = "/tls"
  description = "The path at which the TLS certificates are mounted on the webhook."
}

variable webhook_tls_cert_file_name {
  type = string
  default = "tls.crt"
  description = "The name of the TLS certificate file mounted on the webhook."
}

variable webhook_private_file_name {
  type = string
  default = "tls.key"
  description = "The name of the TLS certificate private key file mounted on the webhook."
}

variable secret_mounts {
  type = list(object({
    secret_name = string
    mount_path = string
  }))
  default = []
  description = "References to Kubernetes secrets to be mounted on the workloads"
}