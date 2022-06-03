resource kubernetes_service_account_v1 kms_vault_operator {
  metadata {
    name = "kms-vault-operator"
    namespace = var.create_namespace ? kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name : data.kubernetes_namespace_v1.ns[var.namespace_name].metadata[0].name
  }
}