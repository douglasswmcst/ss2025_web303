#!/bin/bash

# Student Cafe Application Cleanup Script

echo "🧹 Cleaning up Student Cafe Application..."

# Delete application resources
echo "Deleting application resources..."
kubectl delete -f kong-ingress.yaml --ignore-not-found
kubectl delete -f app-deployment.yaml --ignore-not-found

# Delete Helm releases
echo "Deleting Helm releases..."
helm uninstall kong -n student-cafe --ignore-not-found
helm uninstall consul -n student-cafe --ignore-not-found

# Delete namespace (this will also delete any remaining resources)
echo "Deleting student-cafe namespace..."
kubectl delete namespace student-cafe --ignore-not-found

# Remove Docker images
echo "Removing Docker images..."
docker rmi food-catalog-service:v1 --force 2>/dev/null || true
docker rmi order-service:v1 --force 2>/dev/null || true
docker rmi cafe-ui:v1 --force 2>/dev/null || true

echo "✅ Cleanup complete!"
echo ""
echo "To stop minikube completely:"
echo "minikube stop"
