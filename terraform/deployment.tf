resource kubernetes_deployment_v1 kms_vault_operator {
  metadata {
    name = "kms-vault-operator"
    namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "kms-vault-operator"
      }
    }

    template {
      metadata {
        labels = {
          app = "kms-vault-operator"
        }
        annotations = var.pod_annotations
      }

      spec {
        container {
          name    = "kms-vault-operator"
          image   = "patoarvizu/kms-vault-operator:${var.image_version}"
          command = ["/manager"]
          args    = ["--enable-leader-election", "--vault-authentication-method=${var.vault_authentication_method}", "--sync-period-seconds=${var.sync_period_seconds}"]

          port {
            name           = "http-metrics"
            container_port = 8080
          }

          env {
            name = "WATCH_NAMESPACE"
            value = var.watch_namespace
          }

          env {
            name  = "AWS_REGION"
            value = var.aws_region
          }

          env {
            name  = "VAULT_ADDR"
            value = var.vault_address
          }

          dynamic "env" {
            for_each = !var.tls_enable ? {skip_verify: true} : {}
            content {
              name = "VAULT_SKIP_VERIFY"
              value = "true"
            }
          }
          dynamic "env" {
            for_each = var.tls_enable ? {ca_path: true} : {}
            content {
              name = "VAULT_CAPATH"
              value = "${var.tls_mount_path}/${var.tls_cert_file_name}"
            }
          }

          dynamic "env" {
            for_each = var.auth_method_env_vars
            content {
              name = env.value["name"]
              value = env.value["value"]
            }
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

          dynamic "volume_mount" {
            for_each = var.tls_enable ? {volume_mount: true} : {}
            content {
              name = "tls"
              mount_path = var.tls_mount_path
            }
          }

          image_pull_policy = "IfNotPresent"
        }

        dynamic "volume" {
          for_each = var.tls_enable ? {volume: true} : {}
          content {
            name = "tls"
            secret {
              secret_name = var.tls_secret_name
            }
          }
        }

        service_account_name = kubernetes_service_account_v1.kms_vault_operator.metadata[0].name
      }
    }
  }
}