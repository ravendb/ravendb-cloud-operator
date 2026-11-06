# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |

## What the operator is trusted with

**A RavenDB Cloud API key.** A `CloudAccount` names a Secret holding it. The
operator reads it when it builds a client, keeps it in memory, sends it only as
the `X-Api-Key` header to RavenDB Cloud, and never writes it to status, events
or logs; a log line names a key only by a short hash. No field of a
CloudAccount can send it anywhere else, because the Secret it names could hold
anything: an account that chose its own destination would let whoever can
write CloudAccounts read every Secret in their namespace. A key reaches every cluster in its
account, and RavenDB Cloud has no narrower one, so whoever may create a
CloudCluster next to a CloudAccount can name any cluster in that account that
no other CloudCluster holds: receive its client certificate, reshape it under
`managementPolicy: Full`, and terminate it by deleting the object under
`deletionPolicy: Terminate`. The last two need `creationMode` `CreateOnly` or
`TakeOverOnly` written down, which keeps a manifest from doing either to a
cluster that merely shares its name, but does not stop someone who means to.
Put an account only in a namespace where everyone who can write CloudClusters
may manage every cluster in it.

Writing a CloudAccount is close to reading Secrets in its namespace: the
account can name any key of any Secret there, the operator sends that value to
RavenDB Cloud as the API key, and the account's status says whether the Secret
and the key exist. Grant write on CloudAccounts only where reading those
Secrets would be granted too.

**Write access to Secrets.** Publishing connection details is what the
operator is for. The connection Secret holds each cluster's client
certificate, which RavenDB Cloud issues with Operator clearance: whoever can
read that Secret administers the cluster's databases. Grant read on it as you
would on a database administrator's credential. Regenerating the certificate in
the portal leaves the old one valid; [connecting-applications.md](docs/how-to-use/connecting-applications.md#when-the-certificate-is-renewed)
says how to cut it off.

**Status.** The operator trusts what it wrote in status: `status.productID`
binds an object to a cluster without the checks a first binding runs. Never
grant write on the status subresources (`cloudclusters/status` and the others)
to anyone but the operator.

**A record of who did what.** Every change starts as a Kubernetes object
change, which the API server's audit log records if its policy says so, and
the operator logs each create, take-over and terminate with the product id.
Events expire after an hour, and RavenDB Cloud does not show customers which
tool made a call, so keep the operator's logs where the rest of your audit
trail lives.

## Permissions

| Resource | Verbs | Why |
|---|---|---|
| Secrets | get, list, watch, create, patch | read API keys, publish connection Secrets by server-side apply |
| CloudAccounts, CloudClusters, CloudClusterRestarts | get, list, watch, patch | reconcile them; patch for finalizers and owner references |
| their status | get, update | report |
| CloudClusters' finalizers | update | hold an object until its deletion policy is honoured |
| events.k8s.io events | create, patch | report |

There is no delete on Secrets, which are collected through owner references,
and no create or delete on the operator's own kinds. Both Secret watches are
metadata only, so Secret contents never reach the operator through a watch.

By default the grant is cluster-wide. With `watchNamespace` set, the Helm chart
grants it in that one namespace instead, and that is the install to prefer:
`create` on Secrets lets whoever holds the operator's token create a
service-account token Secret, and with it act as any service account the grant
reaches. Server-side apply needs `create` to write a connection Secret the
first time, so it stays.

## The container

Distroless, one static binary for the manager and one for the kubectl plugin,
no shell, user 65532. Both the Helm chart and the kustomize manifests set
`runAsNonRoot`, `seccompProfile: RuntimeDefault`,
`allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true` and drop all
capabilities. Nothing is written to disk at runtime. The kustomize manifests
label their namespace for the `restricted` Pod Security Standard; for a Helm
install, label the release namespace yourself:

```bash
kubectl label namespace ravendb-cloud-operator-system pod-security.kubernetes.io/enforce=restricted
```

Every release attests its image, `install.yaml` and the kubectl plugin;
[install.md](docs/how-to-use/install.md#checking-what-you-install) shows how to
check them.

Metrics are off by default; turned on, they are served over HTTPS and require
authentication and authorization.

Answers from RavenDB Cloud are capped in size before they are decoded, and the
certificate archive is unpacked in memory into a fixed set of keys, with every
entry and the archive capped, so no entry name reaches a file path.

## Reporting a Vulnerability
Please report security vulnerabilities to security@ravendb.net
