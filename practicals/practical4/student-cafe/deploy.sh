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
helm upgrade --install consul hashicorp/consul \
    --set global.name=consul \
    --namespace student-cafe \
    --set server.replicas=1 \
    --set server.bootstrapExpect=1 \
    --wait --timeout=10m

echo "Waiting for Consul to be ready..."
kubectl wait --for=condition=ready pod -l app=consul -n student-cafe --timeout=300s

echo "Deploying Kong..."
helm upgrade --install kong kong/kong \
    --namespace student-cafe \
    --wait --timeout=10m

echo "Waiting for Kong to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=kong -n student-cafe --timeout=300s

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

# Wait a moment for deployments to start
sleep 10

# Configure Kong ingress
echo "Configuring Kong ingress..."
kubectl apply -f kong-ingress.yaml

# Wait for pods to be ready
echo "Waiting for application pods to be ready..."
kubectl wait --for=condition=ready pod -l app=food-catalog-service -n student-cafe --timeout=300s
kubectl wait --for=condition=ready pod -l app=order-service -n student-cafe --timeout=300s
kubectl wait --for=condition=ready pod -l app=cafe-ui -n student-cafe --timeout=300s

# Show pod status
echo "Pod Status:"
kubectl get pods -n student-cafe

# Get the Kong service URL
echo ""
echo "🎉 Deployment complete!"
echo ""
echo "Access your application at:"
KONG_URL=$(minikube service -n student-cafe kong-kong-proxy --url | head -1)
echo "$KONG_URL"
echo ""
echo "Useful commands:"
echo "  View pods:        kubectl get pods -n student-cafe"
echo "  View services:    kubectl get services -n student-cafe"
echo "  View ingress:     kubectl get ingress -n student-cafe"
echo "  View logs:        kubectl logs -f <pod-name> -n student-cafe"
echo ""
echo "To cleanup:         ./cleanup.sh"
