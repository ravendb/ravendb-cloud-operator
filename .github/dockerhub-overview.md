# RavenDB Cloud Kubernetes Operator

`ravendb/ravendb-cloud-operator` is the controller image for the **RavenDB Cloud Kubernetes Operator**. It runs inside your Kubernetes cluster and manages RavenDB Cloud clusters as Kubernetes objects: it creates a cluster or takes over one that already exists, applies changes to its size, and publishes its URLs and client certificate in a Secret your applications mount.

> **This is the operator, not the database.** Your clusters run in [RavenDB Cloud](https://cloud.ravendb.net); the operator talks to the RavenDB Cloud API with your account's API key. For RavenDB on your own Kubernetes nodes, see [`ravendb/ravendb-operator`](https://hub.docker.com/r/ravendb/ravendb-operator). You do not run this image with `docker run`; it is deployed into Kubernetes with Helm or `install.yaml`.

## Supported tags

- `latest`, and semantic-version tags such as `1.0.0`
- Multi-arch: `linux/amd64`, `linux/arm64`

## Install

```bash
helm repo add ravendb-cloud-operator https://ravendb.github.io/ravendb-cloud-operator/helm
helm install ravendb-cloud-operator ravendb-cloud-operator/ravendb-cloud-operator \
  -n ravendb-cloud-operator-system --create-namespace
```

Or apply `install.yaml` from a [release](https://github.com/ravendb/ravendb-cloud-operator/releases). Then create a `CloudAccount` with your API key and a `CloudCluster`; the [quick start](https://github.com/ravendb/ravendb-cloud-operator#readme) shows both.

## What it does

- `CloudCluster` finds a cluster by display name, takes it over or creates it
- Applies changes to instance type, storage and node count
- Publishes node URLs and the client certificate (PEM and PFX) into a Secret, and renews it
- Keeps the cluster when the object is deleted, unless the manifest says `Terminate`
- `CloudClusterRestart` restarts one node, once
- The image also carries the `kubectl ravendb-cloud` plugin, which exports existing clusters as manifests

Each release is signed with a provenance attestation:

```bash
gh attestation verify oci://docker.io/ravendb/ravendb-cloud-operator:<version> --repo ravendb/ravendb-cloud-operator
```

## Links

- Source and docs: https://github.com/ravendb/ravendb-cloud-operator
- RavenDB Cloud: https://cloud.ravendb.net
- Self-hosted RavenDB on Kubernetes: https://hub.docker.com/r/ravendb/ravendb-operator

The operator is open source (MIT). It needs a RavenDB Cloud account and an API key from its portal.
