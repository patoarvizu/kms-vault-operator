resource "kubernetes_manifest" "customresourcedefinition_kmsvaultsecrets_k8s_patoarvizu_dev" {
  manifest = {
    "apiVersion" = "apiextensions.k8s.io/v1beta1"
    "kind" = "CustomResourceDefinition"
    "metadata" = {
      "annotations" = {
        "controller-gen.kubebuilder.io/version" = "v0.3.0"
      }
      "creationTimestamp" = null
      "name" = "kmsvaultsecrets.k8s.patoarvizu.dev"
    }
    "spec" = {
      "group" = "k8s.patoarvizu.dev"
      "names" = {
        "kind" = "KMSVaultSecret"
        "listKind" = "KMSVaultSecretList"
        "plural" = "kmsvaultsecrets"
        "shortNames" = [
          "kmsvs",
        ]
        "singular" = "kmsvaultsecret"
      }
      "scope" = "Namespaced"
      "subresources" = {
        "status" = {}
      }
      "validation" = {
        "openAPIV3Schema" = {
          "description" = "KMSVaultSecret is the Schema for the kmsvaultsecrets API"
          "properties" = {
            "apiVersion" = {
              "description" = "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources"
              "type" = "string"
            }
            "kind" = {
              "description" = "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
              "type" = "string"
            }
            "metadata" = {
              "type" = "object"
            }
            "spec" = {
              "description" = "KMSVaultSecretSpec defines the desired state of KMSVaultSecret"
              "properties" = {
                "includeSecrets" = {
                  "items" = {
                    "type" = "string"
                  }
                  "type" = "array"
                  "x-kubernetes-list-type" = "set"
                }
                "kvSettings" = {
                  "properties" = {
                    "casIndex" = {
                      "minimum" = 0
                      "type" = "integer"
                    }
                    "engineVersion" = {
                      "enum" = [
                        "v1",
                        "v2",
                      ]
                      "type" = "string"
                    }
                  }
                  "required" = [
                    "engineVersion",
                  ]
                  "type" = "object"
                }
                "path" = {
                  "type" = "string"
                }
                "secretContext" = {
                  "additionalProperties" = {
                    "type" = "string"
                  }
                  "type" = "object"
                }
                "secrets" = {
                  "items" = {
                    "properties" = {
                      "emptySecret" = {
                        "type" = "boolean"
                      }
                      "encryptedSecret" = {
                        "type" = "string"
                      }
                      "key" = {
                        "type" = "string"
                      }
                      "secretContext" = {
                        "additionalProperties" = {
                          "type" = "string"
                        }
                        "type" = "object"
                      }
                    }
                    "required" = [
                      "key",
                    ]
                    "type" = "object"
                  }
                  "type" = "array"
                  "x-kubernetes-list-map-keys" = [
                    "key",
                  ]
                  "x-kubernetes-list-type" = "map"
                }
              }
              "required" = [
                "kvSettings",
                "path",
                "secrets",
              ]
              "type" = "object"
            }
            "status" = {
              "description" = "KMSVaultSecretStatus defines the observed state of KMSVaultSecret"
              "properties" = {
                "created" = {
                  "type" = "boolean"
                }
              }
              "type" = "object"
            }
          }
          "type" = "object"
        }
      }
      "version" = "v1alpha1"
      "versions" = [
        {
          "name" = "v1alpha1"
          "served" = true
          "storage" = true
        },
      ]
    }
  }
}

resource "kubernetes_manifest" "customresourcedefinition_partialkmsvaultsecrets_k8s_patoarvizu_dev" {
  manifest = {
    "apiVersion" = "apiextensions.k8s.io/v1beta1"
    "kind" = "CustomResourceDefinition"
    "metadata" = {
      "annotations" = {
        "controller-gen.kubebuilder.io/version" = "v0.3.0"
      }
      "creationTimestamp" = null
      "name" = "partialkmsvaultsecrets.k8s.patoarvizu.dev"
    }
    "spec" = {
      "group" = "k8s.patoarvizu.dev"
      "names" = {
        "kind" = "PartialKMSVaultSecret"
        "listKind" = "PartialKMSVaultSecretList"
        "plural" = "partialkmsvaultsecrets"
        "shortNames" = [
          "pkmsvs",
        ]
        "singular" = "partialkmsvaultsecret"
      }
      "scope" = "Namespaced"
      "subresources" = {
        "status" = {}
      }
      "validation" = {
        "openAPIV3Schema" = {
          "description" = "PartialKMSVaultSecret is the Schema for the partialkmsvaultsecrets API"
          "properties" = {
            "apiVersion" = {
              "description" = "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources"
              "type" = "string"
            }
            "kind" = {
              "description" = "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
              "type" = "string"
            }
            "metadata" = {
              "type" = "object"
            }
            "spec" = {
              "description" = "PartialKMSVaultSecretSpec defines the desired state of PartialKMSVaultSecret"
              "properties" = {
                "secretContext" = {
                  "additionalProperties" = {
                    "type" = "string"
                  }
                  "type" = "object"
                }
                "secrets" = {
                  "items" = {
                    "properties" = {
                      "emptySecret" = {
                        "type" = "boolean"
                      }
                      "encryptedSecret" = {
                        "type" = "string"
                      }
                      "key" = {
                        "type" = "string"
                      }
                      "secretContext" = {
                        "additionalProperties" = {
                          "type" = "string"
                        }
                        "type" = "object"
                      }
                    }
                    "required" = [
                      "key",
                    ]
                    "type" = "object"
                  }
                  "type" = "array"
                  "x-kubernetes-list-map-keys" = [
                    "key",
                  ]
                  "x-kubernetes-list-type" = "map"
                }
              }
              "required" = [
                "secrets",
              ]
              "type" = "object"
            }
            "status" = {
              "description" = "PartialKMSVaultSecretStatus defines the observed state of PartialKMSVaultSecret"
              "properties" = {
                "created" = {
                  "type" = "boolean"
                }
              }
              "type" = "object"
            }
          }
          "type" = "object"
        }
      }
      "version" = "v1alpha1"
      "versions" = [
        {
          "name" = "v1alpha1"
          "served" = true
          "storage" = true
        },
      ]
    }
  }
}