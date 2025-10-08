# **Module Practical: WEB303 Microservices & Serverless Applications**

## **Practical 5: Refactoring a Monolithic Web Server to Microservices**

### **Objective**

This practical teaches you how to systematically refactor a monolithic application into microservices. You'll start with a working monolithic Student Cafe application and progressively extract independent services, learning both the strategic thinking (identifying service boundaries) and tactical execution (refactoring code, managing data, orchestrating services).

This practical builds on Practical 2 (Consul + API Gateway) and prepares you for advanced topics like gRPC communication and production Kubernetes deployments.

### **Learning Outcomes**

By completing this practical, you will be able to:

- **Learning Outcome 1:** Identify the characteristics, benefits, and trade-offs of monolithic vs. microservices architectures
- **Learning Outcome 2:** Apply domain-driven design principles to identify service boundaries
- **Learning Outcome 3:** Systematically extract services from a monolith while maintaining functionality
- **Learning Outcome 4:** Implement service discovery patterns with Consul in a microservices ecosystem
- **Learning Outcome 5:** Deploy and orchestrate multiple services using Docker Compose
- **Learning Outcome 6:** Understand migration paths toward gRPC and Kubernetes

---

## **Why This Practical Matters**

Most real-world microservices adoptions don't start from scratch—they evolve from existing monolithic applications. Understanding how to decompose a monolith is a critical skill for:

- **Technical Decision Making:** Knowing when and how to split services
- **Risk Management:** Refactoring incrementally to avoid "big bang" failures
- **System Design:** Understanding service boundaries and communication patterns
- **Career Readiness:** Many companies are actively migrating monoliths to microservices

---

## **Part 1: Understanding Service Boundaries**

Before we start coding, we need to understand **why we split the way we do**.

### **Domain-Driven Design Principles**

We use simplified Domain-Driven Design (DDD) concepts to identify service boundaries:

#### **1. Bounded Contexts**

A **bounded context** is a boundary within which a particular model is defined and applicable.

**In Student Cafe:**
- **User Context:** Everything about customers (profiles, authentication)
- **Menu Context:** Everything about food items (catalog, descriptions, pricing)
- **Order Context:** Everything about orders (cart, checkout, order history)

Each context should become a service.

#### **2. Low Coupling, High Cohesion**

- **High Cohesion:** Things that change together should be together
  - Example: Menu items and their prices change together → same service
- **Low Coupling:** Services should depend on each other minimally
  - Example: Changing menu prices shouldn't require touching order service

#### **3. Identify Entities and Aggregates**

**Entities** are objects with unique identities:
- User (identified by user_id)
- MenuItem (identified by menu_item_id)
- Order (identified by order_id)

**Aggregates** are clusters of entities treated as a unit:
- Order Aggregate includes Order + OrderItems (order line items)

**Rule:** Each aggregate should be owned by one service.

#### **4. Business Capabilities**

Ask: "What business capabilities does the system provide?"
- **User Management:** Register users, get user profiles
- **Menu Management:** Maintain food catalog, update prices
- **Order Management:** Create orders, track status

Each capability maps to a service.

### **Applying This to Student Cafe**

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

---

## **Part 2: Build the Monolith (Baseline)**

### **Overview**

We'll start by building a working monolithic application. This gives us:
- A baseline to compare against
- Understanding of the complete functionality
- A working system to refactor incrementally

### **Project Setup**

Create a directory for the monolith:

```bash
mkdir -p practicals/practical5/student-cafe-monolith
cd practicals/practical5/student-cafe-monolith
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

### **Step 1: Initialize Go Module**

```bash
go mod init student-cafe-monolith
go get github.com/go-chi/chi/v5
go get gorm.io/gorm
go get gorm.io/driver/postgres
go mod tidy
```

### **Step 2: Create Database Models**

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

**Why OrderItem stores Price:**
This is a **snapshot** - historical orders aren't affected by future menu price changes.

### **Step 3: Create Database Connection**

**`database/db.go`:**
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

### **Step 4: Create HTTP Handlers**

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

**Notice the tight coupling:**
CreateOrder directly queries users and menu tables in the same database. This works but creates dependencies we'll break apart later.

### **Step 5: Create Main Application**

**`main.go`:**
```go
package main

import (
    "log"
    "net/http"
    "os"
    "student-cafe-monolith/database"
    "student-cafe-monolith/handlers"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    // Connect to database
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=localhost user=postgres password=postgres dbname=student_cafe port=5432 sslmode=disable"
    }

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

### **Step 6: Create Docker Files**

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
      DATABASE_URL: "host=postgres user=postgres password=postgres dbname=student_cafe port=5432 sslmode=disable"

volumes:
  postgres_data:
```

### **Step 7: Test the Monolith**

Start the application:
```bash
docker-compose up --build
```

Test the endpoints:

```bash
# Create a menu item
curl -X POST http://localhost:8080/api/menu \
  -H "Content-Type: application/json" \
  -d '{"name": "Coffee", "description": "Hot coffee", "price": 2.50}'

# Get menu items
curl http://localhost:8080/api/menu

# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Create an order
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 2}]}'

# Get all orders
curl http://localhost:8080/api/orders
```

**Checkpoint:** You now have a working monolith with all features. Everything is in one codebase, one database, one deployment unit.

---

## **Part 3: Extract Menu Service**

### **Overview**

We'll extract menu functionality into an independent service. We start here because:
1. **Read-mostly:** Menu service mostly reads data (safer to extract)
2. **Few dependencies:** Doesn't depend on users or orders
3. **High traffic:** Menu browsing is high-volume, benefits from independent scaling

### **Step 1: Create Menu Service Structure**

Create a new directory:
```bash
mkdir -p practicals/practical5/menu-service
cd practicals/practical5/menu-service
```

Initialize the Go module:
```bash
go mod init menu-service
go get github.com/go-chi/chi/v5
go get gorm.io/gorm
go get gorm.io/driver/postgres
go mod tidy
```

Create the structure:
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

### **Step 2: Implement Menu Service**

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

**`handlers/menu_handlers.go`:**
```go
package handlers

import (
    "encoding/json"
    "net/http"
    "menu-service/database"
    "menu-service/models"

    "github.com/go-chi/chi/v5"
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

    // Menu endpoints (note: no /api prefix)
    r.Get("/menu", handlers.GetMenu)
    r.Post("/menu", handlers.CreateMenuItem)
    r.Get("/menu/{id}", handlers.GetMenuItem)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8082"
    }

    log.Printf("Menu service starting on :%s", port)
    http.ListenAndServe(":"+port, r)
}
```

**`Dockerfile`:**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /menu-service .

FROM alpine:latest
WORKDIR /
COPY --from=builder /menu-service /menu-service
EXPOSE 8082
CMD ["/menu-service"]
```

### **Step 3: Update Root Docker Compose**

Update your root `docker-compose.yml` to include the menu service and its database:

```yaml
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

### **Step 4: Test Menu Service**

```bash
# Start all services
docker-compose up --build

# Test menu service directly
curl -X POST http://localhost:8082/menu \
  -H "Content-Type: application/json" \
  -d '{"name": "Tea", "description": "Hot tea", "price": 1.50}'

# Get menu
curl http://localhost:8082/menu
```

**Key Learning:**
- Menu service has its own database (`menu_db`)
- Monolith still has its own menu_items table
- They're currently separate - no data sharing yet
- This demonstrates the **database-per-service pattern**

---

## **Part 4: Extract User Service**

Following the same pattern, extract the user service.

### **Create User Service**

```bash
mkdir -p practicals/practical5/user-service
cd practicals/practical5/user-service
go mod init user-service
go get github.com/go-chi/chi/v5
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

**Implementation** (following the same structure as menu-service):

- `models/user.go` - User model
- `handlers/user_handlers.go` - CreateUser, GetUser handlers
- `database/db.go` - Database connection
- `main.go` - HTTP server on port 8081
- `Dockerfile`

Add to docker-compose:
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

---

## **Part 5: Extract Order Service (Inter-Service Communication)**

### **Overview**

Order service is the most complex because it has dependencies on both user-service and menu-service.

**Key Challenge:** Order service must:
1. Validate the user exists (call user-service via HTTP)
2. Validate menu items exist (call menu-service via HTTP)

This demonstrates **inter-service communication**.

### **Create Order Service**

```bash
mkdir -p practicals/practical5/order-service
cd practicals/practical5/order-service
go mod init order-service
go get github.com/go-chi/chi/v5
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

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
    userServiceURL := "http://user-service:8081"
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

**Notice:**
- Order service makes HTTP calls to `http://user-service:8081` and `http://menu-service:8082`
- Docker Compose DNS resolves these service names automatically
- No shared database - services communicate via APIs

---

## **Part 6: Add API Gateway**

### **Overview**

Instead of exposing three different ports (8081, 8082, 8083), create a single entry point.

**Benefits:**
- Clients call one URL: `http://localhost:8080`
- Gateway routes `/api/users/*` → user-service
- Gateway routes `/api/menu/*` → menu-service
- Gateway routes `/api/orders/*` → order-service

### **Create API Gateway**

```bash
mkdir -p practicals/practical5/api-gateway
cd practicals/practical5/api-gateway
go mod init api-gateway
go get github.com/go-chi/chi/v5
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

**Add to docker-compose:**
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

### **Test Through Gateway**

Now all requests go through port 8080:

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

---

## **Part 7: Add Consul Service Discovery**

### **Overview**

Replace hardcoded service URLs with dynamic discovery via Consul.

### **Step 1: Add Consul to Docker Compose**

```yaml
  consul:
    image: hashicorp/consul:latest
    ports:
      - "8500:8500"
    command: "agent -dev -client=0.0.0.0 -ui"
```

### **Step 2: Register Services with Consul**

Update each service (user, menu, order) to register with Consul.

**Example for user-service `main.go`:**

Add the import:
```go
import (
    "fmt"
    consulapi "github.com/hashicorp/consul/api"
)
```

Add the registration function:
```go
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
```

Update main function:
```go
func main() {
    // ... existing database connection code ...

    r := chi.NewRouter()
    r.Use(middleware.Logger)

    // Add health endpoint
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    r.Post("/users", handlers.CreateUser)
    r.Get("/users/{id}", handlers.GetUser)

    // Register with Consul
    if err := registerWithConsul("user-service", 8081); err != nil {
        log.Printf("Failed to register with Consul: %v", err)
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }

    log.Printf("User service starting on :%s", port)
    http.ListenAndServe(":"+port, r)
}
```

Don't forget to get the Consul dependency:
```bash
go get github.com/hashicorp/consul/api
```

Repeat for menu-service (port 8082) and order-service (port 8083).

### **Step 3: Update API Gateway to Discover Services**

**`api-gateway/main.go`:**
```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
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

### **Step 4: Update Order Service to Discover Dependencies**

Update order service to discover user-service and menu-service dynamically (similar to API Gateway).

### **Step 5: Test Consul Integration**

```bash
docker-compose up --build

# Open Consul UI
open http://localhost:8500

# Verify all services appear as healthy
# Test that requests still work through gateway
curl http://localhost:8080/api/menu
```

**Try stopping a service:**
```bash
docker-compose stop menu-service

# Gateway should return 503 for menu requests
curl http://localhost:8080/api/menu

# Restart
docker-compose start menu-service

# Service re-registers and becomes healthy again
```

---

## **Submission Requirements**

### **What to Submit**

1. **Complete Microservices Project:**
   - `user-service/` directory with all code
   - `menu-service/` directory with all code
   - `order-service/` directory with all code
   - `api-gateway/` directory with all code
   - Root `docker-compose.yml` orchestrating all services
   - `README.md` documenting your implementation

2. **Documentation (in README.md):**
   - **Architecture Diagram:** Draw the final microservices architecture
   - **Service Boundaries Justification:** Explain why you split services this way
   - **Challenges Encountered:** Document problems and solutions
   - **Screenshots:**
     - Consul UI showing all services healthy
     - Successful order creation (terminal output showing inter-service communication)
     - Logs from order-service showing calls to user/menu services

3. **Reflection Essay (500 words minimum):**
   Answer these questions:
   - Compare monolith vs microservices for this use case
   - What are the trade-offs of the database-per-service pattern?
   - When would you choose NOT to split a monolith?
   - How does order-service validate user_id exists without direct database access?
   - What happens if menu-service is down when creating an order?
   - How would you add caching to improve performance?

### **Grading Criteria**

| Criteria | Weight |
|----------|--------|
| All services run independently | 20% |
| Inter-service communication works correctly | 25% |
| Consul service discovery implemented | 20% |
| API Gateway routes correctly | 15% |
| Documentation and reflection | 20% |

---

## **Troubleshooting Common Issues**

### Issue 1: Service Can't Connect to Database
**Symptom:** Service crashes with "connection refused"
**Solution:** Ensure database service starts first. Use `depends_on` in docker-compose.

### Issue 2: Inter-Service Communication Fails
**Symptom:** Order service can't reach user-service
**Solution:** Verify service names match in docker-compose. Use `docker-compose logs` to check.

### Issue 3: Consul Registration Fails
**Symptom:** Services don't appear in Consul UI
**Solution:** Check Consul is accessible at `consul:8500`. Verify health endpoint works.

### Issue 4: Port Conflicts
**Symptom:** "Address already in use"
**Solution:** Stop other services using ports 8080-8083, 5432-5434, 8500.

### Issue 5: Go Module Checksum Mismatch
**Symptom:** Docker build fails with "SECURITY ERROR - checksum mismatch" for `github.com/go-chi/chi/v5`
**Solution:** The `go.sum` file has incorrect checksums. Regenerate it for each service:

```bash
# For monolith
cd student-cafe-monolith
rm go.sum
go mod tidy

# For each microservice
cd ../user-service && rm go.sum && go mod tidy
cd ../menu-service && rm go.sum && go mod tidy
cd ../order-service && rm go.sum && go mod tidy
cd ../api-gateway && rm go.sum && go mod tidy
```

**Why this happens:** The `go.sum` file may have been manually created or copied with incorrect checksums. Always run `go mod tidy` after creating `go.mod`.

---

## **Conclusion**

Congratulations! You've successfully refactored a monolithic application into microservices. You've learned:

1. **Strategic thinking:** Identifying service boundaries using domain-driven design
2. **Tactical execution:** Extracting services incrementally while maintaining functionality
3. **Operational concerns:** Service discovery, API gateways, orchestration

### **Next Steps**

- **Migrate to gRPC:** Replace HTTP/REST with gRPC for efficient inter-service communication (refer to Practical 1)
- **Deploy to Kubernetes:** Move from Docker Compose to Kubernetes for production-grade orchestration (refer to Practical 4)
- **Add Resilience Patterns:** Implement circuit breakers, retries, and timeouts (future practical)

**Key Takeaway:** Microservices aren't just about splitting code—they're about organizational scaling, independent deployment, and managing complexity through clear boundaries.
