#!/bin/bash

# Student Cafe Application Quick Start Script

set -e  # Exit on any error

echo "🚀 Starting Student Cafe Application Deployment..."

# Function to check if a command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ Error: $1 is not installed. Please install it first."
        exit 1
    fi
}

# Check prerequisites
echo "Checking prerequisites..."
check_command minikube
check_command kubectl
check_command helm
check_command docker

# Check if minikube is running
echo "Checking minikube status..."
if ! minikube status > /dev/null 2>&1; then
    echo "Starting Minikube..."
    minikube start --cpus 2 --memory 4096
    echo "Waiting for minikube to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=300s
fi

# Configure Docker to use minikube's docker daemon
echo "Configuring Docker environment..."
eval $(minikube -p minikube docker-env)

# Create namespace if it doesn't exist
echo "Creating student-cafe namespace..."
kubectl create namespace student-cafe --dry-run=client -o yaml | kubectl apply -f -

# Add Helm repositories and deploy infrastructure
echo "Adding Helm repositories..."
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo add kong https://charts.konghq.com
helm repo update

echo "Deploying Consul..."
if helm upgrade --install consul hashicorp/consul \
    --set global.name=consul \
    --namespace student-cafe \
    --set server.replicas=1 \
    --set server.bootstrapExpect=1 \
    --wait --timeout=10m; then
    echo "  Consul helm deployment successful"
else
    echo "❌ Error: Consul helm deployment failed"
    exit 1
fi

echo "Waiting for Consul to be ready..."
if kubectl wait --for=condition=ready pod -l app=consul -n student-cafe --timeout=300s; then
    echo "  Consul pods are ready"
else
    echo "❌ Error: Consul pods failed to become ready within timeout"
    echo "Current Consul pod status:"
    kubectl get pods -l app=consul -n student-cafe
    exit 1
fi

echo "Deploying Kong..."
if helm upgrade --install kong kong/kong \
    --namespace student-cafe \
    --timeout=10m; then
    echo "  Kong helm deployment initiated"
else
    echo "❌ Error: Kong helm deployment failed"
    exit 1
fi

echo "Waiting for Kong pods to be ready..."
if kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=kong -n student-cafe --timeout=300s; then
    echo "  Kong pods are ready"
else
    echo "❌ Error: Kong pods failed to become ready within timeout"
    echo "Current Kong pod status:"
    kubectl get pods -l app.kubernetes.io/name=kong -n student-cafe
    echo "Kong deployment status:"
    kubectl describe deployment kong-kong -n student-cafe | tail -10
    exit 1
fi

echo "Verifying Kong services..."
kubectl wait --for=condition=ready -l app.kubernetes.io/name=kong pod -n student-cafe --timeout=180s

# Wait for Kong proxy service to be available
echo "Waiting for Kong proxy service..."
while ! kubectl get svc kong-kong-proxy -n student-cafe >/dev/null 2>&1; do
    echo "  Waiting for kong-kong-proxy service..."
    sleep 5
done

# Check Kong proxy status endpoint
echo "Verifying Kong proxy status..."
max_attempts=30
attempt=1
while [ $attempt -le $max_attempts ]; do
    if kubectl exec -n student-cafe deployment/kong-kong -c proxy -- kong health >/dev/null 2>&1; then
        echo "  Kong proxy is healthy"
        break
    fi
    echo "  Attempt $attempt/$max_attempts: Kong proxy not ready yet..."
    sleep 10
    attempt=$((attempt + 1))
done

if [ $attempt -gt $max_attempts ]; then
    echo "❌ Warning: Kong proxy not responding after ${max_attempts} attempts"
    echo "   Continuing anyway - Kong may still be functional"
fi

# Build Docker images
echo "Building Docker images..."
echo "Building food-catalog-service..."
docker build -t food-catalog-service:v1 ./food-catalog-service
echo "Building order-service..."
docker build -t order-service:v1 ./order-service
echo "Building cafe-ui..."
docker build -t cafe-ui:v1 ./cafe-ui

# Deploy application services
echo "Deploying application services..."
kubectl apply -f app-deployment.yaml

# Wait for application deployments to start
echo "Waiting for application deployments to start..."
sleep 10

# Verify Kong ingress controller is ready before applying ingress
echo "Verifying Kong ingress controller readiness..."
max_attempts=20
attempt=1
while [ $attempt -le $max_attempts ]; do
    if kubectl get ingressclass kong >/dev/null 2>&1; then
        echo "  Kong ingress class is available"
        break
    fi
    echo "  Attempt $attempt/$max_attempts: Kong ingress class not ready yet..."
    sleep 5
    attempt=$((attempt + 1))
done

if [ $attempt -gt $max_attempts ]; then
    echo "❌ Error: Kong ingress class not available after ${max_attempts} attempts"
    echo "   Cannot proceed with ingress configuration"
    exit 1
fi

# Configure Kong ingress
echo "Configuring Kong ingress..."
if kubectl apply -f kong-ingress.yaml; then
    echo "  Kong ingress configuration applied successfully"
else
    echo "❌ Error: Failed to apply Kong ingress configuration"
    exit 1
fi

# Verify ingress was created
echo "Verifying ingress configuration..."
if kubectl get ingress cafe-ingress -n student-cafe >/dev/null 2>&1; then
    echo "  Ingress 'cafe-ingress' created successfully"
else
    echo "❌ Warning: Ingress 'cafe-ingress' was not created properly"
fi

# Wait for pods to be ready
echo "Waiting for application pods to be ready..."
kubectl wait --for=condition=ready pod -l app=food-catalog-service -n student-cafe --timeout=300s
kubectl wait --for=condition=ready pod -l app=order-service -n student-cafe --timeout=300s
kubectl wait --for=condition=ready pod -l app=cafe-ui -n student-cafe --timeout=300s

# Show pod status
echo "Pod Status:"
kubectl get pods -n student-cafe

# Configure minikube for LoadBalancer access
echo "Configuring LoadBalancer access..."
if command -v minikube >/dev/null 2>&1; then
    # Check if minikube tunnel is needed
    LB_IP=$(kubectl get svc kong-kong-proxy -n student-cafe -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
    if [ -z "$LB_IP" ] || [ "$LB_IP" = "null" ]; then
        echo "  LoadBalancer IP is pending - using minikube service URLs"
        TUNNEL_NEEDED=true
    else
        echo "  LoadBalancer IP available: $LB_IP"
        TUNNEL_NEEDED=false
    fi
else
    echo "  Not running in minikube environment"
    TUNNEL_NEEDED=false
fi

# Get the Kong service URL
echo ""
echo "🎉 Deployment complete!"
echo ""
echo "Access your application at:"

if [ "$TUNNEL_NEEDED" = "true" ]; then
    # Use minikube service to get NodePort URLs
    echo "Getting service URLs via minikube..."
    KONG_HTTP_URL=$(minikube service kong-kong-proxy -n student-cafe --url | head -1)
    KONG_HTTPS_URL=$(minikube service kong-kong-proxy -n student-cafe --url | tail -1)

    echo "HTTP:  $KONG_HTTP_URL"
    echo "HTTPS: $KONG_HTTPS_URL"
    echo ""
    echo "💡 For LoadBalancer access, run in a separate terminal:"
    echo "   minikube tunnel"
    echo "   Then access via: http://127.0.0.1 (requires tunnel)"
else
    # Use LoadBalancer IP if available
    LB_IP=$(kubectl get svc kong-kong-proxy -n student-cafe -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    HTTP_PORT=$(kubectl get svc kong-kong-proxy -n student-cafe -o jsonpath='{.spec.ports[?(@.name=="kong-proxy")].port}')
    HTTPS_PORT=$(kubectl get svc kong-kong-proxy -n student-cafe -o jsonpath='{.spec.ports[?(@.name=="kong-proxy-tls")].port}')

    echo "HTTP:  http://${LB_IP}:${HTTP_PORT}"
    echo "HTTPS: https://${LB_IP}:${HTTPS_PORT}"
fi
echo ""
echo "📋 Deployment Summary:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check final status of all components
echo "🔍 Component Status:"
CONSUL_STATUS=$(kubectl get pods -l app=consul -n student-cafe --no-headers 2>/dev/null | awk '{print $3}' | head -1)
KONG_STATUS=$(kubectl get pods -l app.kubernetes.io/name=kong -n student-cafe --no-headers 2>/dev/null | awk '{print $3}' | head -1)
INGRESS_STATUS=$(kubectl get ingress cafe-ingress -n student-cafe --no-headers 2>/dev/null | awk '{print "Created"}' || echo "Missing")

if [ "$CONSUL_STATUS" = "Running" ]; then
    echo "  ✅ Consul: $CONSUL_STATUS"
else
    echo "  ❌ Consul: $CONSUL_STATUS"
fi

if [ "$KONG_STATUS" = "Running" ]; then
    echo "  ✅ Kong: $KONG_STATUS"
else
    echo "  ❌ Kong: $KONG_STATUS"
fi

if [ "$INGRESS_STATUS" = "Created" ]; then
    echo "  ✅ Ingress: $INGRESS_STATUS"
else
    echo "  ❌ Ingress: $INGRESS_STATUS"
fi

# Application services status
APP_SERVICES=("food-catalog-service" "order-service" "cafe-ui")
echo ""
echo "🍕 Application Services:"
for service in "${APP_SERVICES[@]}"; do
    SERVICE_STATUS=$(kubectl get pods -l app=$service -n student-cafe --no-headers 2>/dev/null | awk '{print $3}' | head -1)
    if [ "$SERVICE_STATUS" = "Running" ]; then
        echo "  ✅ $service: $SERVICE_STATUS"
    elif [ -z "$SERVICE_STATUS" ]; then
        echo "  ⏳ $service: Pending"
    else
        echo "  ❌ $service: $SERVICE_STATUS"
    fi
done

echo ""
echo "📝 Useful Commands:"
echo "  View pods:        kubectl get pods -n student-cafe"
echo "  View services:    kubectl get services -n student-cafe"
echo "  View ingress:     kubectl get ingress -n student-cafe"
echo "  View logs:        kubectl logs -f <pod-name> -n student-cafe"
echo "  Kong status:      kubectl exec -n student-cafe deployment/kong-kong -- curl -s localhost:8444/status"
echo ""
echo "🧹 To cleanup:      ./cleanup.sh"
