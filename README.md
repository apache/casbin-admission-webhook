# Casbin Admission Webhook

[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/casbin-admission-webhook)](https://goreportcard.com/report/github.com/casbin/casbin-admission-webhook)
[![Build](https://github.com/casbin/casbin-admission-webhook/workflows/CI/badge.svg)](https://github.com/casbin/casbin-admission-webhook/actions)
[![Godoc](https://pkg.go.dev/badge/github.com/casbin/casbin-admission-webhook.svg)](https://pkg.go.dev/github.com/casbin/casbin-admission-webhook)
[![Release](https://img.shields.io/github/release/casbin/casbin-admission-webhook.svg)](https://github.com/casbin/casbin-admission-webhook/releases/latest)
[![Discord](https://img.shields.io/discord/1022748306096537660?logo=discord&label=discord&color=5865F2)](https://discord.gg/S5UjpzGZjN)
[![Sourcegraph](https://sourcegraph.com/github.com/casbin/casbin-admission-webhook/-/badge.svg)](https://sourcegraph.com/github.com/casbin/casbin-admission-webhook?badge)

A Kubernetes Admission Webhook that uses [Casbin](https://casbin.org/) for Policy-as-Code enforcement. This webhook validates Kubernetes API requests based on Casbin policies, enabling fine-grained access control over cluster resources.

## Features

- 🔐 **Policy-based Access Control**: Use Casbin's powerful policy engine for Kubernetes admission control
- 🚀 **Easy Deployment**: Simple Kubernetes manifests and automated certificate generation
- 🧪 **Well-tested**: Comprehensive unit tests and integration tests
- 📦 **Container-ready**: Pre-built Docker images available
- 🔄 **CI/CD**: Automated releases with semantic versioning
- 🛡️ **Secure**: Runs with minimal privileges and read-only root filesystem

## How It Works

The Casbin Admission Webhook acts as a Kubernetes ValidatingWebhookConfiguration that intercepts API requests before they are persisted to etcd. Each request is evaluated against Casbin policies to determine if it should be allowed or denied.

**Flow:**
1. User makes a Kubernetes API request (e.g., create a pod)
2. Kubernetes API server sends an AdmissionReview to the webhook
3. Webhook evaluates the request against Casbin policies
4. Webhook responds with Allow/Deny decision
5. If allowed, Kubernetes proceeds with the request

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.16+)
- kubectl configured to access your cluster
- Docker (for building custom images)

### Installation

1. **Generate TLS certificates and deploy:**

```bash
# Clone the repository
git clone https://github.com/casbin/casbin-admission-webhook.git
cd casbin-admission-webhook

# Generate certificates and deploy to Kubernetes
make deploy
```

This will:
- Create the `casbin-system` namespace
- Generate TLS certificates
- Deploy the webhook server
- Configure the ValidatingWebhookConfiguration

2. **Verify the deployment:**

```bash
kubectl get pods -n casbin-system
kubectl logs -n casbin-system -l app=casbin-admission-webhook
```

### Configuration

#### Casbin Model

The Casbin model defines the request format and matching rules. Default model (`deploy/kubernetes/model.conf`):

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && (r.obj == p.obj || p.obj == "*") && (r.act == p.act || p.act == "*")
```

#### Casbin Policy

The policy file defines access rules (`deploy/kubernetes/policy.csv`):

```csv
# Format: p, subject, object, action
# Allow admin to do everything
p, admin, *, *

# Allow developer to create/update/delete pods in development namespace
p, developer, pods/development, CREATE
p, developer, pods/development, UPDATE
p, developer, pods/development, DELETE

# Allow viewer to get resources
p, viewer, */*, GET
```

**Request Format:**
- `subject`: Kubernetes username (from UserInfo)
- `object`: `{resource}/{namespace}` (e.g., `pods/default`)
- `action`: Kubernetes operation (CREATE, UPDATE, DELETE, etc.)

#### Customize Policies

Edit the ConfigMap to update policies:

```bash
kubectl edit configmap casbin-config -n casbin-system
# Restart pods to reload configuration
kubectl rollout restart deployment casbin-admission-webhook -n casbin-system
```

## Development

### Build from Source

```bash
# Install dependencies
go mod download

# Run tests
make test

# Build binary
make build

# Run locally (requires valid certs and config)
make run
```

### Run Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
make coverage
```

### Build Docker Image

```bash
# Build image
make docker-build

# Build and push (requires Docker Hub credentials)
make docker-push VERSION=v1.0.0
```

## Deployment Options

### Using kubectl

```bash
# Deploy everything
kubectl apply -f deploy/kubernetes/deployment.yaml
kubectl apply -f deploy/kubernetes/webhook-config.yaml
```

### Manual Certificate Generation

If you prefer to manage certificates yourself:

```bash
./deploy/kubernetes/generate-certs.sh casbin-system casbin-admission-webhook
```

### Environment Variables

The webhook supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `WEBHOOK_PORT` | HTTPS server port | `8443` |
| `TLS_CERT_FILE` | Path to TLS certificate | `/etc/webhook/certs/tls.crt` |
| `TLS_KEY_FILE` | Path to TLS private key | `/etc/webhook/certs/tls.key` |
| `CASBIN_MODEL_FILE` | Path to Casbin model file | `/etc/webhook/casbin/model.conf` |
| `CASBIN_POLICY_FILE` | Path to Casbin policy file | `/etc/webhook/casbin/policy.csv` |

## Examples

### Example 1: Role-based Access Control

```csv
# Admins can do everything
p, system:serviceaccount:kube-system:admin, *, *

# Developers can manage pods in dev namespace
p, developer@example.com, pods/dev, CREATE
p, developer@example.com, pods/dev, UPDATE
p, developer@example.com, pods/dev, DELETE

# Viewers can only read
p, viewer@example.com, */*, GET
```

### Example 2: Namespace Isolation

```csv
# Team A can only access team-a namespace
p, team-a, */team-a, *

# Team B can only access team-b namespace
p, team-b, */team-b, *
```

### Example 3: Resource Type Restrictions

```csv
# Allow creating configmaps but not secrets
p, developer, configmaps/*, CREATE
p, developer, secrets/*, ""

# Allow reading everything
p, developer, */*, GET
```

## Troubleshooting

### Webhook not receiving requests

1. Check webhook configuration:
   ```bash
   kubectl get validatingwebhookconfigurations casbin-admission-webhook -o yaml
   ```

2. Verify service endpoints:
   ```bash
   kubectl get endpoints -n casbin-system
   ```

3. Check webhook logs:
   ```bash
   kubectl logs -n casbin-system -l app=casbin-admission-webhook -f
   ```

### Certificate issues

Regenerate certificates:
```bash
./deploy/kubernetes/generate-certs.sh casbin-system casbin-admission-webhook
kubectl rollout restart deployment casbin-admission-webhook -n casbin-system
```

### Policy not working as expected

1. Enable verbose logging by checking pod logs
2. Verify policy syntax in the ConfigMap
3. Test policy locally using Casbin's online editor: https://casbin.org/editor

## Uninstallation

```bash
make undeploy
```

Or manually:
```bash
kubectl delete -f deploy/kubernetes/webhook-config.yaml
kubectl delete -f deploy/kubernetes/deployment.yaml
kubectl delete namespace casbin-system
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Casbin](https://casbin.org/) - The authorization library that powers this webhook
- [Kubernetes](https://kubernetes.io/) - The container orchestration platform

## Related Projects

- [casbin/casbin](https://github.com/casbin/casbin) - An authorization library that supports access control models like ACL, RBAC, ABAC
- [kubernetes-sigs/controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - Kubernetes controller runtime

## Support

- 📖 [Documentation](https://github.com/casbin/casbin-admission-webhook)
- 💬 [Discussions](https://github.com/casbin/casbin-admission-webhook/discussions)
- 🐛 [Issue Tracker](https://github.com/casbin/casbin-admission-webhook/issues)
- 📧 [Casbin Forum](https://forum.casbin.com/)
