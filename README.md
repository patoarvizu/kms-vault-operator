# KMS Vault operator

![Black Lives Matter](https://img.shields.io/badge/BLM-Black%20Lives%20Matter-black)
![CircleCI](https://img.shields.io/circleci/build/github/patoarvizu/kms-vault-operator.svg?label=CircleCI) ![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/patoarvizu/kms-vault-operator.svg) ![Docker Pulls](https://img.shields.io/docker/pulls/patoarvizu/kms-vault-operator.svg) ![Keybase BTC](https://img.shields.io/keybase/btc/patoarvizu.svg) ![Keybase PGP](https://img.shields.io/keybase/pgp/patoarvizu.svg) ![GitHub](https://img.shields.io/github/license/patoarvizu/kms-vault-operator.svg)

<!-- TOC -->

- [Intro](#intro)
- [Description](#description)
- [Configuration](#configuration)
  - [AWS](#aws)
  - [Vault](#vault)
    - [Kubernetes authentication method (`--vault-authentication-method=k8s`)](#kubernetes-authentication-method---vault-authentication-methodk8s)
    - [Vault token authentication method (`--vault-authentication-method=token`)](#vault-token-authentication-method---vault-authentication-methodtoken)
    - [Vault userpass authentication method (`--vault-authentication-method=userpass`)](#vault-userpass-authentication-method---vault-authentication-methoduserpass)
    - [Vault approle authentication method (`--vault-authentication-method=approle`)](#vault-approle-authentication-method---vault-authentication-methodapprole)
    - [Vault github authentication method (`--vault-authentication-method=github`)](#vault-github-authentication-method---vault-authentication-methodgithub)
    - [Vault iam authentication method (`--vault-authentication-method=iam`)](#vault-iam-authentication-method---vault-authentication-methodiam)
  - [Command-line flags](#command-line-flags)
  - [Creating a secret](#creating-a-secret)
  - [Partial secrets](#partial-secrets)
  - [Empty secrets](#empty-secrets)
  - [Validating webhook](#validating-webhook)
    - [Auto-reloading certificate](#auto-reloading-certificate)
  - [Monitoring](#monitoring)
- [For security nerds](#for-security-nerds)
  - [Docker images are signed and published to Docker Hub's Notary server](#docker-images-are-signed-and-published-to-docker-hubs-notary-server)
  - [Docker images are labeled with Git and GPG metadata](#docker-images-are-labeled-with-git-and-gpg-metadata)
- [Multi-architecture images](#multi-architecture-images)
- [Important notes by this project](#important-notes-by-this-project)
  - [Kubernetes namespaces and Vault namespaces](#kubernetes-namespaces-and-vault-namespaces)
  - [Multiple secrets writing to the same location](#multiple-secrets-writing-to-the-same-location)
  - [No validation on target path](#no-validation-on-target-path)
  - [Removing secrets when a `KMSVaultSecret` is deleted.](#removing-secrets-when-a-kmsvaultsecret-is-deleted)
  - [Decryption or decoding errors are ignored (but logged)](#decryption-or-decoding-errors-are-ignored-but-logged)
  - [Support for K/V V2 is limited (as of this version)](#support-for-kv-v2-is-limited-as-of-this-version)
  - [Partial secrets don't validate keys](#partial-secrets-dont-validate-keys)
  - [Partial secrets don't support finalizers (yet)](#partial-secrets-dont-support-finalizers-yet)
- [Help wanted!](#help-wanted)

<!-- /TOC -->

## Intro

We all know (or should know) that keeping secrets in plain text in source control is definitely a big no-no. That forces us to keep sensitive text in a different way like a password manager, or worse, in an encrypted file in the same repository (but where do we keep the password for that?!). So we either have a "first secret" problem, or a "source of truth" problem, or an automation problem.

[KMS](https://aws.amazon.com/kms/) provides a nice solution to the encryption/decryption problem so we can keep encrypted secrets in source control (and thus closer to the code that depends on it). But more often than not, said secret is not very valuable by itself and needs to be decrypted and put where it can be consumed at runtime, e.g. in [Vault](https://www.vaultproject.io).

One pattern that can accomplish this is KMS + Vault + [Terraform](https://www.terraform.io/). By creating a KMS key (which can be done with Terraform), we can encrypt a secret that can be safely stored in source control. It can then be decrypted by a Terraform [`aws_kms_secrets`](https://www.terraform.io/docs/providers/aws/d/kms_secrets.html) data source and passed to a [`vault_generic_secret`](https://www.terraform.io/docs/providers/vault/r/generic_secret.html) resource to be written into Vault. Top it off with the [`inmem`](https://www.vaultproject.io/docs/configuration/storage/in-memory.html) storage backend and the secret will never be in plaintext in durable storage.

While this gets the job done for storing secrets in a secure way, it doesn't lend itself to automation. Options like Terraform Enterprise are often prohibitively expensive, and ad-hoc solutions like running `terraform apply` on a `cron` job are fragile and insecure.

The goal of this operator is to minimize the security risk of secret exposure, while leveraging Kubernetes to provide the automation to minimize configuration drift.

## Description

This operator manages `KMSVaultSecret` CRDs containing secrets encrypted with a KMS key (and base64-encoded) and stores them in Vault. This allows you to securely manage your Vault secrets in source control and because the decryption only happens at runtime, it's only in the operator memory, minimizing the exposure of the plain text secret. The operator resources and scaffolding code are managed and generated automatically with the [operator-sdk framework](https://github.com/operator-framework/operator-sdk).

The AWS credentials to do the decryption operation use [aws-sdk-go](https://github.com/aws/aws-sdk-go) and follows the default [credential precedence order](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html), so if you're going to inject environment variables or configuration files, you should do so on the operator's `Deployment` manifest (e.g. [deploy/operator.yaml](deploy/operator.yaml)).

Each secret will define a `path`, a set of `kvSettings`, and a list of `secrets`. Each `secret` will have a `key` and `encryptedSecret` field of type `string`, and a `secretContext` field that's an arbitrary set of key-value pairs corresponding to the [encryption context](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#encrypt_context) with which the secret was encrypted.

This first version of the operator supports authenticating to Vault via the [Kubernetes auth method](https://www.vaultproject.io/docs/auth/kubernetes.html), the [Userpass authe method](https://www.vaultproject.io/docs/auth/userpass.html), or directly via a [Vault token](https://www.vaultproject.io/docs/auth/token.html). Support for more authentication methods will be added in the future. Note that the configuration required for the operator to perform KMS and Vault operations is not done on the `KMSVaultSecret` CR but on the operator `Deployment` itself, and is documented below.

## Configuration

### AWS

As stated above, the required configuration to allow the operator to decrypt secrets should be injected via the operator `Deployment` manifest (as `AWS_*` environment variables or `~/.aws/credentials`/`~/.aws/config` files). The aws-sdk-go library will also discover IAM roles at runtime if the operator is running on an EC2 instance with an instance role, in which case no additional configuration should be required on the operator.

### Vault

In addition to the AWS configuration, Vault configuration should be injected too. The common `VAULT_*` [environment variables](https://www.vaultproject.io/docs/commands/#environment-variables) will be read and used by the [client](https://github.com/hashicorp/vault/tree/master/api). The variables that would need to be set will vary depending on your environment, but you'll typically want to at least set `VAULT_ADDR`, and either `VAULT_CACERT`/`VAULT_CAPATH` (pointing to a correspoinging mounted file or directory) or `VAULT_SKIP_VERIFY`. This is all done on the operator `Deployment` manifest.

The following sections document the environment variables used by each authentication method and their defaults. They are used in addition to the AWS and base Vault variables described before.

#### Kubernetes authentication method (`--vault-authentication-method=k8s`)

Environment variable | Required? | Default | Description
---------------------|-----------|---------|------------
`VAULT_K8S_ROLE` | N | `kms-vault-operator` | The Vault role that the client should authenticate as on the Kubernetes login endpoint.
`VAULT_K8S_LOGIN_ENDPOINT` | N | `auth/kubernetes/login` | The Kubernetes authentication endpoint in Vault

#### Vault token authentication method (`--vault-authentication-method=token`)

This method simply follows the convention and uses the `VAULT_TOKEN` environment variable to authenticate.

Environment variable | Required? | Default | Description
---------------------|-----------|---------|------------
`VAULT_TOKEN` | Y | | The Vault token used to perform operations on Vault.

#### Vault userpass authentication method (`--vault-authentication-method=userpass`)

Environment variable | Required? | Default | Description
---------------------|-----------|---------|------------
`VAULT_USERNAME` | Y | | The Vault username to use for authentication
`VAULT_PASSWORD` | Y | | The password corresponding to `VAULT_USERNAME`

#### Vault approle authentication method (`--vault-authentication-method=approle`)

Environment variable | Required? | Default | Description
---------------------|-----------|---------|------------
`VAULT_APPROLE_ROLE_ID` | Y | | The AppRole role id to use for authentication
`VAULT_APPROLE_SECRET_ID` | Y | | The AppRole secret id to use for authentication
`VAULT_APPROLE_ENDPOINT` | N | `auth/approle/login` | The Vault endpoint to use for this authentication method

#### Vault github authentication method (`--vault-authentication-method=github`)

Environment variable | Required? | Default | Description
---------------------|-----------|---------|------------
`VAULT_GITHUB_TOKEN` | Y | | The GitHub token to use for authentication
`VAULT_GITHUB_AUTH_ENDPOINT` | N | `auth/github/login` | The Vault endpoint to use for this authentication method

#### Vault iam authentication method (`--vault-authentication-method=iam`)

This authentication method uses [Vault AWS auth method](https://www.vaultproject.io/docs/auth/aws), but specifically the `iam` method only. Support for the `ec2` method might be added in the future.

Since the operator itself is already assumed to be running with some form of access to IAM credentials for decrypting operations (via environment variables, [kube2iam](https://github.com/jtblin/kube2iam), [kiam](https://github.com/uswitch/kiam), etc.), those same credentials can be be used by default to log in to Vault using the `aws` method. However, you can also inject static credentials via the environment variables below if wish to keep decryption credentials separate from login credentials.

Environment variable | Required? | Default | Description
---------------------|-----------|---------|------------
`VAULT_IAM_AWS_ACCESS_KEY_ID` | N | No default, but if not specified, a dynamic access key id will be retrieved at runtime using the standard AWS credentials provider chain, assuming one is available. | A static value to use as the access key ID for Vault login purposes.
`VAULT_IAM_AWS_SECRET_ACCESS_KEY` | N | No default, but if not specified, a dynamic secret access key will be retrieved at runtime using the standard AWS credentials provider chain, assuming one is available. | A static value to use as the secret access key for Vault login purposes.
`VAULT_IAM_ROLE` | N | No default, but if a role is not specified, Vault will try to guess the role name based on the principal name associated with the credentials (e.g. the IAM user name or role). | The name of a Vault role to assume with the IAM credentials provided. **This is the Vault role that must be previously configured, not the IAM role you may be authenticating with.**
`VAULT_IAM_AUTH_ENDPOINT` | N | `auth/aws/login` | The Vault endpoint to use for this authentication method

**NOTE:** the remote Vault instance will also require runtime permissions to perform the IAM validation actions. Those credentials cannot be set by the operator and must be set directly in the target Vault cluster by other means. Refer to the official Vault [documentation](https://www.vaultproject.io/docs/auth/aws#recommended-vault-iam-policy) for the recommended IAM policy.

### Command-line flags

Flag | Default | Description
-----|---------|------------
`--vault-authentication-method` | `token` | Method to be used for the controller to authenticate with Vault.
`--sync-period-seconds` | 120 | Amount of time in seconds to wait between before syncing the secret to Vault

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
  kvSettings:
    engineVersion: v1
  secrets:
    - key: test
      encryptedSecret: <kms-encrypted-secret>
```

After that, assuming the resource is in `deploy/example-kms-vault-secret.yaml`, just run:
```
kubectl apply -f deploy/example-kms-vault-secret.yaml
```

### Partial secrets

In addition to managing `KMSVaultSecret` custom resources, this operator also handles a second type of resource called `PartialKMSVaultSecret`. This CRD is similar to `KMSVaultSecret` but only supports the `secrets` field, and doesn't have its own controller. Instead, the purpose of this resource is to hold secrets that can be included in a `KMSVaultSecret`, via the `includeSecrets` field. The single `kmsvaultsecret_controller.go` will aggregate the included secrets along with those of the resource itself and write them all together as a single item in Vault. To keep things as simple as possible, the first iteration of this feature won't support nesting `PartialKMSVaultSecret`s (e.g. by including `PartialKMSVaultSecret`s in other `PartialKMSVaultSecret`s). Rather, the way to include multiple partial secrets is to just list them all in the `includeSecrets` field of the `KMSVaultSecret` resource.

Because of their abstract nature, `PartialKMSVaultSecret`s don't have a path, Vault authenticating method, or KV settings, but they do support the KMS secret encryption context, which is passed down to the concrete `KMSVaultSecret` object.

### Empty secrets

Although rarely an empty string is required as a secret, sometimes it is needed for backwards compatibility or as a placeholder. Since an empty string is not a valid KMS-encrypted string, the CRD includes a field that signals to the operator that an empty string should be put in the indicated path and field. To do this, simply set `emptySecret: true` to each individual item under `secrets` that you want to inject as a an empty string. Note that when you do this, the operator will ignore anything set in the `encryptedSecret` field, even if it's a valid KMS-encrypted string.

### Validating webhook

The Docker image contains another binary (`kms-vault-validating-webhook`) that can be used as a server that a `ValidatingWebhookConfiguration` calls to validate either `KMSVaultSecret`s or `PartialKMSVaultSecret`s and prevent them from being picked up by the controller in the first place. Since this binary is separate from the main one, it would need to be deployed either as a sidecar or as a separate `Deployment`, as well as requiring its own `Service`. You can find an example of how to deploy it as a sidecar [here](deploy/operator.yaml).

Keep in mind that a `ValidatingWebhookConfiguration` requires a valid CA bundle to trust the webhook over TLS. While this can be any certificate generated offline, you can also use [`cert-manager`](https://github.com/jetstack/cert-manager/) to make it easy to generate certificates as Kubernetes `Secret`s and mount them on containers (like the webhook), or to inject the corresponding CA bundle in `ValidatingWebhookConfiguration`s.

#### Auto-reloading certificate

The webhook performs a hot reload if the underlying TLS certificate (indicated by the `-tls-cert-file` flag) on disk is modified. This is helpful when using automatic certificate provisioners like cert-manager that will do automatic rotation of the certificates but can't control the lifecycle of the workloads using the certificate.

The way this is achieved is by initially loading the certificate and keeping it in a local cache, then using the [radovskyb/watcher](https://github.com/radovskyb/watcher) library to watch for changes on the file and updating the cached version if the file changes.

### Monitoring

~~If your Kubernetes cluster is running the Prometheus [operator](https://github.com/coreos/prometheus-operator), this operator will automatically create an additional `Service` called `kms-vault-operator-metrics` and a corresponding `ServiceMonitor` of the same name. This monitor will scrape the operator for metrics on two different ports. Port 8383 will post general metrics about the running process, while port 8686 will post metrics about the custom resources managed by the operator. More information can be found on the Operator SDK [website](https://sdk.operatorframework.io/docs/golang/monitoring/prometheus/).~~

Up until version `v0.14.0`, this operator was using a version of the operator-sdk that supported automatic creation a `Service` and `ServiceMonitor` objects to scrape Prometheus metrics, but that functionality has been removed. If you're running the Prometheus operator in your cluster and you want to scrape metrics for this operator, you're going to have to explicitly create them yourself, querying the `/metrics` endpoint on port `:8080`.

## For security nerds

**NOTE:** Due to technical issues with the Notary client, starting on January 4th 2023 and until further notice new images will NOT be signed. The images will still be built for multi-architecture, and will include the Git and GPG metadata, but they won't pass Docker Content Trust validation if you have it enabled.

### Docker images are signed and published to Docker Hub's Notary server

The [Notary](https://github.com/theupdateframework/notary) project is a CNCF incubating project that aims to provide trust and security to software distribution. Docker Hub runs a Notary server at https://notary.docker.io for the repositories it hosts.

[Docker Content Trust](https://docs.docker.com/engine/security/trust/content_trust/) is the mechanism used to verify digital signatures and enforce security by adding a validating layer.

You can inspect the signed tags for this project by doing `docker trust inspect --pretty docker.io/patoarvizu/kms-vault-operator`, or (if you already have `notary` installed) `notary -d ~/.docker/trust/ -s https://notary.docker.io list docker.io/patoarvizu/kms-vault-operator`.

If you run `docker pull` with `DOCKER_CONTENT_TRUST=1`, the Docker client will only pull images that come from registries that have a Notary server attached (like Docker Hub).

### Docker images are labeled with Git and GPG metadata

In addition to the digital validation done by Docker on the image itself, you can do your own human validation by making sure the image's content matches the Git commit information (including tags if there are any) and that the GPG signature on the commit matches the key on the commit on github.com.

For example, if you run `docker pull patoarvizu/kms-vault-operator:898c158e9c3c313984d9a0c7947f0f3178501cf2` to pull the image tagged with that commit id, then run `docker inspect patoarvizu/kms-vault-operator:898c158e9c3c313984d9a0c7947f0f3178501cf2 | jq -r '.[0].ContainerConfig.Labels'` (assuming you have [jq](https://stedolan.github.io/jq/) installed) you should see that the `GIT_COMMIT` label matches the tag on the image. Furthermore, if you go to https://github.com/patoarvizu/kms-vault-operator/commit/898c158e9c3c313984d9a0c7947f0f3178501cf2 (notice the matching commit id), and click on the **Verified** button, you should be able to confirm that the GPG key ID used to match this commit matches the value of the `SIGNATURE_KEY` label, and that the key belongs to the `AUTHOR_EMAIL` label. When an image belongs to a commit that was tagged, it'll also include a `GIT_TAG` label, to further validate that the image matches the content.

Keep in mind that this isn't tamper-proof. A malicious actor with access to publish images can create one with malicious content but with values for the labels matching those of a valid commit id. However, when combined with Docker Content Trust, the certainty of using a legitimate image is increased because the chances of a bad actor having both the credentials for publishing images, as well as Notary signing credentials are significantly lower and even in that scenario, compromised signing keys can be revoked or rotated.

Here's the list of included Docker labels:

- `AUTHOR_EMAIL`
- `COMMIT_TIMESTAMP`
- `GIT_COMMIT`
- `GIT_TAG`
- `SIGNATURE_KEY`

## Multi-architecture images

Manifests published with the semver tag (e.g. `patoarvizu/kms-vault-operator:v0.15.0`), as well as `latest` are multi-architecture manifest lists. In addition to those, there are architecture-specific tags that correspond to an image manifest directly, tagged with the corresponding architecture as a suffix, e.g. `v0.15.0-amd64`. Both types (image manifests or manifest lists) are signed with Notary as described above.

Here's the list of architectures the images are being built for, and their corresponding suffixes for images:

- `linux/amd64`, `-amd64`
- `linux/arm64`, `-arm64`
- `linux/arm/v7`, `-arm7`

## Important notes by this project

### Kubernetes namespaces and Vault namespaces

The `KMSVaultSecret` CRD is a namespaced resource, but please note that the [Kubernetes namespace](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/) doesn't map to a [Vault namespace](https://www.vaultproject.io/docs/enterprise/namespaces/index.html). Support for Vault namespaces is outside of the scope of this project, since they're only available in Vault enterprise for now.

Additionally, the `PartialKMSVaultSecret` CRD is also namespaced, but the `includeSecrets` field on a `KMSVaultSecret` object will only discover partial secrets within the same namespace. Support for referencing secrets from another namespace will be added in a future version.

### Multiple secrets writing to the same location

Also, the operator doesn't make any guarantees or checks about `KMSVaultSecret`s in different namespaces writing to the same Vault paths. The operator is designed to **continuously** write the secret, so if two or more resources are pointing to the same location, the operator will constantly overwrite them.

### No validation on target path

Because the controller is designed to write the secret to Vault continuously, it doesn't perform any validation on what may exist on the configured path before writing to it. Be careful when deploying a `KMSVaultSecret` to make sure you don't overwrite your existing secrets.

### Removing secrets when a `KMSVaultSecret` is deleted.

The kms-vault-operator controller supports removing secrets from Vault by setting `delete.k8s.patoarvizu.dev` as a [Kubernetes finalizer](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers). Support for this for K/V V1 is simple since secrets are not versioned, but when the secret is for K/V V2, deleting a `KMSVaultSecret` object will delete **ALL** of its versions and metadata from Vault, so handle it with care. If the secret is V2, the path for the `DELETE` operation is the same as the input one, replacing `secret/data/` with `secret/metadata/`. There is currently no support for removing a single version of a K/V V2 secret.

### Decryption or decoding errors are ignored (but logged)

If the [validating webhook](#validating-webhook) mentioned above is deployed, then the controller won't (in theory) need to deal with erroneous secrets since they're never committed to storage. However, if the webhook is not in place, and a secret is incorrectly encoded or encrypted (including if the encryption context doesn't match the secret), the operator will log the error, skip those secrets, and continue writing the rest. This applies to individual items in the `secrets` list, i.e. the controller will still apply other secrets within the same `KMSVaultSecret` even if one of them fails. The controller, however, will trigger an event of type `Warning` for each `encryptedSecret` that it wasn't able to decode or decrypt.

### Support for K/V V2 is limited (as of this version)

The `KMSVaultSecret` CRD supports specifying `kvSettings.engineVersion: v2` and a check-and-set index with `kvSettings.casIndex` but support for it is limited. For example, the operator doesn't doesn't enforce or validate that the `path` is V2-friendly, and no metadata operations are available.

### Partial secrets don't validate keys

The operator doesn't do any validation in regards to keys included both in the `KMSVaultSecret` object and in any included `PartialKMSVaultSecret`s (or if they exist in multiple included `PartialKMSVaultSecret`s), and it also doesn't provide any guarantees regarding the order of precedence of secrets. Because of this, you should make sure that your secret aggregation avoids these overlaps. Support for validating this can be added in a future version.

### Partial secrets don't support finalizers (yet)

Because `PartialKMSVaultSecret`s don't have their own controller (as of the latest version), it's not possible to handle finalizers on them (specifically the `delete.k8s.patoarvizu.dev` finalizer). That means that the presence of finalizers on those objects won't have the same effect as setting them on `KMSVaultSecret` objects. If a `PartialKMSVaultSecret` object is deleted, the direct effect is that any `KMSVaultSecret` that includes a deleted `PartialKMSVaultSecret` could see the included keys deleted from Vault! This would only happen if the Vault KV secret backend is v1, since to update v2 secrets you need to update the `casIndex` field. Supporting finalizers for partial secrets could be added to a future version.

## Help wanted!

I (the author of this operator) am not a "real" golang developer and the code probably shows it. It's also my first venture into writing a Kubernetes operator. I'd welcome any Issues or PRs on this repo, including (and specially) support for additional Vault authentication methods.