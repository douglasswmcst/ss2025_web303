# Student Cafe Microservices Application

## Architecture Overview

This project implements a microservices-based cafe ordering system using Go, React, Kong API Gateway, Consul for service discovery, and Kubernetes for orchestration. The system consists of:

- **Frontend (React.js):** A single-page application providing the user interface for students to browse menu items and place orders.
- **food-catalog-service (Go & Chi):** A microservice responsible for providing a list of available food items.
- **order-service (Go & Chi):** A microservice for creating and managing food orders.
- **Service Discovery (Consul):** Allows microservices to find and communicate with each other.
- **API Gateway (Kong):** Single entry point for all external traffic with intelligent routing.
- **Containerization & Orchestration (Docker & Kubernetes):** All components are containerized and deployed on a local Kubernetes cluster.

## User Flow

1. Student's browser loads the React application
2. React app makes API calls to Kong API Gateway
3. Kong routes traffic to appropriate microservices:
   - `/api/catalog` → `food-catalog-service`
   - `/api/orders` → `order-service`
4. The `order-service` communicates with `food-catalog-service` via Consul service discovery
5. All components run as pods within a Kubernetes cluster

## Prerequisites

Ensure you have the following tools installed:

- **Go** (version 1.18+)
- **Node.js & npm** (for the React frontend)
- **Docker** (for containerizing apps)
- **Minikube** (for local Kubernetes cluster)
- **kubectl** (Kubernetes command-line tool)
- **Helm** (Kubernetes package manager)

## Setup Instructions

### 1. Start Minikube and Configure Docker

```bash
# Start your local Kubernetes cluster
minikube start --cpus 4 --memory 4096

# Point your local docker client to minikube's docker daemon
eval $(minikube -p minikube docker-env)
```

### 2. Create Kubernetes Namespace

```bash
kubectl create namespace student-cafe
```

### 3. Deploy Infrastructure Services

**Deploy Consul:**

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install consul hashicorp/consul --set global.name=consul --namespace student-cafe --set server.replicas=1 --set server.bootstrapExpect=1
```

**Deploy Kong:**

```bash
helm repo add kong https://charts.konghq.com
helm repo update
helm install kong kong/kong --namespace student-cafe
```

### 4. Build Docker Images

From the project root directory:

```bash
docker build -t food-catalog-service:v1 ./food-catalog-service
docker build -t order-service:v1 ./order-service
docker build -t cafe-ui:v1 ./cafe-ui
```

### 5. Deploy Application Services

```bash
kubectl apply -f app-deployment.yaml
```

### 6. Configure Kong API Gateway

```bash
kubectl apply -f kong-ingress.yaml
```

### 7. Access the Application

Get the external IP address for Kong:

```bash
minikube service -n student-cafe kong-kong-proxy --url
```

Open the returned URL in your web browser to access the Student Cafe application.

## API Endpoints

- **GET /api/catalog/items** - Retrieve list of available food items
- **POST /api/orders/orders** - Create a new food order

## Project Structure

```
student-cafe/
├── food-catalog-service/
│   ├── main.go
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── order-service/
│   ├── main.go
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── cafe-ui/
│   ├── src/
│   ├── public/
│   ├── Dockerfile
│   ├── package.json
│   └── package-lock.json
├── app-deployment.yaml
├── kong-ingress.yaml
└── README.md
```

## Development Commands

### View Running Pods

```bash
kubectl get pods -n student-cafe
```

### View Services

```bash
kubectl get services -n student-cafe
```

### View Logs

```bash
kubectl logs -f <pod-name> -n student-cafe
```

### Check Ingress Status

```bash
kubectl get ingress -n student-cafe
```

## Troubleshooting

- If pods are not starting, check image pull policy is set to `IfNotPresent`
- Ensure minikube docker environment is configured with `eval $(minikube -p minikube docker-env)`
- Verify all services are healthy with `kubectl get pods -n student-cafe`
- Check service discovery is working by examining pod logs

## Challenges and Solutions

- **Service Discovery**: Implemented Consul for dynamic service discovery instead of hardcoding IP addresses
- **API Gateway Routing**: Used Kong ingress with path-based routing to route traffic to appropriate microservices
- **Container Orchestration**: Leveraged Kubernetes deployments and services for scalable container management
- **Development Workflow**: Used minikube's docker environment to build images locally without pushing to remote registry

## Future Enhancements

- Implement resilience patterns (timeout, retry, circuit breaker)
- Add monitoring and logging with Prometheus and Grafana
- Implement database persistence
- Add authentication and authorization
- Create CI/CD pipeline for automated deployments
