# Build the manager binary, and the kubectl plugin beside it. Both base images
# are pinned by digest, which a tag cannot promise, and Dependabot moves them.
FROM golang:1.26@sha256:cfa27959837144bf4649802ba38e59daffd0dd2a7994aa83daaf113285b0f551 AS builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev

WORKDIR /workspace

# Dependencies first, so a source-only change reuses the layer.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# The caches outlive the build, so a one-line change recompiles one package
# rather than everything.
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    LDFLAGS="-s -w -X github.com/ravendb/ravendb-cloud-operator/internal/version.Version=${VERSION}" \
 && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="${LDFLAGS}" -o manager ./cmd \
 && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="${LDFLAGS}" -o kubectl-ravendb_cloud ./cmd/kubectl-ravendb_cloud

# Distroless keeps the runtime free of a shell and a package manager. The
# operator holds a RavenDB Cloud API key in memory, so the smaller the attack
# surface around it, the better.
FROM gcr.io/distroless/static:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
WORKDIR /
COPY --from=builder /workspace/manager .
# The kubectl plugin rides along so its export can run inside the cluster, next
# to the Secret holding the key, with nothing to download: a Job, or kubectl run
# with --overrides, running /kubectl-ravendb_cloud export with
# RAVENDB_CLOUD_API_KEY taken from that Secret through a secretKeyRef. labs/04
# shows the kubectl run form.
COPY --from=builder /workspace/kubectl-ravendb_cloud .
USER 65532:65532

ENTRYPOINT ["/manager"]
