resource kubernetes_service_v1 kms_vault_operator_metrics {
  for_each = var.enable_prometheus_monitoring ? {"monitor": true} : {}
  metadata {
    name = "kms-vault-operator-metrics"
    namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name

    labels = {
      app = "kms-vault-operator"
    }
  }

  spec {
    port {
      name        = "http-metrics"
      protocol    = "TCP"
      port        = 8080
      target_port = "http-metrics"
    }

    selector = {
      app = "kms-vault-operator"
    }

    type = "ClusterIP"
  }
}

resource kubernetes_service_v1 kms_vault_validating_webhook_metrics {
  for_each = var.enable_validating_webhook && var.enable_prometheus_monitoring ? {"monitor": true} : {}
  metadata {
    name = "kms-vault-validating-webhook-metrics"
    namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name

    labels = {
      app = "kms-vault-validating-webhook-metrics"
    }
  }

  spec {
    port {
      name        = "webhook-metrics"
      protocol    = "TCP"
      port        = 8081
      target_port = "webhook-metrics"
    }

    selector = {
      app = "kms-vault-validating-webhook"
    }

    type = "ClusterIP"
  }
}

resource kubernetes_manifest servicemonitor_kms_vault_operator {
  for_each = var.enable_prometheus_monitoring ? {"monitor": true} : {}
  manifest = {
    apiVersion = "monitoring.coreos.com/v1"
    kind = "ServiceMonitor"
    metadata = {
      name = "kms-vault-operator"
      namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name
    }
    spec = {
      endpoints = [
        {
          path = "/metrics"
          port = "http-metrics"
        },
      ]
      selector = {
        matchLabels = {
          app = "kms-vault-operator"
        }
      }
    }
  }
}

resource kubernetes_manifest servicemonitor_kms_vault_operator_webhook {
  for_each = var.enable_validating_webhook && var.enable_prometheus_monitoring ? {"monitor": true} : {}
  manifest = {
    apiVersion = "monitoring.coreos.com/v1"
    kind = "ServiceMonitor"
    metadata = {
      name = "kms-vault-operator-webhook"
      namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name
    }
    spec = {
      endpoints = [
        {
          path = "/"
          port = "webhook-metrics"
        },
      ]
      selector = {
        matchLabels = {
          app = "kms-vault-validating-webhook-metrics"
        }
      }
    }
  }
}