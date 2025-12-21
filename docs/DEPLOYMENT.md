# Deployment Guide

This guide provides step-by-step instructions for deploying the Casbin Admission Webhook to your Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (v1.16 or later)
- kubectl configured to access your cluster
- Cluster admin permissions
- openssl (for certificate generation)

## Quick Deployment

The fastest way to deploy is using the Makefile:

```bash
make deploy
```

This will:
1. Create the `casbin-system` namespace
2. Generate TLS certificates
3. Deploy all necessary resources
4. Configure the ValidatingWebhookConfiguration

## Manual Deployment

If you prefer to deploy manually or customize the deployment:

### Step 1: Create Namespace

```bash
kubectl create namespace casbin-system
```

### Step 2: Generate TLS Certificates

The webhook requires TLS certificates to communicate securely with the Kubernetes API server.

```bash
cd deploy/kubernetes
./generate-certs.sh casbin-system casbin-admission-webhook
```

### Step 3: Deploy the Webhook

```bash
kubectl apply -f deploy/kubernetes/deployment.yaml
```

### Step 4: Verify Deployment

```bash
kubectl get pods -n casbin-system
kubectl logs -n casbin-system -l app=casbin-admission-webhook
```

### Step 5: Deploy Webhook Configuration

```bash
kubectl apply -f deploy/kubernetes/webhook-config.yaml
```

## Customizing Policies

Edit the ConfigMap to change policies:

```bash
kubectl edit configmap casbin-config -n casbin-system
kubectl rollout restart deployment casbin-admission-webhook -n casbin-system
```

## Uninstallation

```bash
make undeploy
```
