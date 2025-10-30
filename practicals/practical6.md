# Practical 6: Comprehensive Testing for Microservices

## Objective

Building on Practical 5A, this practical teaches you how to:

1. Implement **unit tests** for individual gRPC service methods
2. Create **integration tests** that test multiple services working together
3. Develop **end-to-end (E2E) tests** that validate the entire system through the API
4. Use testing best practices including mocks, test isolation, and coverage reporting
5. Automate testing with Make commands for CI/CD pipelines

## Why Testing Matters

### The Testing Pyramid

```GUI (Manual Testing)
       /\
      /  \     E2E Tests (Few, Slow, Expensive)
     /----\
    /      \   Integration Tests (Some, Medium Speed)
   /--------\
  /          \ Unit Tests (Many, Fast, Cheap)
 /------------\
```

**Unit Tests** (70%):

- Test individual functions/methods in isolation
- Fast execution (milliseconds)
- Easy to debug
- Should be the majority of your tests

**Integration Tests** (20%):

- Test multiple components working together
- Medium speed (seconds)
- Verify service interactions
- Test without external dependencies (use in-memory databases)

**End-to-End Tests** (10%):

- Test the entire system as a user would
- Slow execution (seconds to minutes)
- Validate real-world scenarios
- Most expensive to maintain

### Benefits of Comprehensive Testing

1. **Confidence in Changes**: Refactor without fear of breaking things
2. **Documentation**: Tests show how code should be used
3. **Regression Prevention**: Catch bugs before they reach production
4. **Faster Development**: Find issues early when they're cheap to fix
5. **CI/CD Enablement**: Automated testing enables continuous deployment

## Project Structure

```
practical6-example/
├── user-service/
│   ├── grpc/
│   │   ├── server.go           # gRPC service implementation
│   │   └── server_test.go      # ← Unit tests
│   ├── database/
│   ├── models/
│   └── main.go
├── menu-service/
│   ├── grpc/
│   │   ├── server.go
│   │   └── server_test.go      # ← Unit tests
│   ├── database/
│   ├── models/
│   └── main.go
├── order-service/
│   ├── grpc/
│   │   ├── server.go
│   │   └── server_test.go      # ← Unit tests (with mocks)
│   ├── database/
│   ├── models/
│   └── main.go
├── api-gateway/
│   ├── handlers/
│   ├── grpc/
│   └── main.go
├── tests/
│   ├── integration/
│   │   └── integration_test.go  # ← Integration tests
│   └── e2e/
│       └── e2e_test.go          # ← End-to-end tests
├── Makefile                      # ← Test automation
└── docker-compose.yml
```

## Testing Framework Used

Testify for Golang

`https://github.com/stretchr/testify`

## Phase 1: Unit Testing

Unit tests verify individual functions work correctly in isolation.

### 1.1: User Service Unit Tests

**File**: `user-service/grpc/server_test.go`

#### Key Testing Concepts

**Test Setup and Teardown**:

```go
func setupTestDB(t *testing.T) *gorm.DB {
    // Create in-memory SQLite database
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    require.NoError(t, err)

    // Auto-migrate models
    err = db.AutoMigrate(&models.User{})
    require.NoError(t, err)

    return db
}

func teardownTestDB(t *testing.T, db *gorm.DB) {
    sqlDB, err := db.DB()
    require.NoError(t, err)
    sqlDB.Close()
}
```

**Why SQLite?**

- In-memory database is fast (no disk I/O)
- No external dependencies needed
- Same SQL interface as PostgreSQL for basic operations
- Perfect for unit tests

**Test Structure - Table-Driven Tests**:

```go
func TestCreateUser(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer teardownTestDB(t, db)
    database.DB = db

    server := NewUserServer()

    // Define test cases
    tests := []struct {
        name        string
        request     *userv1.CreateUserRequest
        wantErr     bool
        expectedMsg string
    }{
        {
            name: "successful user creation",
            request: &userv1.CreateUserRequest{
                Name:        "John Doe",
                Email:       "john@example.com",
                IsCafeOwner: false,
            },
            wantErr: false,
        },
        {
            name: "create cafe owner",
            request: &userv1.CreateUserRequest{
                Name:        "Jane Owner",
                Email:       "jane@cafeshop.com",
                IsCafeOwner: true,
            },
            wantErr: false,
        },
    }

    // Execute tests
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            resp, err := server.CreateUser(ctx, tt.request)

            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.NotZero(t, resp.User.Id)
                assert.Equal(t, tt.request.Name, resp.User.Name)
            }
        })
    }
}
```

**Benefits of Table-Driven Tests**:

- Easy to add new test cases
- DRY (Don't Repeat Yourself)
- Clear test documentation
- Each test runs independently

#### Testing gRPC Error Codes

```go
func TestGetUser(t *testing.T) {
    // ... setup ...

    tests := []struct {
        name        string
        userID      uint32
        wantErr     bool
        expectedErr codes.Code  // ← gRPC error code
    }{
        {
            name:    "get existing user",
            userID:  1,
            wantErr: false,
        },
        {
            name:        "get non-existent user",
            userID:      9999,
            wantErr:     true,
            expectedErr: codes.NotFound,  // ← Expect 404
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := server.GetUser(ctx, &userv1.GetUserRequest{Id: tt.userID})

            if tt.wantErr {
                require.Error(t, err)
                st, ok := status.FromError(err)
                require.True(t, ok)
                assert.Equal(t, tt.expectedErr, st.Code())  // ← Verify error code
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.userID, resp.User.Id)
            }
        })
    }
}
```

**Common gRPC Error Codes**:

- `codes.NotFound`: Resource doesn't exist (404)
- `codes.InvalidArgument`: Bad request (400)
- `codes.Internal`: Server error (500)
- `codes.Unavailable`: Service down (503)

### 1.2: Menu Service Unit Tests

**File**: `menu-service/grpc/server_test.go`

Similar structure to user service, but tests menu-specific logic:

```go
func TestCreateMenuItem(t *testing.T) {
    // ... setup ...

    tests := []struct {
        name    string
        request *menuv1.CreateMenuItemRequest
        wantErr bool
    }{
        {
            name: "successful menu item creation",
            request: &menuv1.CreateMenuItemRequest{
                Name:        "Cappuccino",
                Description: "Espresso with steamed milk",
                Price:       4.50,
            },
            wantErr: false,
        },
        {
            name: "create item with zero price",
            request: &menuv1.CreateMenuItemRequest{
                Name:  "Water",
                Price: 0.0,
            },
            wantErr: false,  // Business decision: free items allowed
        },
    }

    // ... test execution ...
}
```

#### Testing Floating Point Numbers

```go
func TestPriceHandling(t *testing.T) {
    testCases := []struct {
        name  string
        price float64
    }{
        {"integer price", 5.0},
        {"two decimal places", 5.99},
        {"very small price", 0.01},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            resp, err := server.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
                Name:  "Test",
                Price: tc.price,
            })

            require.NoError(t, err)
            // Use InDelta for float comparison (allows small difference)
            assert.InDelta(t, tc.price, resp.MenuItem.Price, 0.001)
        })
    }
}
```

**Why `InDelta`?**

- Floating point arithmetic isn't exact
- `assert.Equal(3.14, 3.14)` might fail due to precision
- `assert.InDelta(3.14, 3.14, 0.001)` allows tiny differences

### 1.3: Order Service Unit Tests with Mocks

**File**: `order-service/grpc/server_test.go`

Order service is more complex because it depends on user and menu services. We use **mocks** to simulate these dependencies.

#### Creating Mocks

**Important**: Make sure to import `google.golang.org/grpc` in your test file for the `grpc.CallOption` type.

```go
import (
    // ... other imports ...
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// MockUserServiceClient simulates the user service
type MockUserServiceClient struct {
    mock.Mock  // Embeds testify's Mock functionality
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, req *userv1.GetUserRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*userv1.GetUserResponse), args.Error(1)
}

// Similar mock for MenuServiceClient...
```

#### Using Mocks in Tests

```go
func TestCreateOrder_Success(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer teardownTestDB(t, db)
    database.DB = db

    // Create mocks
    mockUserClient := new(MockUserServiceClient)
    mockMenuClient := new(MockMenuServiceClient)

    // Inject mocks into server
    server := &OrderServer{
        UserClient: mockUserClient,
        MenuClient: mockMenuClient,
    }

    // Define mock behavior
    mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 1}).
        Return(&userv1.GetUserResponse{
            User: &userv1.User{Id: 1, Name: "Test User"},
        }, nil)

    mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 1}).
        Return(&menuv1.GetMenuItemResponse{
            MenuItem: &menuv1.MenuItem{Id: 1, Name: "Coffee", Price: 2.50},
        }, nil)

    // Test
    ctx := context.Background()
    resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
        UserId: 1,
        Items:  []*orderv1.OrderItemRequest{
            {MenuItemId: 1, Quantity: 2},
        },
    })

    // Assert
    require.NoError(t, err)
    assert.Equal(t, uint32(1), resp.Order.UserId)
    assert.Len(t, resp.Order.OrderItems, 1)
    assert.InDelta(t, 2.50, resp.Order.OrderItems[0].Price, 0.001)

    // Verify mocks were called
    mockUserClient.AssertExpectations(t)
    mockMenuClient.AssertExpectations(t)
}
```

**Why Mocks?**

- **Isolation**: Test order service without starting user/menu services
- **Speed**: No network calls, instant responses
- **Control**: Simulate any scenario (success, failure, edge cases)
- **Deterministic**: Same results every time

#### Testing Error Scenarios

```go
func TestCreateOrder_InvalidUser(t *testing.T) {
    // ... setup ...

    // Mock user service returning error
    mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 999}).
        Return(nil, status.Errorf(codes.NotFound, "user not found"))

    // Test
    resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
        UserId: 999,
        Items:  []*orderv1.OrderItemRequest{{MenuItemId: 1, Quantity: 1}},
    })

    // Assert error
    require.Error(t, err)
    assert.Nil(t, resp)
    st, ok := status.FromError(err)
    require.True(t, ok)
    assert.Equal(t, codes.InvalidArgument, st.Code())
    assert.Contains(t, st.Message(), "user not found")
}
```

### Running Unit Tests

```bash
# Run all unit tests
make test-unit

# Run specific service tests
make test-unit-user
make test-unit-menu
make test-unit-order

# Run with coverage
make test-coverage
```

Expected output:

```
=== User Service Unit Tests ===
=== RUN   TestCreateUser
=== RUN   TestCreateUser/successful_user_creation
=== RUN   TestCreateUser/create_cafe_owner
--- PASS: TestCreateUser (0.01s)
    --- PASS: TestCreateUser/successful_user_creation (0.00s)
    --- PASS: TestCreateUser/create_cafe_owner (0.00s)
=== RUN   TestGetUser
--- PASS: TestGetUser (0.01s)
PASS
ok      user-service/grpc       0.123s
```

## Phase 2: Integration Testing

Integration tests verify that multiple services work together correctly.

### 2.1: Integration Test Architecture

**File**: `tests/integration/integration_test.go`

Integration tests use **bufconn** (in-memory gRPC connections) to test services together without network overhead.

#### Setting Up Test Services

```go
const bufSize = 1024 * 1024

var (
    userListener  *bufconn.Listener
    menuListener  *bufconn.Listener
    orderListener *bufconn.Listener
)

func setupUserService(t *testing.T) {
    // Setup in-memory database
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    require.NoError(t, err)

    err = db.AutoMigrate(&usermodels.User{})
    require.NoError(t, err)

    userdatabase.DB = db

    // Create gRPC server with bufconn (in-memory listener)
    userListener = bufconn.Listen(bufSize)
    s := grpc.NewServer()
    userv1.RegisterUserServiceServer(s, usergrpc.NewUserServer())

    // Start server in background
    go func() {
        if err := s.Serve(userListener); err != nil {
            log.Fatalf("Server exited: %v", err)
        }
    }()
}
```

**What is bufconn?**

- In-memory gRPC connection (no TCP/sockets)
- Faster than real network
- No port conflicts
- Perfect for integration tests

#### Connecting to Test Services

```go
func bufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
    return func(ctx context.Context, url string) (net.Conn, error) {
        return listener.Dial()
    }
}

// In test:
conn, err := grpc.DialContext(ctx, "bufnet",
    grpc.WithContextDialer(bufDialer(userListener)),
    grpc.WithTransportCredentials(insecure.NewCredentials()))
require.NoError(t, err)
defer conn.Close()

client := userv1.NewUserServiceClient(conn)
```

### 2.2: Complete Order Flow Integration Test

This test validates the entire order creation flow across all three services:

```go
func TestIntegration_CompleteOrderFlow(t *testing.T) {
    // Setup all three services
    setupUserService(t)
    defer userListener.Close()

    setupMenuService(t)
    defer menuListener.Close()

    ctx := context.Background()

    // Connect to user service
    userConn, err := grpc.DialContext(ctx, "bufnet",
        grpc.WithContextDialer(bufDialer(userListener)),
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    require.NoError(t, err)
    defer userConn.Close()

    userClient := userv1.NewUserServiceClient(userConn)

    // Connect to menu service
    menuConn, err := grpc.DialContext(ctx, "bufnet",
        grpc.WithContextDialer(bufDialer(menuListener)),
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    require.NoError(t, err)
    defer menuConn.Close()

    menuClient := menuv1.NewMenuServiceClient(menuConn)

    // Setup order service with connections to other services
    setupOrderService(t, userConn, menuConn)
    defer orderListener.Close()

    orderConn, err := grpc.DialContext(ctx, "bufnet",
        grpc.WithContextDialer(bufDialer(orderListener)),
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    require.NoError(t, err)
    defer orderConn.Close()

    orderClient := orderv1.NewOrderServiceClient(orderConn)

    // Step 1: Create a user
    userResp, err := userClient.CreateUser(ctx, &userv1.CreateUserRequest{
        Name:        "Integration User",
        Email:       "integration@test.com",
        IsCafeOwner: false,
    })
    require.NoError(t, err)
    userID := userResp.User.Id

    // Step 2: Create menu items
    item1, err := menuClient.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
        Name:        "Coffee",
        Description: "Hot coffee",
        Price:       2.50,
    })
    require.NoError(t, err)

    item2, err := menuClient.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
        Name:        "Sandwich",
        Description: "Ham sandwich",
        Price:       5.00,
    })
    require.NoError(t, err)

    // Step 3: Create an order
    orderResp, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
        UserId: userID,
        Items: []*orderv1.OrderItemRequest{
            {MenuItemId: item1.MenuItem.Id, Quantity: 2},
            {MenuItemId: item2.MenuItem.Id, Quantity: 1},
        },
    })

    require.NoError(t, err)
    assert.NotZero(t, orderResp.Order.Id)
    assert.Equal(t, userID, orderResp.Order.UserId)
    assert.Equal(t, "pending", orderResp.Order.Status)
    assert.Len(t, orderResp.Order.OrderItems, 2)

    // Verify prices were snapshotted
    assert.InDelta(t, 2.50, orderResp.Order.OrderItems[0].Price, 0.001)
    assert.InDelta(t, 5.00, orderResp.Order.OrderItems[1].Price, 0.001)

    // Step 4: Retrieve the order
    getOrderResp, err := orderClient.GetOrder(ctx, &orderv1.GetOrderRequest{
        Id: orderResp.Order.Id,
    })

    require.NoError(t, err)
    assert.Equal(t, orderResp.Order.Id, getOrderResp.Order.Id)
    assert.Len(t, getOrderResp.Order.OrderItems, 2)
}
```

**What This Tests**:

1. User creation via user service
2. Menu item creation via menu service
3. Order creation with validation (user exists, menu items exist)
4. Price snapshotting (order stores current price)
5. Order retrieval with items

### 2.3: Validation Integration Test

Test that order service properly validates data from other services:

```go
func TestIntegration_OrderValidation(t *testing.T) {
    // Setup services...

    orderClient := orderv1.NewOrderServiceClient(orderConn)

    // Try to create order with invalid user
    t.Run("invalid user", func(t *testing.T) {
        _, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
            UserId: 9999,  // ← Non-existent user
            Items: []*orderv1.OrderItemRequest{
                {MenuItemId: 1, Quantity: 1},
            },
        })

        require.Error(t, err)
        assert.Contains(t, err.Error(), "user not found")
    })

    // Create valid user
    userResp, err := userClient.CreateUser(ctx, &userv1.CreateUserRequest{
        Name: "Valid User", Email: "valid@test.com",
    })
    require.NoError(t, err)

    // Try to create order with invalid menu item
    t.Run("invalid menu item", func(t *testing.T) {
        _, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
            UserId: userResp.User.Id,
            Items: []*orderv1.OrderItemRequest{
                {MenuItemId: 9999, Quantity: 1},  // ← Non-existent item
            },
        })

        require.Error(t, err)
        assert.Contains(t, err.Error(), "menu item 9999 not found")
    })
}
```

### 2.4: Concurrent Order Test

Test that the system handles concurrent requests correctly:

```go
func TestIntegration_ConcurrentOrders(t *testing.T) {
    // Setup and create test data...

    // Create multiple orders concurrently
    numOrders := 10
    errChan := make(chan error, numOrders)
    respChan := make(chan *orderv1.CreateOrderResponse, numOrders)

    for i := 0; i < numOrders; i++ {
        go func() {
            resp, err := orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
                UserId: userID,
                Items: []*orderv1.OrderItemRequest{
                    {MenuItemId: itemID, Quantity: 1},
                },
            })
            errChan <- err
            respChan <- resp
        }()
    }

    // Collect results
    for i := 0; i < numOrders; i++ {
        err := <-errChan
        resp := <-respChan
        require.NoError(t, err)
        assert.NotZero(t, resp.Order.Id)
    }

    // Verify all orders were created
    ordersResp, err := orderClient.GetOrders(ctx, &orderv1.GetOrdersRequest{})
    require.NoError(t, err)
    assert.Len(t, ordersResp.Orders, numOrders)
}
```

**What This Tests**:

- Thread safety / concurrent access
- Database transaction handling
- No race conditions
- Order IDs are unique

### Running Integration Tests

```bash
make test-integration
```

Expected output:

```
=== RUN   TestIntegration_CreateUser
--- PASS: TestIntegration_CreateUser (0.02s)
=== RUN   TestIntegration_CompleteOrderFlow
--- PASS: TestIntegration_CompleteOrderFlow (0.05s)
=== RUN   TestIntegration_ConcurrentOrders
--- PASS: TestIntegration_ConcurrentOrders (0.10s)
PASS
ok      integration-tests       0.201s
```

## Phase 3: End-to-End Testing

E2E tests validate the entire system from the perspective of an external client using HTTP requests.

### 3.1: E2E Test Architecture

**File**: `tests/e2e/e2e_test.go`

E2E tests require the full system to be running with Docker Compose.

#### Test Configuration

```go
var apiGatewayURL = getEnv("API_GATEWAY_URL", "http://localhost:8080")

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

#### Helper Function for HTTP Requests

```go
func makeRequest(method, path string, body interface{}) (*http.Response, error) {
    var bodyReader io.Reader
    if body != nil {
        jsonData, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        bodyReader = bytes.NewBuffer(jsonData)
    }

    req, err := http.NewRequest(method, apiGatewayURL+path, bodyReader)
    if err != nil {
        return nil, err
    }

    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    client := &http.Client{Timeout: 10 * time.Second}
    return client.Do(req)
}
```

#### Waiting for Services

```go
func TestMain(m *testing.M) {
    // Wait for services to be ready
    fmt.Println("Waiting for services...")
    maxRetries := 30
    for i := 0; i < maxRetries; i++ {
        resp, err := http.Get(apiGatewayURL + "/api/users")
        if err == nil && resp.StatusCode < 500 {
            resp.Body.Close()
            fmt.Println("Services ready!")
            break
        }
        if i == maxRetries-1 {
            fmt.Println("Services not ready")
            os.Exit(1)
        }
        time.Sleep(2 * time.Second)
    }

    code := m.Run()
    os.Exit(code)
}
```

### 3.2: Complete Order Flow E2E Test

```go
func TestE2E_CompleteOrderFlow(t *testing.T) {
    // Step 1: Create a user
    userReq := map[string]interface{}{
        "name":          "E2E User",
        "email":         fmt.Sprintf("e2e-%d@test.com", time.Now().Unix()),
        "is_cafe_owner": false,
    }

    userResp, err := makeRequest("POST", "/api/users", userReq)
    require.NoError(t, err)
    defer userResp.Body.Close()

    assert.Equal(t, http.StatusCreated, userResp.StatusCode)

    var user User
    err = json.NewDecoder(userResp.Body).Decode(&user)
    require.NoError(t, err)

    // Step 2: Create menu items
    item1Req := map[string]interface{}{
        "name":        "Coffee",
        "description": "Hot coffee",
        "price":       2.50,
    }

    item1Resp, err := makeRequest("POST", "/api/menu", item1Req)
    require.NoError(t, err)
    defer item1Resp.Body.Close()

    var item1 MenuItem
    err = json.NewDecoder(item1Resp.Body).Decode(&item1)
    require.NoError(t, err)

    // (Create item2 similarly...)

    // Step 3: Create order
    orderReq := map[string]interface{}{
        "user_id": user.ID,
        "items": []map[string]interface{}{
            {"menu_item_id": item1.ID, "quantity": 2},
            {"menu_item_id": item2.ID, "quantity": 1},
        },
    }

    orderResp, err := makeRequest("POST", "/api/orders", orderReq)
    require.NoError(t, err)
    defer orderResp.Body.Close()

    assert.Equal(t, http.StatusCreated, orderResp.StatusCode)

    var order Order
    err = json.NewDecoder(orderResp.Body).Decode(&order)
    require.NoError(t, err)

    assert.NotZero(t, order.ID)
    assert.Equal(t, user.ID, order.UserID)
    assert.Len(t, order.OrderItems, 2)

    // Step 4: Retrieve order
    getOrderResp, err := makeRequest("GET", fmt.Sprintf("/api/orders/%d", order.ID), nil)
    require.NoError(t, err)
    defer getOrderResp.Body.Close()

    assert.Equal(t, http.StatusOK, getOrderResp.StatusCode)
}
```

**What This Tests**:

- HTTP API endpoints work correctly
- API Gateway routes requests properly
- Protocol translation (HTTP → gRPC → HTTP) works
- Full system integration
- Real database interactions
- Docker networking

### 3.3: Error Handling E2E Test

```go
func TestE2E_OrderValidation(t *testing.T) {
    t.Run("invalid user", func(t *testing.T) {
        orderReq := map[string]interface{}{
            "user_id": 999999,
            "items": []map[string]interface{}{
                {"menu_item_id": 1, "quantity": 1},
            },
        }

        resp, err := makeRequest("POST", "/api/orders", orderReq)
        require.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
    })
}
```

### 3.4: Running E2E Tests

```bash
# Start services first
make docker-up

# Wait a few seconds for services to initialize
sleep 10

# Run E2E tests
make test-e2e

# Or do everything in one command
make test-e2e-docker  # Starts services, runs tests, stops services
```

Expected output:

```
=== RUN   TestE2E_CompleteOrderFlow
--- PASS: TestE2E_CompleteOrderFlow (0.15s)
=== RUN   TestE2E_OrderValidation
=== RUN   TestE2E_OrderValidation/invalid_user
=== RUN   TestE2E_OrderValidation/invalid_menu_item
--- PASS: TestE2E_OrderValidation (0.05s)
PASS
ok      e2e-tests       0.523s
```

## Phase 4: Test Automation with Makefile

The Makefile provides convenient commands for running tests.

### Available Make Commands

```bash
make help                 # Show all available commands

# Testing
make test-unit           # Run all unit tests
make test-unit-user      # Run only user service tests
make test-unit-menu      # Run only menu service tests
make test-unit-order     # Run only order service tests
make test-integration    # Run integration tests
make test-e2e            # Run E2E tests (services must be running)
make test-e2e-docker     # Start services, run E2E tests, stop services
make test                # Run unit and integration tests
make test-all            # Run all tests including E2E
make test-coverage       # Generate coverage reports

# Docker
make docker-build        # Build Docker images
make docker-up           # Start all services
make docker-down         # Stop all services
make docker-logs         # Show logs from all services

# Development
make install-deps        # Install testing dependencies
make dev-setup           # Complete dev setup (deps + proto + docker)

# CI/CD
make ci-test             # Run tests suitable for CI
make ci-full             # Full CI pipeline
```

### Test Coverage

```bash
make test-coverage
```

This generates HTML coverage reports:

- `user-service/coverage.html`
- `menu-service/coverage.html`
- `order-service/coverage.html`

Open in browser to see line-by-line coverage:

```bash
open user-service/coverage.html
```

**Coverage Goals**:

- Unit tests: Aim for 80%+ coverage
- Critical paths: 100% coverage
- Error handling: Test all error cases

## Testing Best Practices

### 1. Test Independence

**Good**:

```go
func TestCreateUser(t *testing.T) {
    db := setupTestDB(t)  // ← Fresh database
    defer teardownTestDB(t, db)
    // Test code...
}
```

**Bad**:

```go
var sharedDB *gorm.DB  // ← Shared state

func TestCreateUser(t *testing.T) {
    // Uses sharedDB - tests affect each other
}
```

**Why**: Tests should run in any order and not affect each other.

### 2. Use Table-Driven Tests

**Good**:

```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {"valid input", "test@example.com", false},
    {"invalid email", "not-an-email", true},
    {"empty email", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test using tt.input
    })
}
```

**Why**: Easy to add cases, DRY, clear documentation.

### 3. Test Error Cases

**Don't just test happy paths**:

```go
func TestCreateOrder(t *testing.T) {
    // Test success
    t.Run("successful order", func(t *testing.T) { /* ... */ })

    // Test failure scenarios
    t.Run("invalid user", func(t *testing.T) { /* ... */ })
    t.Run("invalid menu item", func(t *testing.T) { /* ... */ })
    t.Run("empty order", func(t *testing.T) { /* ... */ })
}
```

### 4. Use Descriptive Test Names

**Good**:

```go
func TestCreateUser_DuplicateEmail_ReturnsConflictError(t *testing.T)
```

**Bad**:

```go
func TestUser(t *testing.T)
```

### 5. Assert Clearly

**Good**:

```go
require.NoError(t, err, "Failed to create user")
assert.Equal(t, "John", user.Name, "User name should match request")
```

**Bad**:

```go
if err != nil || user.Name != "John" {
    t.Fail()  // ← What failed?
}
```

### 6. Test One Thing Per Test

**Good**:

```go
func TestGetUser_ExistingUser_ReturnsUser(t *testing.T) { /* ... */ }
func TestGetUser_NonExistentUser_ReturnsNotFound(t *testing.T) { /* ... */ }
```

**Bad**:

```go
func TestGetUser(t *testing.T) {
    // Tests 5 different scenarios in one function
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.23"

      - name: Install dependencies
        run: make install-deps

      - name: Generate proto code
        run: make proto-generate

      - name: Run unit tests
        run: make test-unit

      - name: Run integration tests
        run: make test-integration

      - name: Build Docker images
        run: make docker-build

      - name: Run E2E tests
        run: make test-e2e-docker

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./user-service/coverage.out,./menu-service/coverage.out,./order-service/coverage.out
```

## Troubleshooting

### Issue 1: Tests Fail with "module not found"

**Solution**:

```bash
make install-deps
```

### Issue 2: E2E Tests Timeout

**Cause**: Services not ready

**Solution**:

```bash
# Check services are running
docker compose ps

# Check logs
docker compose logs

# Wait longer before running tests
sleep 15
make test-e2e
```

### Issue 3: Port Already in Use

**Solution**:

```bash
# Stop existing services
make docker-down

# Kill processes on port 8080
lsof -ti:8080 | xargs kill -9

# Restart
make docker-up
```

### Issue 4: Coverage Report Empty

**Solution**:

```bash
# Ensure test files are in same package as code
# user-service/grpc/server.go
# user-service/grpc/server_test.go  ← Same directory

# Run with -v flag to see which tests run
cd user-service && go test -v -coverprofile=coverage.out ./grpc/...
```

## Key Learnings

### 1. Testing Pyramid

- **70% Unit Tests**: Fast, isolated, test individual functions
- **20% Integration Tests**: Test services working together
- **10% E2E Tests**: Validate entire system

### 2. Test Types Comparison

| Aspect     | Unit            | Integration       | E2E           |
| ---------- | --------------- | ----------------- | ------------- |
| Speed      | Milliseconds    | Seconds           | Minutes       |
| Scope      | Single function | Multiple services | Entire system |
| Isolation  | Complete        | Partial           | None          |
| Debugging  | Easy            | Medium            | Hard          |
| Cost       | Low             | Medium            | High          |
| Confidence | Low             | Medium            | High          |

### 3. Mocking Strategy

- **Unit tests**: Mock all external dependencies
- **Integration tests**: Use real service implementations, in-memory databases
- **E2E tests**: No mocks, full production-like setup

### 4. Database Testing

- **Unit/Integration**: SQLite in-memory (fast, no cleanup needed)
- **E2E**: PostgreSQL in Docker (production-like)

### 5. When to Write Each Test Type

**Unit Test** when:

- Testing business logic
- Testing error handling
- Testing edge cases
- Fast feedback needed

**Integration Test** when:

- Testing service interactions
- Testing protocol translation
- Testing data flow
- Verifying contracts

**E2E Test** when:

- Testing critical user journeys
- Validating deployment
- Testing API contracts
- Pre-production validation

## Real-World Applications

### 1. Test-Driven Development (TDD)

Write tests first, then implement:

```go
// 1. Write failing test
func TestCalculateOrderTotal(t *testing.T) {
    order := Order{Items: []Item{{Price: 10, Quantity: 2}}}
    total := CalculateTotal(order)
    assert.Equal(t, 20.0, total)
}

// 2. Run test - it fails (function doesn't exist)

// 3. Implement minimum code to pass
func CalculateTotal(order Order) float64 {
    var total float64
    for _, item := range order.Items {
        total += item.Price * float64(item.Quantity)
    }
    return total
}

// 4. Run test - it passes

// 5. Refactor if needed
```

### 2. Regression Testing

When fixing a bug:

```go
// 1. Add test that reproduces the bug
func TestCreateOrder_NegativeQuantity_ReturnsError(t *testing.T) {
    // This test initially fails, showing the bug exists
    resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
        UserId: 1,
        Items:  []*orderv1.OrderItemRequest{
            {MenuItemId: 1, Quantity: -5},  // ← Bug: negative allowed
        },
    })

    require.Error(t, err)
    assert.Contains(t, err.Error(), "quantity must be positive")
}

// 2. Fix the bug in the code

// 3. Test passes - bug is fixed and won't come back
```

### 3. Continuous Integration

```bash
# In CI pipeline
make ci-full

# If tests pass → deploy to staging
# If tests fail → block deployment
```

## Next Steps

### For This Practical

1. **Run all tests**:

   ```bash
   make dev-setup      # One-time setup
   make test-all       # Run everything
   ```

2. **Generate coverage report**:

   ```bash
   make test-coverage
   open */coverage.html
   ```

3. **Experiment with failures**:

   - Break a test intentionally
   - See how it fails
   - Fix it

4. **Add your own test**:
   - Test updating a user
   - Test deleting a menu item
   - Test order status changes

### For Production Systems

1. **Add more test types**:

   - Performance tests (load testing with k6)
   - Security tests (SQL injection, XSS)
   - Chaos engineering (kill random services)

2. **Improve coverage**:

   - Aim for 80%+ code coverage
   - Focus on critical paths first
   - Test all error conditions

3. **Automate everything**:

   - Run tests on every commit
   - Block PRs if tests fail
   - Deploy automatically if tests pass

4. **Monitor test health**:
   - Track test execution time
   - Fix flaky tests immediately
   - Remove obsolete tests

## Submission Requirements

- Submit your updated repository with all tests implemented
- Currently the tests are able to complete successfully with the unit tests
- Integration and E2E tests should also pass if the services are running correctly. 
- Both are not implemented fully yet.
However, both are failing due to issues with go.sum imports which has been defined in the Dockerfile for each service.

Task
- Ensure that the setup of integration and e2e are working and pass successfully

Files to Submit
- All test files under `tests/integration/` and `tests/e2e/`
- Any updates to service code required to make tests pass
- Screenshot of test results showing all tests passing in the terminal


## Conclusion

You now understand:

1. **Unit Testing**: Testing individual functions with mocks
2. **Integration Testing**: Testing services working together
3. **E2E Testing**: Testing the entire system as users would
4. **Test Automation**: Using Make for consistent test execution
5. **Best Practices**: Independence, table-driven tests, clear assertions
6. **CI/CD**: Integrating tests into deployment pipelines

Testing is not optional—it's essential for:

- **Confidence**: Deploy without fear
- **Quality**: Catch bugs early
- **Documentation**: Tests show how code should work
- **Speed**: Automated tests are faster than manual testing

## Additional Resources

- [Go Testing Best Practices](https://golang.org/pkg/testing/)
- [Testify Library](https://github.com/stretchr/testify)
- [gRPC Testing Guide](https://grpc.io/docs/languages/go/basics/#testing)
- [The Practical Test Pyramid](https://martinfowler.com/articles/practical-test-pyramid.html)
- [Test-Driven Development](https://martinfowler.com/bliki/TestDrivenDevelopment.html)
