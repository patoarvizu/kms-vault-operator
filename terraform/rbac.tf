resource kubernetes_cluster_role_v1 kms_vault_operator {
  metadata {
    name = "kms-vault-operator"
  }

  rule {
    verbs      = ["create", "patch"]
    api_groups = [""]
    resources  = ["events"]
  }

  rule {
    verbs      = ["get", "list", "watch"]
    api_groups = [""]
    resources  = ["namespaces"]
  }

  rule {
    verbs      = ["*"]
    api_groups = ["k8s.patoarvizu.dev"]
    resources  = ["kmsvaultsecrets", "partialkmsvaultsecrets"]
  }
}

resource kubernetes_cluster_role_binding_v1 kms_vault_operator {
  metadata {
    name = "kms-vault-operator"
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account_v1.kms_vault_operator.metadata[0].name
    namespace = kubernetes_service_account_v1.kms_vault_operator.metadata[0].namespace
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role_v1.kms_vault_operator.metadata[0].name
  }
}

resource kubernetes_role_v1 kms_vault_operator {
  metadata {
    name = "kms-vault-operator"
    namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name
  }

  rule {
    verbs      = ["*"]
    api_groups = [""]
    resources  = ["configmaps"]
  }

  rule {
    verbs          = ["update"]
    api_groups     = ["apps"]
    resources      = ["deployments/finalizers"]
    resource_names = ["kms-vault-operator"]
  }

  rule {
    verbs      = ["get", "list", "watch"]
    api_groups = ["k8s.patoarvizu.dev"]
    resources  = ["kmsvaultsecrets", "partialkmsvaultsecrets"]
  }

  rule {
    verbs      = ["update"]
    api_groups = ["k8s.patoarvizu.dev"]
    resources  = ["kmsvaultsecrets/finalizers"]
  }

  rule {
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
    api_groups = ["coordination.k8s.io"]
    resources  = ["leases"]
  }
}

resource kubernetes_role_binding_v1 kms_vault_operator {
  metadata {
    name = "kms-vault-operator"
    namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name
  }

  subject {
    kind = "ServiceAccount"
    name      = kubernetes_service_account_v1.kms_vault_operator.metadata[0].name
    namespace = kubernetes_service_account_v1.kms_vault_operator.metadata[0].namespace
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role_v1.kms_vault_operator.metadata[0].name
  }
}