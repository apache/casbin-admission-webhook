#!/bin/bash

set -e

NAMESPACE=${1:-casbin-system}
SERVICE_NAME=${2:-casbin-admission-webhook}

# Create a temporary directory for certificates
TMPDIR=$(mktemp -d)
echo "Temporary directory: $TMPDIR"

# Generate CA key and certificate
openssl genrsa -out ${TMPDIR}/ca.key 2048
openssl req -x509 -new -nodes -key ${TMPDIR}/ca.key -sha256 -days 365 -out ${TMPDIR}/ca.crt \
  -subj "/CN=casbin-admission-webhook-ca"

# Generate server key
openssl genrsa -out ${TMPDIR}/tls.key 2048

# Create certificate signing request
cat <<EOF > ${TMPDIR}/csr.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICE_NAME}
DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc
DNS.4 = ${SERVICE_NAME}.${NAMESPACE}.svc.cluster.local
EOF

# Generate certificate signing request
openssl req -new -key ${TMPDIR}/tls.key -out ${TMPDIR}/tls.csr \
  -subj "/CN=${SERVICE_NAME}.${NAMESPACE}.svc" \
  -config ${TMPDIR}/csr.conf

# Sign the certificate with CA
openssl x509 -req -in ${TMPDIR}/tls.csr -CA ${TMPDIR}/ca.crt -CAkey ${TMPDIR}/ca.key \
  -CAcreateserial -out ${TMPDIR}/tls.crt -days 365 -sha256 \
  -extensions v3_req -extfile ${TMPDIR}/csr.conf

# Create Kubernetes secret
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic casbin-webhook-certs \
  --from-file=tls.key=${TMPDIR}/tls.key \
  --from-file=tls.crt=${TMPDIR}/tls.crt \
  --namespace=${NAMESPACE} \
  --dry-run=client -o yaml | kubectl apply -f -

# Get CA bundle in base64
CA_BUNDLE=$(cat ${TMPDIR}/ca.crt | base64 | tr -d '\n')

# Update webhook configuration with CA bundle
cat deploy/kubernetes/webhook-config.yaml | \
  sed "s|caBundle: \"\"|caBundle: ${CA_BUNDLE}|g" | \
  kubectl apply -f -

echo "Certificates created successfully!"
echo "CA Bundle: ${CA_BUNDLE}"

# Clean up
rm -rf ${TMPDIR}
