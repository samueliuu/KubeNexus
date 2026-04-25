#!/bin/sh
set -e

KUBENEXUS_SERVER="${1:-}"
CLUSTER_TOKEN="${2:-}"

if [ -z "$KUBENEXUS_SERVER" ] || [ -z "$CLUSTER_TOKEN" ]; then
    echo "Usage: $0 <KUBENEXUS_SERVER_URL> <CLUSTER_TOKEN>"
    echo "Example: $0 https://kubenexus.example.com cn-xxxx-xxxx-xxxx"
    exit 1
fi

echo "=== KubeNexus K3s + Agent Installer ==="

if ! command -v k3s >/dev/null 2>&1; then
    echo "[1/3] Installing K3s..."
    curl -sfL https://get.k3s.io | sh -
    echo "K3s installed successfully."
    mkdir -p ~/.kube
    cp /etc/rancher/k3s/k3s.yaml ~/.kube/config 2>/dev/null || true
else
    echo "[1/3] K3s already installed, skipping."
fi

echo "[2/3] Waiting for K3s to be ready..."
max_wait=60
waited=0
while ! kubectl get nodes >/dev/null 2>&1; do
    sleep 2
    waited=$((waited + 2))
    if [ $waited -ge $max_wait ]; then
        echo "K3s not ready after ${max_wait}s, exiting."
        exit 1
    fi
done
echo "K3s is ready."

echo "[3/3] Deploying KubeNexus Agent..."
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: kubenexus-system
---
apiVersion: v1
kind: Secret
metadata:
  name: kubenexus-agent-config
  namespace: kubenexus-system
type: Opaque
stringData:
  SERVER_URL: "${KUBENEXUS_SERVER}"
  CLUSTER_TOKEN: "${CLUSTER_TOKEN}"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubenexus-agent
  namespace: kubenexus-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubenexus-agent
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubenexus-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubenexus-agent
subjects:
- kind: ServiceAccount
  name: kubenexus-agent
  namespace: kubenexus-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubenexus-agent
  namespace: kubenexus-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubenexus-agent
  template:
    metadata:
      labels:
        app: kubenexus-agent
    spec:
      serviceAccountName: kubenexus-agent
      containers:
      - name: agent
        image: kubenexus/agent:latest
        env:
        - name: SERVER_URL
          valueFrom:
            secretKeyRef:
              name: kubenexus-agent-config
              key: SERVER_URL
        - name: CLUSTER_TOKEN
          valueFrom:
            secretKeyRef:
              name: kubenexus-agent-config
              key: CLUSTER_TOKEN
        resources:
          limits:
            cpu: 200m
            memory: 256Mi
          requests:
            cpu: 50m
            memory: 64Mi
EOF

echo ""
echo "=== Installation Complete ==="
echo "Agent deployed to kubenexus-system namespace."
echo "Check status: kubectl get pods -n kubenexus-system"
