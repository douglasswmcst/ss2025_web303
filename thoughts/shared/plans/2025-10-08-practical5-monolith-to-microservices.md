# Practical 5: Refactoring a Monolithic Web Server to Microservices

## Overview

This practical teaches students how to systematically refactor a monolithic application into microservices. Students will start with a working monolithic Student Cafe application and progressively extract independent services, learning both the strategic thinking (identifying service boundaries) and tactical execution (refactoring code, managing data, orchestrating services).

This practical builds on Practical 2 (Consul + API Gateway) and Practical 4 (Kubernetes basics), preparing students for advanced topics like gRPC communication and production Kubernetes deployments.

## Learning Outcomes

- **Learning Outcome 1:** Identify the characteristics, benefits, and trade-offs of monolithic vs. microservices architectures
- **Learning Outcome 2:** Apply domain-driven design principles to identify service boundaries
- **Learning Outcome 3:** Systematically extract services from a monolith while maintaining functionality
- **Learning Outcome 4:** Implement service discovery patterns with Consul in a microservices ecosystem
- **Learning Outcome 5:** Deploy and orchestrate multiple services using Docker Compose
- **Learning Outcome 6:** Understand migration paths toward gRPC and Kubernetes

## Why This Practical Matters

Most real-world microservices adoptions don't start from scratch—they evolve from existing monolithic applications. Understanding how to decompose a monolith is a critical skill for:

- **Technical Decision Making:** Knowing when and how to split services
- **Risk Management:** Refactoring incrementally to avoid "big bang" failures
- **System Design:** Understanding service boundaries and communication patterns
- **Career Readiness:** Many companies are actively migrating monoliths to microservices

## Current State Analysis

### The Monolithic Student Cafe Application

We start with a **monolithic application** that handles all Student Cafe operations in a single codebase:

**Structure:**
```
student-cafe-monolith/
├── main.go                 # Single entry point
├── handlers/
│   ├── menu_handlers.go    # Menu/catalog endpoints
│   ├── order_handlers.go   # Order management endpoints
│   └── user_handlers.go    # User management endpoints
├── models/
│   ├── menu.go             # Menu item data structures
│   ├── order.go            # Order data structures
│   └── user.go             # User data structures
├── database/
│   └── db.go               # Single database connection
├── go.mod
└── Dockerfile
```

**Characteristics of the Monolith:**

1. **Single Deployment Unit:** All features deployed together
2. **Shared Database:** One PostgreSQL database with all tables (users, menu_items, orders)
3. **Tight Coupling:** All code shares memory, direct function calls
4. **Single Technology Stack:** All Go, one framework (Chi router)
5. **Simple to Run:** One command starts everything

**Pain Points (Why We Need to Refactor):**

1. **Scaling Inefficiency:** Can't scale menu browsing independently from order processing
2. **Deployment Risk:** Small change to menu requires redeploying entire app
3. **Team Bottlenecks:** Multiple teams can't work independently on different features
4. **Technology Lock-in:** Can't use different tech for different features (e.g., Python for ML recommendations)
5. **Failure Blast Radius:** Bug in user service crashes entire application
6. **Database Contention:** All features compete for same database connections

### Example Monolith Code

The monolith exposes these endpoints on a single server (`:8080`):

- `GET /api/menu` - List menu items
- `POST /api/menu` - Add menu item
- `GET /api/orders` - List orders
- `POST /api/orders` - Create order
- `GET /api/users/{id}` - Get user
- `POST /api/users` - Create user

All data lives in one database:
```sql
-- Single database schema
CREATE TABLE users (...);
CREATE TABLE menu_items (...);
CREATE TABLE orders (...);
CREATE TABLE order_items (...); -- Junction table
```

## Desired End State

### The Microservices Architecture

After refactoring, we'll have **three independent services**:

```
student-cafe-microservices/
├── api-gateway/              # Entry point (port 8080)
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── user-service/             # User management (port 8081)
│   ├── main.go
│   ├── models/
│   ├── handlers/
│   ├── database/
│   ├── go.mod
│   └── Dockerfile
├── menu-service/             # Menu/catalog (port 8082)
│   ├── main.go
│   ├── models/
│   ├── handlers/
│   ├── database/
│   ├── go.mod
│   └── Dockerfile
├── order-service/            # Order management (port 8083)
│   ├── main.go
│   ├── models/
│   ├── handlers/
│   ├── database/
│   ├── go.mod
│   └── Dockerfile
└── docker-compose.yml        # Orchestration
```

**Characteristics of the Microservices:**

1. **Independent Deployment:** Each service can be deployed separately
2. **Separate Databases:** Each service owns its data (database-per-service pattern)
3. **HTTP/REST Communication:** Services communicate via HTTP APIs
4. **Service Discovery:** Services find each other via Consul
5. **API Gateway:** Single entry point routes to appropriate services
6. **Independent Scaling:** Scale menu-service independently during peak browsing

**Benefits Achieved:**

1. **Independent Scaling:** Scale popular menu browsing without scaling entire system
2. **Isolated Failures:** Bug in user-service doesn't crash order-service
3. **Team Autonomy:** Different teams own different services
4. **Technology Freedom:** Could rebuild menu-service in Python if needed
5. **Faster Deployments:** Deploy order-service updates without touching menu-service
6. **Clear Ownership:** Each service has clear boundaries and responsibilities

## What We're NOT Doing

To keep this practical focused on core refactoring concepts, we explicitly exclude:

1. **gRPC Communication:** We'll use HTTP/REST first (gRPC comes in next section)
2. **Kubernetes Deployment:** We'll use Docker Compose first (K8s comes in next section)
3. **Resilience Patterns:** No circuit breakers, retries, timeouts yet
4. **Event-Driven Communication:** No message queues or event buses
5. **API Versioning:** No version management strategies
6. **Distributed Tracing:** No observability tools like Jaeger
7. **Authentication/Authorization:** No JWT, OAuth, or security layers
8. **Data Migration Strategies:** Simplified data separation approach

## How to Identify Service Boundaries

Before we start coding, we need to understand **why we split the way we do**.

### Domain-Driven Design Principles

We use simplified Domain-Driven Design (DDD) concepts to identify service boundaries:

#### 1. **Bounded Contexts**

A **bounded context** is a boundary within which a particular model is defined and applicable.

**In Student Cafe:**
- **User Context:** Everything about customers (profiles, authentication)
- **Menu Context:** Everything about food items (catalog, descriptions, pricing)
- **Order Context:** Everything about orders (cart, checkout, order history)

Each context should become a service.

#### 2. **Low Coupling, High Cohesion**

- **High Cohesion:** Things that change together should be together
  - Example: Menu items and their prices change together → same service
- **Low Coupling:** Services should depend on each other minimally
  - Example: Changing menu prices shouldn't require touching order service

#### 3. **Identify Entities and Aggregates**

**Entities** are objects with unique identities:
- User (identified by user_id)
- MenuItem (identified by menu_item_id)
- Order (identified by order_id)

**Aggregates** are clusters of entities treated as a unit:
- Order Aggregate includes Order + OrderItems (order line items)

**Rule:** Each aggregate should be owned by one service.

#### 4. **Business Capabilities**

Ask: "What business capabilities does the system provide?"
- **User Management:** Register users, get user profiles
- **Menu Management:** Maintain food catalog, update prices
- **Order Management:** Create orders, track status

Each capability maps to a service.

### Applying This to Student Cafe

| Capability | Entities | Bounded Context | Service |
|------------|----------|-----------------|---------|
| Manage customers | User | User Context | user-service |
| Manage food catalog | MenuItem | Menu Context | menu-service |
| Process orders | Order, OrderItem | Order Context | order-service |

**Why this split makes sense:**

1. **User Service:**
   - Changes when: User profile requirements change
   - Scales when: New user registrations spike
   - Independent because: User data is independent of menu

2. **Menu Service:**
   - Changes when: New menu items added, prices updated
   - Scales when: Many users browsing menu
   - Independent because: Menu can be read without orders

3. **Order Service:**
   - Changes when: Order workflow changes (e.g., add delivery)
   - Scales when: Lunch rush - many concurrent orders
   - Needs both: References users and menu items (but doesn't own them)

**Cross-Service Dependencies:**

The order-service needs to:
1. Validate the user exists (call user-service)
2. Validate menu items exist (call menu-service)
3. Store order with references to user_id and menu_item_ids

This is **inter-service communication** and is normal in microservices.

## Implementation Approach

We'll use a **Strangler Fig Pattern** - incrementally replace monolith parts with services while keeping the system running.

### Refactoring Strategy

1. **Start with Monolith:** Ensure we have a working baseline
2. **Extract Read-Only Service First:** Begin with menu-service (safest)
3. **Extract User Service:** Simple, few dependencies
4. **Extract Order Service:** Most complex, depends on others
5. **Add API Gateway:** Route requests to appropriate services
6. **Add Consul:** Enable dynamic service discovery
7. **Orchestrate with Docker Compose:** Manage all services together

## Phase 1: Build and Run the Monolith

### Overview
Establish a working monolith as our baseline. This ensures we understand the complete functionality before decomposing.

### Changes Required

#### 1. Create Monolith Project Structure

```bash
mkdir student-cafe-monolith
cd student-cafe-monolith
```

Create the following structure:
```
student-cafe-monolith/
├── main.go
├── models/
│   ├── user.go
│   ├── menu.go
│   └── order.go
├── handlers/
│   ├── user_handlers.go
│   ├── menu_handlers.go
│   └── order_handlers.go
├── database/
│   └── db.go
├── go.mod
├── Dockerfile
└── docker-compose.yml
```

#### 2. Database Models (`models/`)

**`models/user.go`:**
```go
package models

import "gorm.io/gorm"

type User struct {
    gorm.Model
    Name  string `json:"name"`
    Email string `json:"email" gorm:"unique"`
}
```

**`models/menu.go`:**
```go
package models

import "gorm.io/gorm"

type MenuItem struct {
    gorm.Model
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}
```

**`models/order.go`:**
```go
package models

import "gorm.io/gorm"

type Order struct {
    gorm.Model
    UserID     uint        `json:"user_id"`
    Status     string      `json:"status"` // "pending", "completed"
    OrderItems []OrderItem `json:"order_items" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
    gorm.Model
    OrderID    uint    `json:"order_id"`
    MenuItemID uint    `json:"menu_item_id"`
    Quantity   int     `json:"quantity"`
    Price      float64 `json:"price"` // Snapshot price at order time
}
```

**Why these models?**
- **User:** Represents a customer
- **MenuItem:** Represents a food item in the catalog
- **Order & OrderItem:** Order is an aggregate containing multiple order items
- Notice: OrderItem stores `Price` - this is a **snapshot** so historical orders aren't affected by menu price changes

#### 3. Database Connection (`database/db.go`)

```go
package database

import (
    "log"
    "student-cafe-monolith/models"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) error {
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return err
    }

    // Auto-migrate all tables
    err = DB.AutoMigrate(&models.User{}, &models.MenuItem{}, &models.Order{}, &models.OrderItem{})
    if err != nil {
        return err
    }

    log.Println("Database connected and migrated")
    return nil
}
```

**Why one connection?**
In the monolith, everything shares a single database connection. This is simple but creates contention.

#### 4. HTTP Handlers (`handlers/`)

**`handlers/user_handlers.go`:**
```go
package handlers

import (
    "encoding/json"
    "net/http"
    "student-cafe-monolith/database"
    "student-cafe-monolith/models"

    "github.com/go-chi/chi/v5"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := database.DB.Create(&user).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    var user models.User
    if err := database.DB.First(&user, id).Error; err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

**`handlers/menu_handlers.go`:**
```go
package handlers

import (
    "encoding/json"
    "net/http"
    "student-cafe-monolith/database"
    "student-cafe-monolith/models"
)

func GetMenu(w http.ResponseWriter, r *http.Request) {
    var items []models.MenuItem
    if err := database.DB.Find(&items).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(items)
}

func CreateMenuItem(w http.ResponseWriter, r *http.Request) {
    var item models.MenuItem
    if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := database.DB.Create(&item).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(item)
}
```

**`handlers/order_handlers.go`:**
```go
package handlers

import (
    "encoding/json"
    "net/http"
    "student-cafe-monolith/database"
    "student-cafe-monolith/models"
)

type CreateOrderRequest struct {
    UserID uint `json:"user_id"`
    Items  []struct {
        MenuItemID uint `json:"menu_item_id"`
        Quantity   int  `json:"quantity"`
    } `json:"items"`
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Validate user exists
    var user models.User
    if err := database.DB.First(&user, req.UserID).Error; err != nil {
        http.Error(w, "User not found", http.StatusBadRequest)
        return
    }

    // Create order
    order := models.Order{
        UserID: req.UserID,
        Status: "pending",
    }

    // Build order items
    for _, item := range req.Items {
        var menuItem models.MenuItem
        if err := database.DB.First(&menuItem, item.MenuItemID).Error; err != nil {
            http.Error(w, "Menu item not found", http.StatusBadRequest)
            return
        }

        orderItem := models.OrderItem{
            MenuItemID: item.MenuItemID,
            Quantity:   item.Quantity,
            Price:      menuItem.Price, // Snapshot current price
        }
        order.OrderItems = append(order.OrderItems, orderItem)
    }

    if err := database.DB.Create(&order).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(order)
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
    var orders []models.Order
    if err := database.DB.Preload("OrderItems").Find(&orders).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(orders)
}
```

**Why this structure?**
- All handlers share the same database connection
- CreateOrder demonstrates **tight coupling** - directly queries users and menu tables
- This works but creates dependencies we'll break apart

#### 5. Main Application (`main.go`)

```go
package main

import (
    "log"
    "net/http"
    "student-cafe-monolith/database"
    "student-cafe-monolith/handlers"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    // Connect to database
    dsn := "host=localhost user=postgres password=postgres dbname=student_cafe port=5432 sslmode=disable"
    if err := database.Connect(dsn); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Setup router
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // User routes
    r.Post("/api/users", handlers.CreateUser)
    r.Get("/api/users/{id}", handlers.GetUser)

    // Menu routes
    r.Get("/api/menu", handlers.GetMenu)
    r.Post("/api/menu", handlers.CreateMenuItem)

    // Order routes
    r.Post("/api/orders", handlers.CreateOrder)
    r.Get("/api/orders", handlers.GetOrders)

    log.Println("Monolith server starting on :8080")
    http.ListenAndServe(":8080", r)
}
```

#### 6. Docker Setup

**`Dockerfile`:**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /monolith .

FROM alpine:latest
WORKDIR /
COPY --from=builder /monolith /monolith
EXPOSE 8080
CMD ["/monolith"]
```

**`docker-compose.yml`:**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: student_cafe
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  monolith:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    environment:
      - DB_HOST=postgres
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=student_cafe

volumes:
  postgres_data:
```

**Note:** Update `main.go` to use environment variables for the DSN.

#### 7. Initialize Go Module

```bash
go mod init student-cafe-monolith
go get github.com/go-chi/chi/v5
go get gorm.io/gorm
go get gorm.io/driver/postgres
go mod tidy
```

### Success Criteria

#### Automated Verification:
- [x] Go modules download successfully: `go mod download`
- [x] Application compiles: `go build`
- [x] Docker image builds: `docker-compose build`
- [x] Services start successfully: `docker-compose up`
- [x] Database migrates tables: Check logs for "Database connected and migrated"

#### Manual Verification:
- [x] Create a menu item:
  ```bash
  curl -X POST http://localhost:8080/api/menu \
    -H "Content-Type: application/json" \
    -d '{"name": "Coffee", "description": "Hot coffee", "price": 2.50}'
  ```
- [x] Get menu items: `curl http://localhost:8080/api/menu`
- [x] Create a user:
  ```bash
  curl -X POST http://localhost:8080/api/users \
    -H "Content-Type: application/json" \
    -d '{"name": "John Doe", "email": "john@example.com"}'
  ```
- [x] Create an order:
  ```bash
  curl -X POST http://localhost:8080/api/orders \
    -H "Content-Type: application/json" \
    -d '{"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 2}]}'
  ```
- [x] Get all orders: `curl http://localhost:8080/api/orders`
- [x] Verify order contains correct user_id and menu_item_id references

**At this point:** You have a working monolith. All features work. Now we start decomposing.

---

## Phase 2: Extract Menu Service (Read-Only)

### Overview
Extract the menu/catalog functionality into an independent service. We start here because:
1. **Read-only:** Menu service mostly reads data (safer to extract)
2. **Few dependencies:** Doesn't depend on users or orders
3. **High traffic:** Menu browsing is high-volume, benefits from independent scaling

### Strategy

1. Create a new `menu-service` codebase
2. Copy menu-related models and handlers
3. Give it its own database (separate schema)
4. Run alongside monolith
5. Update monolith to proxy menu requests to the new service (gradual migration)

### Changes Required

#### 1. Create Menu Service Structure

```bash
mkdir menu-service
cd menu-service
```

Structure:
```
menu-service/
├── main.go
├── models/
│   └── menu.go
├── handlers/
│   └── menu_handlers.go
├── database/
│   └── db.go
├── go.mod
└── Dockerfile
```

#### 2. Menu Service Code

**`models/menu.go`:** (Same as monolith)
```go
package models

import "gorm.io/gorm"

type MenuItem struct {
    gorm.Model
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}
```

**`database/db.go`:**
```go
package database

import (
    "log"
    "menu-service/models"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) error {
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return err
    }

    // Only migrate menu-related tables
    err = DB.AutoMigrate(&models.MenuItem{})
    if err != nil {
        return err
    }

    log.Println("Menu database connected")
    return nil
}
```

**`handlers/menu_handlers.go`:** (Same logic as monolith)
```go
package handlers

import (
    "encoding/json"
    "net/http"
    "menu-service/database"
    "menu-service/models"
)

func GetMenu(w http.ResponseWriter, r *http.Request) {
    var items []models.MenuItem
    if err := database.DB.Find(&items).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(items)
}

func CreateMenuItem(w http.ResponseWriter, r *http.Request) {
    var item models.MenuItem
    if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := database.DB.Create(&item).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(item)
}
```

**`main.go`:**
```go
package main

import (
    "log"
    "net/http"
    "os"
    "menu-service/database"
    "menu-service/handlers"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    // Connect to dedicated menu database
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=localhost user=postgres password=postgres dbname=menu_db port=5432 sslmode=disable"
    }

    if err := database.Connect(dsn); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    r := chi.NewRouter()
    r.Use(middleware.Logger)

    // Menu endpoints
    r.Get("/menu", handlers.GetMenu)
    r.Post("/menu", handlers.CreateMenuItem)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8082"
    }

    log.Printf("Menu service starting on :%s", port)
    http.ListenAndServe(":"+port, r)
}
```

**Why separate database?**
- **Database-per-service pattern:** Each service owns its data
- **Independent scaling:** Menu database can be optimized for reads
- **Failure isolation:** Menu database issues don't affect order database
- **Migration flexibility:** Can change menu schema without coordinating with other teams

#### 3. Update Docker Compose

Add menu service and its database:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: student_cafe
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  menu-db:
    image: postgres:13
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: menu_db
    ports:
      - "5433:5432"  # Different host port
    volumes:
      - menu_data:/var/lib/postgresql/data

  monolith:
    build: ./student-cafe-monolith
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    environment:
      DATABASE_URL: "host=postgres user=postgres password=postgres dbname=student_cafe port=5432 sslmode=disable"

  menu-service:
    build: ./menu-service
    ports:
      - "8082:8082"
    depends_on:
      - menu-db
    environment:
      DATABASE_URL: "host=menu-db user=postgres password=postgres dbname=menu_db port=5432 sslmode=disable"
      PORT: "8082"

volumes:
  postgres_data:
  menu_data:
```

**Why two databases?**
We now have:
- `student_cafe` database (monolith) - contains users, orders, AND menu_items
- `menu_db` database (menu-service) - contains ONLY menu_items

Eventually we'll migrate menu data and remove menu_items from the monolith.

### Success Criteria

#### Automated Verification:
- [x] Menu service compiles: `cd menu-service && go build`
- [x] Docker Compose starts all services: `docker-compose up`
- [x] Menu service is reachable: `curl http://localhost:8082/menu`
- [x] Menu database is created: Check logs for "Menu database connected"

#### Manual Verification:
- [x] Add menu item directly to menu-service:
  ```bash
  curl -X POST http://localhost:8082/menu \
    -H "Content-Type: application/json" \
    -d '{"name": "Tea", "description": "Hot tea", "price": 1.50}'
  ```
- [x] Verify menu item appears: `curl http://localhost:8082/menu`
- [x] Verify monolith still works: `curl http://localhost:8080/api/menu`
- [x] Note: Monolith and menu-service have separate data at this point

**At this point:** Menu service runs independently. Next, we'll make the monolith route menu requests to it.

---

## Phase 3: Extract User Service

### Overview
Extract user management into a dedicated service. Similar process to menu service.

### Changes Required

#### 1. Create User Service Structure

```bash
mkdir user-service
cd user-service
```

Create similar structure to menu-service with user-specific models and handlers.

**`models/user.go`:**
```go
package models

import "gorm.io/gorm"

type User struct {
    gorm.Model
    Name  string `json:"name"`
    Email string `json:"email" gorm:"unique"`
}
```

**`handlers/user_handlers.go`:** (Copy from monolith, adjust imports)

**`main.go`:**
```go
package main

import (
    "log"
    "net/http"
    "os"
    "user-service/database"
    "user-service/handlers"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=localhost user=postgres password=postgres dbname=user_db port=5432 sslmode=disable"
    }

    if err := database.Connect(dsn); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    r := chi.NewRouter()
    r.Use(middleware.Logger)

    r.Post("/users", handlers.CreateUser)
    r.Get("/users/{id}", handlers.GetUser)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }

    log.Printf("User service starting on :%s", port)
    http.ListenAndServe(":"+port, r)
}
```

#### 2. Update Docker Compose

Add user-service and user-db:

```yaml
  user-db:
    image: postgres:13
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: user_db
    ports:
      - "5434:5432"
    volumes:
      - user_data:/var/lib/postgresql/data

  user-service:
    build: ./user-service
    ports:
      - "8081:8081"
    depends_on:
      - user-db
    environment:
      DATABASE_URL: "host=user-db user=postgres password=postgres dbname=user_db port=5432 sslmode=disable"
      PORT: "8081"
```

### Success Criteria

#### Automated Verification:
- [x] User service compiles and starts
- [x] User database initializes
- [x] Service is reachable: `curl http://localhost:8081/users/1`

#### Manual Verification:
- [x] Create user in user-service:
  ```bash
  curl -X POST http://localhost:8081/users \
    -H "Content-Type: application/json" \
    -d '{"name": "Alice", "email": "alice@example.com"}'
  ```
- [x] Get user: `curl http://localhost:8081/users/1`
- [x] Verify monolith still has its own users table

---

## Phase 4: Extract Order Service (Complex)

### Overview
Order service is the most complex because it has dependencies on both user-service and menu-service.

**Key Challenge:** Order service needs to validate that:
1. The user_id exists (must call user-service)
2. The menu_item_ids exist (must call menu-service)

This demonstrates **inter-service communication**.

### Changes Required

#### 1. Create Order Service with HTTP Clients

**`handlers/order_handlers.go`:**
```go
package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "order-service/database"
    "order-service/models"
)

type CreateOrderRequest struct {
    UserID uint `json:"user_id"`
    Items  []struct {
        MenuItemID uint `json:"menu_item_id"`
        Quantity   int  `json:"quantity"`
    } `json:"items"`
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Call user-service to validate user exists
    userServiceURL := "http://user-service:8081" // Direct service name
    userResp, err := http.Get(fmt.Sprintf("%s/users/%d", userServiceURL, req.UserID))
    if err != nil || userResp.StatusCode != http.StatusOK {
        http.Error(w, "User not found", http.StatusBadRequest)
        return
    }

    // Create order
    order := models.Order{
        UserID: req.UserID,
        Status: "pending",
    }

    // Validate each menu item by calling menu-service
    menuServiceURL := "http://menu-service:8082"
    for _, item := range req.Items {
        // Get menu item to snapshot price
        menuResp, err := http.Get(fmt.Sprintf("%s/menu/%d", menuServiceURL, item.MenuItemID))
        if err != nil || menuResp.StatusCode != http.StatusOK {
            http.Error(w, "Menu item not found", http.StatusBadRequest)
            return
        }

        var menuItem struct {
            Price float64 `json:"price"`
        }
        json.NewDecoder(menuResp.Body).Decode(&menuItem)

        orderItem := models.OrderItem{
            MenuItemID: item.MenuItemID,
            Quantity:   item.Quantity,
            Price:      menuItem.Price,
        }
        order.OrderItems = append(order.OrderItems, orderItem)
    }

    if err := database.DB.Create(&order).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(order)
}
```

**Why hardcoded URLs?**
For now, we hardcode `http://user-service:8081`. Docker Compose DNS resolves service names. Later, we'll replace this with Consul service discovery.

#### 2. Add Menu Item Retrieval to Menu Service

Menu service needs a `GET /menu/{id}` endpoint:

**`handlers/menu_handlers.go`:**
```go
func GetMenuItem(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    var item models.MenuItem
    if err := database.DB.First(&item, id).Error; err != nil {
        http.Error(w, "Menu item not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(item)
}
```

Add route in `menu-service/main.go`:
```go
r.Get("/menu/{id}", handlers.GetMenuItem)
```

### Success Criteria

#### Automated Verification:
- [x] Order service compiles and starts
- [x] Order service can reach user-service: Check logs during order creation
- [x] Order service can reach menu-service: Check logs during order creation

#### Manual Verification:
- [x] Create user in user-service
- [x] Create menu item in menu-service
- [x] Create order in order-service referencing both:
  ```bash
  curl -X POST http://localhost:8083/orders \
    -H "Content-Type: application/json" \
    -d '{"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 2}]}'
  ```
- [x] Verify order contains correct user_id and menu_item_id
- [x] Verify order fails if user_id doesn't exist
- [x] Verify order fails if menu_item_id doesn't exist

**At this point:** All three services run independently and communicate via HTTP.

---

## Phase 5: Add API Gateway

### Overview
Instead of exposing three different ports (8081, 8082, 8083), we create a single entry point that routes requests.

**Benefits:**
- Clients call one URL: `http://localhost:8080`
- Gateway routes `/api/users/*` → user-service
- Gateway routes `/api/menu/*` → menu-service
- Gateway routes `/api/orders/*` → order-service

### Changes Required

#### 1. Create API Gateway

```bash
mkdir api-gateway
cd api-gateway
```

**`main.go`:**
```go
package main

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)

    // Route /api/users/* to user-service
    r.HandleFunc("/api/users*", proxyTo("http://user-service:8081", "/users"))

    // Route /api/menu/* to menu-service
    r.HandleFunc("/api/menu*", proxyTo("http://menu-service:8082", "/menu"))

    // Route /api/orders/* to order-service
    r.HandleFunc("/api/orders*", proxyTo("http://order-service:8083", "/orders"))

    log.Println("API Gateway starting on :8080")
    http.ListenAndServe(":8080", r)
}

func proxyTo(targetURL, stripPrefix string) http.HandlerFunc {
    target, _ := url.Parse(targetURL)
    proxy := httputil.NewSingleHostReverseProxy(target)

    return func(w http.ResponseWriter, r *http.Request) {
        // Strip /api prefix
        r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
        log.Printf("Proxying %s to %s%s", r.Method, targetURL, r.URL.Path)
        proxy.ServeHTTP(w, r)
    }
}
```

**Why a gateway?**
- **Single entry point:** Clients don't need to know about individual services
- **Routing:** Gateway decides which service handles the request
- **Cross-cutting concerns:** Can add authentication, rate limiting here

#### 2. Update Docker Compose

```yaml
  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - user-service
      - menu-service
      - order-service
```

### Success Criteria

#### Automated Verification:
- [x] Gateway compiles and starts
- [x] Gateway can reach all backend services

#### Manual Verification:
- [x] All previous curl commands work but through gateway:
  ```bash
  # Users
  curl -X POST http://localhost:8080/api/users \
    -H "Content-Type: application/json" \
    -d '{"name": "Bob", "email": "bob@example.com"}'

  # Menu
  curl http://localhost:8080/api/menu

  # Orders
  curl -X POST http://localhost:8080/api/orders \
    -H "Content-Type: application/json" \
    -d '{"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 1}]}'
  ```
- [x] Verify gateway logs show routing decisions

**At this point:** API Gateway provides unified entry point to all microservices.

---

## Phase 6: Add Consul for Service Discovery

### Overview
Replace hardcoded service URLs with dynamic discovery via Consul.

**Why?**
- **Dynamic environments:** Services can move, scale, restart
- **No hardcoded IPs:** Services register themselves
- **Health checking:** Consul only routes to healthy instances

### Changes Required

#### 1. Add Consul to Docker Compose

```yaml
  consul:
    image: hashicorp/consul:latest
    ports:
      - "8500:8500"
    command: "agent -dev -client=0.0.0.0 -ui"
```

#### 2. Update Each Service to Register with Consul

Add to **user-service**, **menu-service**, **order-service**:

**Example for user-service `main.go`:**
```go
import (
    consulapi "github.com/hashicorp/consul/api"
)

func registerWithConsul(serviceName string, port int) error {
    config := consulapi.DefaultConfig()
    config.Address = "consul:8500"

    consul, err := consulapi.NewClient(config)
    if err != nil {
        return err
    }

    hostname, _ := os.Hostname()

    registration := &consulapi.AgentServiceRegistration{
        ID:      fmt.Sprintf("%s-%s", serviceName, hostname),
        Name:    serviceName,
        Port:    port,
        Address: hostname,
        Check: &consulapi.AgentServiceCheck{
            HTTP:     fmt.Sprintf("http://%s:%d/health", hostname, port),
            Interval: "10s",
            Timeout:  "3s",
        },
    }

    return consul.Agent().ServiceRegister(registration)
}

func main() {
    // ... existing code ...

    // Register with Consul
    if err := registerWithConsul("user-service", 8081); err != nil {
        log.Printf("Failed to register with Consul: %v", err)
    }

    // Add health endpoint
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    // ... start server ...
}
```

Repeat for menu-service (port 8082) and order-service (port 8083).

#### 3. Update API Gateway to Discover Services

**`api-gateway/main.go`:**
```go
import (
    consulapi "github.com/hashicorp/consul/api"
)

func discoverService(serviceName string) (string, error) {
    config := consulapi.DefaultConfig()
    config.Address = "consul:8500"

    consul, err := consulapi.NewClient(config)
    if err != nil {
        return "", err
    }

    services, _, err := consul.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return "", err
    }

    if len(services) == 0 {
        return "", fmt.Errorf("no healthy instances of %s", serviceName)
    }

    service := services[0].Service
    return fmt.Sprintf("http://%s:%d", service.Address, service.Port), nil
}

func proxyToService(serviceName, stripPrefix string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Discover service dynamically
        targetURL, err := discoverService(serviceName)
        if err != nil {
            http.Error(w, err.Error(), http.StatusServiceUnavailable)
            return
        }

        target, _ := url.Parse(targetURL)
        proxy := httputil.NewSingleHostReverseProxy(target)

        r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
        log.Printf("Proxying to %s at %s", serviceName, targetURL)
        proxy.ServeHTTP(w, r)
    }
}

func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)

    r.HandleFunc("/api/users*", proxyToService("user-service", "/users"))
    r.HandleFunc("/api/menu*", proxyToService("menu-service", "/menu"))
    r.HandleFunc("/api/orders*", proxyToService("order-service", "/orders"))

    log.Println("API Gateway starting on :8080")
    http.ListenAndServe(":8080", r)
}
```

#### 4. Update Order Service to Discover Dependencies

**`order-service/handlers/order_handlers.go`:**
```go
func discoverService(serviceName string) (string, error) {
    // ... same discoverService function as gateway ...
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
    // ... decode request ...

    // Discover user-service
    userServiceURL, err := discoverService("user-service")
    if err != nil {
        http.Error(w, "User service unavailable", http.StatusServiceUnavailable)
        return
    }

    // Discover menu-service
    menuServiceURL, err := discoverService("menu-service")
    if err != nil {
        http.Error(w, "Menu service unavailable", http.StatusServiceUnavailable)
        return
    }

    // ... rest of handler using discovered URLs ...
}
```

### Success Criteria

#### Automated Verification:
- [ ] Consul starts: `docker-compose up consul`
- [ ] All services register: Check http://localhost:8500/ui
- [ ] All services show healthy: Green status in Consul UI

#### Manual Verification:
- [ ] Open Consul UI: http://localhost:8500
- [ ] Verify user-service, menu-service, order-service appear
- [ ] Test gateway routes still work (now via Consul discovery)
- [ ] Stop menu-service: `docker-compose stop menu-service`
- [ ] Verify Consul marks it unhealthy
- [ ] Verify gateway returns 503 for menu requests
- [ ] Restart menu-service: `docker-compose start menu-service`
- [ ] Verify it re-registers and becomes healthy

**At this point:** Full microservices architecture with service discovery!

---

## Testing Strategy

### Unit Tests

For each service, create unit tests for handlers:

**Example for user-service:**
```go
// user-service/handlers/user_handlers_test.go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "user-service/handlers"
)

func TestCreateUser(t *testing.T) {
    // Setup test database
    // ...

    reqBody := []byte(`{"name": "Test User", "email": "test@example.com"}`)
    req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    handlers.CreateUser(w, req)

    if w.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got %d", w.Code)
    }
}
```

### Integration Tests

Test inter-service communication:

```bash
# Create user
USER_ID=$(curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Integration Test", "email": "test@example.com"}' | jq -r '.id')

# Create menu item
ITEM_ID=$(curl -X POST http://localhost:8080/api/menu \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Item", "price": 5.00}' | jq -r '.id')

# Create order linking both
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": $USER_ID, \"items\": [{\"menu_item_id\": $ITEM_ID, \"quantity\": 1}]}"
```

### Manual Testing Scenarios

1. **Normal Flow:** Create user → Create menu items → Create order
2. **Failure Scenarios:**
   - Try creating order with non-existent user_id (should fail)
   - Try creating order with non-existent menu_item_id (should fail)
3. **Service Resilience:**
   - Stop user-service, try creating order (should fail gracefully)
   - Restart user-service, try again (should work)

---

## Performance Considerations

### Monolith vs Microservices Performance

**Monolith Advantages:**
- Single process: No network latency for internal calls
- Single database: No distributed query overhead
- Simpler caching: Shared in-memory cache

**Microservices Trade-offs:**
- Network latency: Order service makes 2+ HTTP calls (user, menu)
- Distributed queries: Can't do SQL joins across services
- More infrastructure: 3 services + gateway + consul vs 1 service

**Optimizations:**

1. **Caching:** Cache user/menu data in order-service
2. **Async Communication:** Use events for non-critical updates
3. **Batch APIs:** Add bulk endpoints (e.g., GET /menu?ids=1,2,3)
4. **Connection Pooling:** Reuse HTTP connections between services

---

## Migration Notes

### Data Migration Strategy

When moving from monolith to microservices:

1. **Dual-Write Phase:**
   - Monolith writes to both old and new databases
   - Ensures data consistency during transition

2. **Read-Switch Phase:**
   - Start reading from new service
   - Verify data correctness

3. **Cleanup Phase:**
   - Remove old tables from monolith database
   - Decommission old code paths

**For this practical:**
We simplified by creating separate databases from the start. In production, you'd migrate existing data:

```sql
-- Example: Copy menu_items from monolith to menu-service
INSERT INTO menu_db.menu_items
SELECT * FROM student_cafe.menu_items;
```

### Handling Foreign Keys

**Problem:** Orders reference user_id and menu_item_id, but those tables are now in different databases.

**Solution:** Store IDs only, validate via API calls
- Order service stores `user_id` as integer (not foreign key)
- Validates existence by calling user-service HTTP API
- Same for menu_item_id

**Trade-off:** No database-level referential integrity. Must enforce in application code.

---

## Next Steps: Path to gRPC and Kubernetes

### Part 1: Migrating to gRPC

**Why gRPC?**
- Faster than HTTP/JSON (binary protocol)
- Strongly-typed contracts (protobuf)
- Streaming support

**Migration Steps:**

1. **Define Protobuf Contracts:**
   ```protobuf
   // user.proto
   service UserService {
       rpc GetUser(GetUserRequest) returns (UserResponse);
       rpc CreateUser(CreateUserRequest) returns (UserResponse);
   }
   ```

2. **Replace HTTP Handlers with gRPC Servers:**
   - user-service implements UserService
   - menu-service implements MenuService
   - order-service implements OrderService

3. **Update Order Service to Use gRPC Clients:**
   ```go
   // Instead of HTTP call
   userResp, err := http.Get(fmt.Sprintf("%s/users/%d", url, userID))

   // Use gRPC
   userClient := pb.NewUserServiceClient(conn)
   userResp, err := userClient.GetUser(ctx, &pb.GetUserRequest{Id: userID})
   ```

4. **Keep API Gateway as HTTP → gRPC Bridge:**
   - Gateway still accepts HTTP from clients
   - Converts to gRPC calls internally

**Resources:**
- Refer back to Practical 1 (gRPC basics)
- Refer to Practical 3 (gRPC with databases)

### Part 2: Migrating to Kubernetes

**Why Kubernetes?**
- Production-grade orchestration
- Auto-scaling, self-healing
- Load balancing, service mesh

**Migration Steps:**

1. **Create Kubernetes Manifests:**
   ```yaml
   # user-service-deployment.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: user-service
   spec:
     replicas: 2  # Scale to 2 instances
     selector:
       matchLabels:
         app: user-service
     template:
       metadata:
         labels:
           app: user-service
       spec:
         containers:
         - name: user-service
           image: user-service:latest
           ports:
           - containerPort: 8081
   ```

2. **Deploy to Minikube:**
   ```bash
   minikube start
   kubectl apply -f user-service-deployment.yaml
   kubectl apply -f user-service.yaml
   ```

3. **Replace Consul with Kubernetes DNS:**
   - Services discover each other via K8s service names
   - `http://user-service.default.svc.cluster.local:8081`

4. **Add Kong Gateway:**
   - Replace simple API gateway with Kong for advanced routing

**Resources:**
- Refer to Practical 4 (Kubernetes with Kong)

---

## Common Pitfalls and Troubleshooting

### 1. Circular Dependencies
**Problem:** Service A calls Service B which calls Service A
**Solution:** Redesign service boundaries. Use events for notifications.

### 2. Data Inconsistency
**Problem:** User deleted from user-service, but orders still reference user_id
**Solution:** Implement soft deletes, or use sagas/distributed transactions

### 3. Network Failures
**Problem:** Order service can't reach user-service (network issue)
**Solution:** Add retries, timeouts, circuit breakers (covered in future practical)

### 4. Database Schema Conflicts
**Problem:** Multiple services need same data
**Solution:** Duplicate data (eventual consistency) or rethink service boundaries

### 5. Service Discovery Failures
**Problem:** Consul down, services can't find each other
**Solution:** Local DNS fallbacks, health checks, monitoring

---

## Learning Reflection Questions

After completing this practical, students should be able to answer:

1. **Strategic:**
   - Why did we split user, menu, and order into separate services?
   - What are the trade-offs of the database-per-service pattern?
   - When would you NOT split a monolith?

2. **Tactical:**
   - How does order-service validate user_id exists without direct database access?
   - What happens if menu-service is down when creating an order?
   - How does the API gateway route requests to the correct service?

3. **Operational:**
   - How do you scale menu-service independently?
   - What happens when a service crashes and restarts?
   - How does Consul know a service is healthy?

---

## References

- **Original Monolith Code:** `student-cafe-monolith/` directory
- **Final Microservices Code:** `student-cafe-microservices/` directory
- **Related Practicals:**
  - Practical 1: gRPC basics (next step: add gRPC to these services)
  - Practical 2: Consul + API Gateway (foundation for this practical)
  - Practical 4: Kubernetes deployment (next step: deploy these to K8s)
- **Further Reading:**
  - "Building Microservices" by Sam Newman
  - "Domain-Driven Design" by Eric Evans
  - Martin Fowler's "Microservices" article

---

## Submission Requirements

### What to Submit

1. **Complete Microservices Project:**
   - `user-service/` directory with all code
   - `menu-service/` directory with all code
   - `order-service/` directory with all code
   - `api-gateway/` directory with all code
   - `docker-compose.yml` orchestrating all services
   - `README.md` documenting your approach

2. **Documentation (in README.md):**
   - **Architecture Diagram:** Draw the final microservices architecture
   - **Service Boundaries Justification:** Explain why you split services this way
   - **Challenges Encountered:** Document problems and solutions
   - **Screenshots:**
     - Consul UI showing all services healthy
     - Successful order creation (showing inter-service communication)
     - Logs from order-service showing calls to user/menu services

3. **Reflection Essay (500 words):**
   - Compare monolith vs microservices for this use case
   - When would you choose one over the other?
   - What did you learn about service boundaries?

### Grading Criteria

| Criteria | Weight |
|----------|--------|
| All services run independently | 20% |
| Inter-service communication works | 25% |
| Consul service discovery implemented | 20% |
| API Gateway routes correctly | 15% |
| Documentation and reflection | 20% |

---

## Conclusion

This practical taught you how to systematically refactor a monolithic application into microservices. You learned:

1. **Strategic thinking:** Identifying service boundaries using domain-driven design
2. **Tactical execution:** Extracting services incrementally while maintaining functionality
3. **Operational concerns:** Service discovery, API gateways, orchestration

You now have a foundation to:
- Migrate gRPC for efficient inter-service communication (next practical)
- Deploy to Kubernetes for production-grade orchestration (future topic)
- Add resilience patterns like circuit breakers and retries (future topic)

**Key Takeaway:** Microservices aren't just about splitting code—they're about organizational scaling, independent deployment, and managing complexity through clear boundaries.
