resource kubernetes_namespace_v1 ns {
  for_each = var.create_namespace ? {(var.namespace_name): true} : {}
  metadata {
    name = var.namespace_name
  }
}