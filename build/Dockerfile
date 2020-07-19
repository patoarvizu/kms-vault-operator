FROM golang:1.13 as builder

COPY . /go/src/github.com/patoarvizu/kms-vault-operator/

WORKDIR /go/src/github.com/patoarvizu/kms-vault-operator/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /kms-vault-validating-webhook /go/src/github.com/patoarvizu/kms-vault-operator/cmd/webhook/main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.2

ARG GIT_COMMIT="unspecified"
LABEL GIT_COMMIT=$GIT_COMMIT

ARG GIT_TAG=""
LABEL GIT_TAG=$GIT_TAG

ARG COMMIT_TIMESTAMP="unspecified"
LABEL COMMIT_TIMESTAMP=$COMMIT_TIMESTAMP

ARG AUTHOR_EMAIL="unspecified"
LABEL AUTHOR_EMAIL=$AUTHOR_EMAIL

ARG SIGNATURE_KEY="undefined"
LABEL SIGNATURE_KEY=$SIGNATURE_KEY

ENV OPERATOR=/usr/local/bin/kms-vault-operator \
    USER_UID=1001 \
    USER_NAME=kms-vault-operator

# install operator binary
COPY build/_output/bin/kms-vault-operator ${OPERATOR}

COPY --from=builder /kms-vault-validating-webhook /usr/local/bin/kms-vault-validating-webhook

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
