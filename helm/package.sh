#!/bin/bash

helm package helm/kms-vault-operator/
version=$(cat helm/kms-vault-operator/Chart.yaml | yaml2json | jq -r '.version')
mv kms-vault-operator-$version.tgz docs/
helm repo index docs --url https://patoarvizu.github.io/kms-vault-operator
helm-docs
mv helm/kms-vault-operator/README.md docs/index.md
git add docs/