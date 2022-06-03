resource kubernetes_service_v1 kms_vault_validating_webhook {
  for_each = var.enable_validating_webhook ? {"webhook": true} : {}
  metadata {
    name = "kms-vault-validating-webhook"
    namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name

    labels = {
      app = "kms-vault-validating-webhook"
    }
  }

  spec {
    port {
      protocol    = "TCP"
      port        = 443
      target_port = "https"
    }

    selector = {
      app = "kms-vault-validating-webhook"
    }

    type = "ClusterIP"
  }
}

resource kubernetes_deployment_v1 kms_vault_validating_webhook {
  for_each = var.enable_validating_webhook ? {"webhook": true} : {}
  metadata {
    name = "kms-vault-validating-webhook"
    namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name
  }

  spec {
    replicas = var.webhook_replicas

    selector {
      match_labels = {
        app = "kms-vault-validating-webhook"
      }
    }

    template {
      metadata {
        labels = {
          app = "kms-vault-validating-webhook"
        }
        annotations = var.webhook_pod_annotations
      }

      spec {
        container {
          name    = "kms-vault-validating-webhook"
          image   = "patoarvizu/kms-vault-operator:${var.webhook_image_version}"
          command = [
            "/kms-vault-validating-webhook",
            "-tls-cert-file",
            "${var.webhook_tls_mount_path}/${var.webhook_tls_cert_file_name}",
            "-tls-key-file",
            "${var.webhook_tls_mount_path}/${var.webhook_private_file_name}"
          ]

          port {
            name           = "https"
            container_port = 4443
          }

          port {
            name           = "webhook-metrics"
            container_port = 8081
          }

          env {
            name  = "AWS_REGION"
            value = var.aws_region
          }

          dynamic "env" {
            for_each = var.iam_credentials_env_vars
            content {
              name = env.value["name"]
              value = env.value["value"]
            }
          }

          dynamic "env" {
            for_each = var.iam_credentials_env_from_vars
            content {
              name = env.value["name"]
              value_from {
                secret_key_ref {
                  key = env.value["secret_ref_key"]
                  name = env.value["secret_ref_name"]
                }
              }
            }
          }

          volume_mount {
            name       = "tls"
            mount_path = "/tls"
          }

          image_pull_policy = "IfNotPresent"
        }

        volume {
          name = "tls"

          secret {
            secret_name = "kms-vault-validating-webhook"
          }
        }

        service_account_name = kubernetes_service_account_v1.kms_vault_operator.metadata[0].name
      }
    }
  }
}

resource kubernetes_manifest certificate_kms_vault_validating_webhook {
  for_each = var.enable_validating_webhook && var.webhook_cert_manager_inject_secret ? {"certificate": true} : {}
  manifest = {
    apiVersion = var.webhook_cert_manager_api_version
    kind = "Certificate"
    metadata = {
      name = "kms-vault-validating-webhook"
      namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name
    }
    spec = {
      commonName = "kms-vault-validating-webhook"
      dnsNames = [
        "kms-vault-validating-webhook",
        "kms-vault-validating-webhook.${var.namespace_name}",
        "kms-vault-validating-webhook.${var.namespace_name}.svc",
      ]
      duration = format("%s0m0s", var.webhook_cert_manager_duration)
      issuerRef = {
        kind = var.webhook_cert_manager_kind
        name = var.webhook_cert_manager_name
      }
      renewBefore = format("%s0m0s", var.webhook_cert_manager_renew_before)
      secretName = "kms-vault-validating-webhook"
    }
  }
}

resource kubernetes_validating_webhook_configuration_v1 kms_vault_validating_webhook {
  for_each = var.enable_validating_webhook ? {"webhook": true} : {}
  metadata {
    name = "kms-vault-validating-webhook"
    annotations = var.webhook_cert_manager_inject_secret ? {"cert-manager.io/inject-ca-from" = "${var.namespace_name}/kms-vault-validating-webhook"} : {}
  }

  webhook {
    name = "kms-vault-validating-webhook.patoarvizu.dev"

    client_config {
      service {
        namespace = var.namespace_name
        name      = "kms-vault-validating-webhook"
      }
      ca_bundle = var.webhook_ca_bundle
    }

    rule {
      operations = ["CREATE", "UPDATE"]
      api_versions = ["v1alpha1"]
      api_groups = ["k8s.patoarvizu.dev"]
      resources = ["kmsvaultsecrets", "partialkmsvaultsecrets"]
    }

    failure_policy = var.webhook_failure_policy

    namespace_selector {
      dynamic "match_expressions" {
        for_each = var.webhook_namespace_selector_expressions
        content {
          key = match_expressions.value["key"]
          operator = match_expressions.value["operator"]
        }
      }
    }
  }
  lifecycle {
    ignore_changes = [
      webhook[0].client_config[0].ca_bundle # Ignoring changes to the ca_bundle attirbute, since this is usually dynamic
    ]
  }
}