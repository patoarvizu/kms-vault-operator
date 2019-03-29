# KMS Vault operator

[![CircleCI](https://circleci.com/gh/patoarvizu/kms-vault-operator.svg?style=svg)](https://circleci.com/gh/patoarvizu/kms-vault-operator)

<!-- TOC -->

- [KMS Vault operator](#kms-vault-operator)
    - [Configuration](#configuration)
        - [AWS](#aws)
        - [Vault](#vault)
            - [Kubernetes authentication method (`vaultAuthMethod: k8s`)](#kubernetes-authentication-method-vaultauthmethod-k8s)
            - [Vault token authentication method (`vaultAuthMethod: token`)](#vault-token-authentication-method-vaultauthmethod-token)
        - [Deploying the operator](#deploying-the-operator)
        - [Creating a secret](#creating-a-secret)
    - [Important notes by this project](#important-notes-by-this-project)
            - [Kubernetes namespaces and Vault namespaces](#kubernetes-namespaces-and-vault-namespaces)
            - [Multiple secrets writing to the same location](#multiple-secrets-writing-to-the-same-location)
            - [No validation on target path](#no-validation-on-target-path)
            - [Removing secrets when a `KMSVaultSecret` is deleted.](#removing-secrets-when-a-kmsvaultsecret-is-deleted)
            - [KMS encryption context is not supported (as of this version)](#kms-encryption-context-is-not-supported-as-of-this-version)
            - [Support for K/V V2 is limited (as of this version)](#support-for-kv-v2-is-limited-as-of-this-version)
    - [Help wanted!](#help-wanted)

<!-- /TOC -->

This operator manages `KMSVaultSecret` CRDs containing secrets encrypted with a KMS key (and base64-encoded) and stores them in Vault. This allows you to securely manage your Vault secrets in source control and because the decryption only happens at runtime, it's only in the operator memory, minimizing the exposure of the plain text secret. The operator resources and scaffolding code are managed and generated automatically with the [operator-sdk framework](https://github.com/operator-framework/operator-sdk).

The AWS credentials to do the decryption operation use [aws-sdk-go](https://github.com/aws/aws-sdk-go) and follows the default [credential precedence order](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html), so if you're going to inject environment variables or configuration files, you should do so on the operator's `Deployment` manifest (e.g. [deploy/operator.yaml](deploy/operator.yaml)).

Each secret will define a `path`, a `vaultAuthMethod`, a set of `kvSettings`, and a list of `secrets`.

This first version of the operator supports authenticating to Vault via the [Kubernetes auth method](https://www.vaultproject.io/docs/auth/kubernetes.html) (i.e. `vaultAuthMethod: k8s`) or directly via a [Vault token](https://www.vaultproject.io/docs/auth/token.html) (i.e. `vaultAuthMethod: token`). Support for more authentication methods will be added in the future. Note that the configuration required for the operator to perform KMS and Vault operations is not done on the `KMSVaultSecret` CR but on the operator `Deployment` itself, and is documented below.

## Configuration

### AWS

As stated above, the required configuration to allow the operator to decrypt secrets should be injected via the operator `Deployment` manifest (as `AWS_*` environment variables or `~/.aws/credentials`/`~/.aws/config` files). The aws-sdk-go library will also discover IAM roles at runtime if the operator is running on an EC2 instance with an instance role, in which case no additional configuration should be required on the operator.

### Vault

In addition to the AWS configuration, Vault configuration should be injected too. The common `VAULT_*` [environment variables](https://www.vaultproject.io/docs/commands/#environment-variables) will be read and used by the [client](https://github.com/hashicorp/vault/tree/master/api). The variables that would need to be set will vary depending on your environment, but you'll typically want to at least set `VAULT_ADDR`, and either `VAULT_CACERT`/`VAULT_CAPATH` (pointing to a correspoinging mounted file or directory) or `VAULT_SKIP_VERIFY`. This is all done on the operator `Deployment` manifest.

The following sections document the environment variables used by each authentication method and their defaults. They are used in addition to the AWS and base Vault variables described before.

#### Kubernetes authentication method (`vaultAuthMethod: k8s`)

Environment variable | Required? | Default | Description
---------------------|-----------|---------|------------
`VAULT_K8S_ROLE` | N | `kms-vault-operator` | The Vault role that the client should authenticate as on the Kubernetes login endpoint.
`VAULT_K8S_LOGIN_ENDPOINT` | N | `auth/kubernetes/login` | The Kubernetes authentication endpoint in Vault

#### Vault token authentication method (`vaultAuthMethod: token`)

This method simply follows the convention and uses the `VAULT_TOKEN` environment variable to authenticate.

Environment variable | Required? | Default | Description
---------------------|-----------|---------|------------
`VAULT_TOKEN` | Y | | The Vault token used to perform operations on Vault.

### Deploying the operator

Assuming `deploy/operator.yaml` has been configured appropriately, run the following to deploy the operator.

```
kubectl apply -f deploy/crds/k8s_v1alpha1_kmsvaultsecret_crd.yaml
kubectl apply -f deploy/serviceaccount.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/operator.yaml
```

### Creating a secret

Setting up the AWS resources required to encrypt/decrypt a KMS secret are outside of the scope of this documentation (but you can read about KMS [here](https://docs.aws.amazon.com/kms/latest/developerguide/overview.html)), so we're going to assume that the KMS key already exists and that the credentials used to encrypt the secret offline, as well as those use to decrypt it at runtime have the required permissions to perform those actions. If you're Terraform-oriented, you might find [these modules](https://github.com/patoarvizu/terraform-kms-encryption/tree/master/modules) useful for dealing with KMS.

From the command line, you can use the [aws kms encrypt command](https://docs.aws.amazon.com/cli/latest/reference/kms/encrypt.html), e.g.
```
aws kms encrypt --key-id <key-id-or-alias> --plaintext "Hello world" --output text --query CiphertextBlob
```
Then take that output and set it as an item the `spec.secrets` array of your `KMSVaultSecret` resource, e.g.
```
apiVersion: k8s.patoarvizu.dev/v1alpha1
kind: KMSVaultSecret
metadata:
  name: example-kmsvaultsecret
  namesace: default
spec:
  path: secret/test/kms-vault-secret
  vaultAuthMethod: <auth-method>
  kvSettings:
    engineVersion: v1
  secrets:
    - key: test
      encryptedSecret: <kms-encrypted-secret>
```

Make sure you also set the appropriate `vaultAuthMethod` based on your setup. After that, assuming the resource is in `deploy/example-kms-vault-secret.yaml`, just run:
```
kubectl apply -f deploy/example-kms-vault-secret.yaml
```

## Important notes by this project

#### Kubernetes namespaces and Vault namespaces

The `KMSVaultSecret` CRD is a namespaced resource, but please note that the [Kubernetes namespace](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/) doesn't map to a [Vault namespace](https://www.vaultproject.io/docs/enterprise/namespaces/index.html). Support for Vault namespaces is outside of the scope of this project, since they're only available in Vault enterprise for now.

#### Multiple secrets writing to the same location

Also, the operator doesn't make any guarantees or checks about `KMSVaultSecret`s in different namespaces writing to the same Vault paths. The operator is designed to **continuously** write the secret, so if two or more resources are pointing to the same location, the operator will constantly overwrite them.

#### No validation on target path

Because the controller is designed to write the secret to Vault continuously, it doesn't perform any validation on what may exist on the configured path before writing to it. Be careful when deploying a `KMSVaultSecret` to make sure you don't overwrite your existing secrets.

#### Removing secrets when a `KMSVaultSecret` is deleted.

The kms-vault-operator controller doesn't implement any [Kubernetes finalizers](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers), so deleting the custom resource won't delete the secret from Vault.

#### KMS encryption context is not supported (as of this version)

The operator or `KMSVaultSecret` CRD doesn't support passing [encryption context](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#encrypt_context) to the `Decrypt` operation, so only secrets encrypted without encryption context are supported.

#### Support for K/V V2 is limited (as of this version)

The `KMSVaultSecret` CRD supports specifying `kvSettings.engineVersion: v2` and a check-and-set index with `kvSettings.casIndex` but support for it is limited. For example, the operator doesn't doesn't enforce or validate that the `path` is V2-friendly, and no metadata operations are available.

## Help wanted!

I (the author of this operator) am not a "real" golang developer and the code probably shows it. It's also my first venture into writing a Kubernetes operator. I'd welcome any Issues or PRs on this repo, including (and specially) support for additional Vault authentication methods.